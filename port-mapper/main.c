#include <dlfcn.h>
#include <fcntl.h>
#include <getopt.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/errno.h>
#include <time.h>
#include <xtables.h>

#include "libiptc/libiptc.h"

struct xtables_globals iptables_globals = { 0 };
struct xt_mark_tginfo2 {
	__u32 mark, mask;
};

int
main(void)
{
	int                err;
	struct xtc_handle* handle;
	const char*        chain         = NULL;
	const char*        table         = "mangle";
	const char*        desired_chain = "PREROUTING";

	iptables_globals.program_name = "p1";
	err = xtables_init_all(&iptables_globals, NFPROTO_IPV4);
	if (err < 0) {
		fprintf(stderr,
		        "name=%s ver=%s - failed to initialize xtables\n",
		        iptables_globals.program_name,
		        iptables_globals.program_version);
		exit(1);
	}

	handle = iptc_init(table);
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
		if (strcmp(chain, desired_chain)) {
			continue;
		}

		printf("chain=%s\n", chain);

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
