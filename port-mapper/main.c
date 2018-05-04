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

typedef struct mark_mapping {
	__u16 destination_port;
	__u32 firewall_mark;
} m_mark_mapping_t;

int
m_count_rules()
{
	int err = 0;

	return err;
}

int
m_get_mark_mappings()
{
	return 0;
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

int
m_init()
{
	return 0;
}

int
m_chain_exists(struct xtc_handle* handle)
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

int
main(void)
{
	int                     err;
	struct xtc_handle*      handle;
	const char*             chain = NULL;
	const struct ipt_entry* rule;
	unsigned int            rule_count = 0;

	iptables_globals.program_name = "fwmark-mapper";
	err = xtables_init_all(&iptables_globals, NFPROTO_IPV4);
	if (err < 0) {
		fprintf(stderr,
		        "name=%s ver=%s - failed to initialize xtables\n",
		        iptables_globals.program_name,
		        iptables_globals.program_version);
		exit(1);
	}

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
	if (m_chain_exists(handle) == -1) {
		fprintf(stderr,
		        "expected chain " M_CHAIN
		        " to look for fwmark mappings doesn't exist\n");
		exit(1);
	}

	// count the number of rules
	rule = iptc_first_rule(chain, handle);
	while (rule) {
		rule_count += 1;
		rule = iptc_next_rule(rule, handle);
	}

	// nothing to do if there are no rules
	if (rule_count == 0) {
		exit(0);
	}

	// allocate a mark_mappings array that can
	// fit the expected number of rule mappings
	m_mark_mapping_t* mark_mappings =
	  malloc(rule_count * sizeof(mark_mappings));
	if (mark_mappings == NULL) {
		perror("malloc");
		fprintf(stderr,
		        "failed to allocate enough memory for "
		        "mark mappings\n");
		exit(1);
	}

	// populate the array with the mappings
	rule = iptc_first_rule(chain, handle);
	for (unsigned int __i = 0; __i < rule_count; __i++) {
		m_get_mark_mapping_from_rule(rule, &mark_mappings[__i]);
		rule = iptc_next_rule(rule, handle);
	}

	// show the mappings
	for (unsigned int __i = 0; __i < rule_count; __i++) {
		printf("mark=%d,port=%d\n",
		       mark_mappings[__i].firewall_mark,
		       mark_mappings[__i].destination_port);
	}

	exit(0);
}
