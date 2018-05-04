package mapper

// #cgo LDFLAGS: -lip4tc -lxtables
// #include "./mapper.h"
import (
	"C"
)

import (
	"github.com/pkg/errors"
)

const struct_size = 32

type Mapping struct {
	DestinationPort uint16
	FirewallMark    uint16
}

func Init() (err error) {
	errno := C.m_init()
	if errno != 0 {
		err = errors.Errorf("mapper's init failed")
		return
	}

	return
}

func GetMappings() (res []*Mapping) {
	var i C.ushort = 0

	mappings := C.m_get_mark_mappings()
	if mappings == nil {
		return
	}

	defer C.m_destroy_mark_mappings(mappings)

	if mappings.length == 0 {
		return
	}

	res = make([]*Mapping, mappings.length)
	for ; i < mappings.length; i++ {
		mapping := C.m_get_mark_mapping_at(mappings, i)

		res[i] = &Mapping{
			DestinationPort: uint16(mapping.destination_port),
			FirewallMark:    uint16(mapping.firewall_mark),
		}
	}

	return
}
