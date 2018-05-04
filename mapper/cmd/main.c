#include "../mapper.h"

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

	for (unsigned int i = 0; i < mappings->length; i++) {
		printf("mark=%d,port=%d\n",
		       mappings->data[i]->firewall_mark,
		       mappings->data[i]->destination_port);
	}

	m_destroy_mark_mappings(mappings);

	return 0;
}
