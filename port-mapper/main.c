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

struct xtables_globals iptables_globals = { 0 } ;

int
main(void)
{
	int                err;
	struct xtc_handle* handle;
	const char*        chain     = NULL;
	const char*        tablename = "mangle";

	iptables_globals.program_name = "p1";
	err = xtables_init_all(&iptables_globals, NFPROTO_IPV4);
	if (err < 0) {
		fprintf(stderr,
		        "name=%s ver=%s - failed to initialize xtables\n",
		        iptables_globals.program_name,
		        iptables_globals.program_version);
		exit(1);
	}

	handle = iptc_init(tablename);
	if (!handle) {
		fprintf(stderr,
		        "failed to initialize table handle: %s\n",
		        iptc_strerror(errno));
		exit(errno);
	}

        for (chain = iptc_first_chain(handle);
                        chain;
                        chain = iptc_next_chain(handle)) {
                printf(":%s ", chain);
        }

	exit(0);
}
