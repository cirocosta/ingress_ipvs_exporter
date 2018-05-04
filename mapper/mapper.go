package mapper

// #cgo LDFLAGS: -lip4tc -lxtables
// #include "./mapper.h"
import (
	"C"
)

import (
	"unsafe"

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
	var i C.ushort = 1

	mappings := C.m_get_mark_mappings()
	if mappings == nil {
		return
	}

	defer C.m_destroy_mark_mappings(mappings)

	if mappings.length == 0 {
		return
	}

	res = make([]*Mapping, mappings.length)
	for ; i < mappings.length+1; i++ {
		ptr := unsafe.Pointer(
			uintptr(unsafe.Pointer(mappings.data)) +
				uintptr(struct_size*int(i)))
		casted := (*C.struct_mark_mapping)(ptr)

		res[i-1] = &Mapping{
			DestinationPort: uint16(casted.destination_port),
			FirewallMark:    uint16(casted.firewall_mark),
		}
	}

	return
}
