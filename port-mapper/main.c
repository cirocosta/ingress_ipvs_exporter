#include <dlfcn.h>
#include <fcntl.h>
#include <getopt.h>
#include <libiptc/libiptc.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/errno.h>
#include <time.h>
#include <xtables.h>

// M_PRGRAM_NAME defines the program name set in
// the internal iptables configuration so that
// helper functions output such name when called.
#define M_PROGRAM_NAME "docker-fwmark-mapper"

// M_CHAIN defines the chain that we should
// look for {destination_port:mark} tuples.
//
// The chain is supposed to exist and already be
// populated within the current network namespace.
#define M_CHAIN "PREROUTING"

// M_CHAIN defines the table to filter the
// rule lookups.
//
// Given that docker swarm perform the packet marking
// at the mangle table, this is the one we use when
// looking for the mark definitions.
#define M_TABLE "mangle"

struct xtables_globals iptables_globals = { 0 };

struct xt_mark_tginfo2 {
	__u32 mark;
	__u32 mask;
};

/**
 * m_mark_mapping_t unites both destination_port and
 * firewall_mark as retrieved from iptables rules.
 */
typedef struct mark_mapping {
	__u16 destination_port;
	__u32 firewall_mark;
} m_mark_mapping_t;

/**
 * m_mark_mappings_t holds an array of mappings and
 * the array length to facilitate list operations.
 */
typedef struct mark_mappings {
	__u16              length;
	m_mark_mapping_t** data;
} m_mark_mappings_t;

/**
 * m_new_mark_mappings instantiates a mark_mappings
 * struct that holds an array of mark_mapping instances
 * with an additional `length` field to auxiliate in
 * iterations and destruction.
 */
m_mark_mappings_t*
m_new_mark_mappings(__u16 length)
{
	m_mark_mappings_t* m = malloc(sizeof *m);
	if (m == NULL) {
		goto ERR_ALOC;
	}

	m->data = malloc(length * sizeof(*m->data));
	if (m->data == NULL) {
		goto ERR_ALOC;
	}

	for (unsigned int i = 0; i < length; i++) {
		m->data[i] = malloc(sizeof *m->data[i]);
		if (m->data[i] == NULL) {
			goto ERR_ALOC;
		}
	}

	m->length = length;

	return m;

ERR_ALOC:
	perror("malloc");
	fprintf(stderr,
	        "failed to allocate enough memory for "
	        "mark mappings\n");
	exit(1);
}

/**
 * m_destroy_mark_mappings takes care of properly freeing
 * any allocated memory and settings the respective
 * pointers to NULL to make sure they're not used
 * later.
 */
void
m_destroy_mark_mappings(m_mark_mappings_t* m)
{
	if (m == NULL) {
		return;
	}

	for (int i = 0; i < m->length; i++) {
		if (m->data[i] == NULL) {
			continue;
		}

		free(m->data[i]);
		m->data[i] = NULL;
	}

	free(m->data);
	m->data = NULL;

	free(m);
}

int
m_get_mark_mapping_from_rule(const struct ipt_entry* rule,
                             m_mark_mapping_t*       mapping)
{
	const struct xt_entry_target* rule_target;
	const struct xt_tcp*          tcp_info;
	const struct xt_mark_tginfo2* mark_info;

	rule_target = ipt_get_target((struct ipt_entry*)rule);
	if (!rule->target_offset) {
		return 0;
	}

	struct xt_entry_match* match = { 0 };
	for (unsigned int __i = sizeof(struct ipt_entry);
	     __i < rule->target_offset;
	     __i += match->u.match_size) {
		match = (void*)rule + __i;

		tcp_info = (struct xt_tcp*)match->data;

		if (tcp_info->dpts[0] != 0 || tcp_info->dpts[1] != 0xFFFF) {
			mapping->destination_port = tcp_info->dpts[0];
			break;
		}
	}

	mark_info              = (const void*)rule_target->data;
	mapping->firewall_mark = mark_info->mark;
	return 0;
}

/**
 * m_init takes care of initializing the internal global
 * variables that the xtables lib depends on.
 *
 * This method is meant to be called only once (at startup,
 * before executing other methods).
 *
 *
 * TODO what else is performed?
 */
int
m_init()
{
	int err = 0;

	iptables_globals.program_name = M_PROGRAM_NAME;
	err = xtables_init_all(&iptables_globals, NFPROTO_IPV4);
	if (err < 0) {
		return err;
	}

	return err;
}

/**
 * _m_chain_exists is an internal method that verifies whether
 * the desired chain exists in the current table.
 */
int
_m_chain_exists(struct xtc_handle* handle)
{
	const char* chain = NULL;

	for (chain = iptc_first_chain(handle); chain;
	     chain = iptc_next_chain(handle)) {
		if (!strcmp(chain, M_CHAIN)) {
			return 0;
		}
	}

	return -1;
}

m_mark_mappings_t*
m_get_mark_mappings()
{
	struct xtc_handle*      handle;
	const struct ipt_entry* rule;
	unsigned int            rule_count = 0;
	m_mark_mappings_t*      mappings;

	// take a snapshot of the iptables rules at the
	// current point in time
	handle = iptc_init(M_TABLE);
	if (!handle) {
		fprintf(stderr,
		        "failed to initialize table handle: %s\n",
		        iptc_strerror(errno));
		exit(errno);
	}

	// check if chain exists
	if (_m_chain_exists(handle) == -1) {
		fprintf(stderr,
		        "expected chain " M_CHAIN
		        " to look for fwmark mappings doesn't exist\n");
		exit(1);
	}

	// count the number of rules
	rule = iptc_first_rule(M_CHAIN, handle);
	while (rule) {
		rule_count += 1;
		rule = iptc_next_rule(rule, handle);
	}

	// nothing to do if there are no rules
	if (rule_count == 0) {
		exit(0);
	}

	// create the mappings holder
	mappings = m_new_mark_mappings(rule_count);

	// populate the array with the mappings
	rule = iptc_first_rule(M_CHAIN, handle);
	for (unsigned int i = 0; i < rule_count; i++) {
		m_get_mark_mapping_from_rule(rule, mappings->data[i]);
		rule = iptc_next_rule(rule, handle);
	}

	iptc_free(handle);

	return mappings;
}

int
main(void)
{
	int err = 0;

	err = m_init();
	if (err != 0) {
		fprintf(stderr,
		        "failed to initialize global xtable variables\n");
		exit(1);
	}

	m_mark_mappings_t* mappings = m_get_mark_mappings();

	// show the mappings
	for (unsigned int i = 0; i < mappings->length; i++) {
		printf("mark=%d,port=%d\n",
		       mappings->data[i]->firewall_mark,
		       mappings->data[i]->destination_port);
	}

	m_destroy_mark_mappings(mappings);

	exit(0);
}
