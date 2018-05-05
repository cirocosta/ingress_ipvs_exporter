package mapper

// #cgo LDFLAGS: -lip4tc -lxtables
// #include "./mapper.h"
import (
	"C"
)

import (
	"os"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	errno := C.m_init()
	if errno != 0 {
		log.Fatal().Msg("mapper's init failed")
		os.Exit(int(errno))
	}
}

type Mapper struct {
	logger zerolog.Logger
}

func NewMapper() (m Mapper, err error) {
	m.logger = zerolog.New(os.Stdout).
		With().
		Str("from", "mapper").
		Logger()

	return
}

func (m Mapper) GetMappings() (res map[uint32]uint16, err error) {
	mappings := C.m_get_mark_mappings()
	if mappings == nil {
		return
	}

	defer C.m_destroy_mark_mappings(mappings)

	if mappings.length == 0 {
		return
	}

	res = make(map[uint32]uint16)

	var i C.ushort = 0
	for ; i < mappings.length; i++ {
		mapping := C.m_get_mark_mapping_at(mappings, i)
		if mapping == nil {
			err = errors.Errorf("couldn't retrieve fwmark "+
				"mapping at position %d",
				i)
			return
		}

		res[uint32(mapping.firewall_mark)] =
			uint16(mapping.destination_port)
	}

	return
}
