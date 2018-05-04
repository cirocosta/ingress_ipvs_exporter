#ifndef _MAPPER_H
#define _MAPPER_H

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

// m_get_mark_mapping_at retrieves a mark_mapping from
// mark_mappings by accessing `data` and  then retrieving
// the field.
m_mark_mapping_t*
m_get_mark_mapping_at(m_mark_mappings_t* m, __u16 pos);

/**
 * m_new_mark_mappings instantiates a mark_mappings
 * struct that holds an array of mark_mapping instances
 * with an additional `length` field to auxiliate in
 * iterations and destruction.
 */
m_mark_mappings_t*
m_new_mark_mappings(__u16 length);

/**
 * m_destroy_mark_mappings takes care of properly freeing
 * any allocated memory and settings the respective
 * pointers to NULL to make sure they're not used
 * later.
 */
void
m_destroy_mark_mappings(m_mark_mappings_t* m);

/**
 * m_get_mark_mappings retrieves a m_mark_mappings_t instance
 * that contains all the fwmark mappings as seens by iptables
 * in the current namespace.
 *
 * note.: during the retrieval, allocations are performed. Don't
 * forget to free the `m_mark_mappings_t` structure after using
 * it.
 */
m_mark_mappings_t*
m_get_mark_mappings();

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
m_init();

#endif
