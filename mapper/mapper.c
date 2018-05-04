#include "./mapper.h"

struct xtables_globals iptables_globals = { 0 };

struct xt_mark_tginfo2 {
	__u32 mark;
	__u32 mask;
};

m_mark_mapping_t*
m_get_mark_mapping_at(m_mark_mappings_t* m, __u16 pos)
{
	if (pos > m->length) {
		return NULL;
	}

	return m->data[pos];
}

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
	        "failed to allocate enough memory for %d "
	        "mark mappings\n",
	        length);
	exit(1);
}

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
_m_get_mark_mapping_from_rule(const struct ipt_entry* rule,
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
	m_mark_mappings_t*      mappings   = NULL;

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
		return mappings;
	}

	// create the mappings holder
	mappings = m_new_mark_mappings(rule_count);

	// populate the array with the mappings
	rule = iptc_first_rule(M_CHAIN, handle);
	for (unsigned int i = 0; i < rule_count; i++) {
		_m_get_mark_mapping_from_rule(rule, mappings->data[i]);
		rule = iptc_next_rule(rule, handle);
	}

	iptc_free(handle);

	return mappings;
}
