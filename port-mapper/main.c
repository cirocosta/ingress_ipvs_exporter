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

// MAPPER_CHAIN defines the chain that we should
// look for {destination_port:mark} tuples.
//
// The chain is supposed to exist and already be
// populated within the current network namespace.
#define MAPPER_CHAIN "PREROUTING"

// MAPPER_CHAIN defines the table to filter the
// rule lookups.
//
// Given that docker swarm perform the packet marking
// at the mangle table, this is the one we use when
// looking for the mark definitions.
#define MAPPER_TABLE "mangle"

struct xtables_globals iptables_globals = { 0 };

struct xt_mark_tginfo2 {
	__u32 mark, mask;
};

int
main(void)
{
	int                err;
	struct xtc_handle* handle;
	const char*        chain = NULL;

	iptables_globals.program_name = "p1";
	err = xtables_init_all(&iptables_globals, NFPROTO_IPV4);
	if (err < 0) {
		fprintf(stderr,
		        "name=%s ver=%s - failed to initialize xtables\n",
		        iptables_globals.program_name,
		        iptables_globals.program_version);
		exit(1);
	}

	handle = iptc_init(MAPPER_TABLE);
	if (!handle) {
		fprintf(stderr,
		        "failed to initialize table handle: %s\n",
		        iptc_strerror(errno));
		exit(errno);
	}

	const struct ipt_entry*       rule;
	const struct xt_entry_target* rule_target;
	const struct xt_tcp*          tcp_info;

	for (chain = iptc_first_chain(handle); chain;
	     chain = iptc_next_chain(handle)) {
		if (strcmp(chain, MAPPER_CHAIN)) {
			continue;
		}

		rule        = iptc_first_rule(chain, handle);
		rule_target = ipt_get_target((struct ipt_entry*)rule);

		if (!rule->target_offset) {
			continue;
		}

		struct xt_entry_match* match = { 0 };
		for (unsigned int __i = sizeof(struct ipt_entry);
		     __i < rule->target_offset;
		     __i += match->u.match_size) {
			match = (void*)rule + __i;

			tcp_info = (struct xt_tcp*)match->data;

			if (tcp_info->dpts[0] != 0 ||
			    tcp_info->dpts[1] != 0xFFFF) {
				printf("--dport %u:%u\n",
				       tcp_info->dpts[0],
				       tcp_info->dpts[1]);
			}
		}

		const struct xt_mark_tginfo2* mark_info =
		  (const void*)rule_target->data;
		printf("mark: %d\n", mark_info->mark);
	}

	exit(0);
}
