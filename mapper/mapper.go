package mapper

// #cgo LDFLAGS: -lip4tc -lxtables
// #include "./mapper.h"
import (
	"C"
)

import (
	"os"
	"runtime"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vishvananda/netns"
)

func init() {
	errno := C.m_init()
	if errno != 0 {
		log.Fatal().Msg("mapper's init failed")
		os.Exit(int(errno))
	}
}

type MapperConfig struct {
	NamespacePath string
}

type Mapper struct {
	nsHandle netns.NsHandle
	logger   zerolog.Logger
}

func NewMapper(cfg MapperConfig) (m Mapper, err error) {
	if cfg.NamespacePath == "" {
		err = errors.Errorf("NamespacePath must be provided")
		return
	}

	m.nsHandle, err = netns.GetFromPath(cfg.NamespacePath)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to retrieve ns from path %s",
			cfg.NamespacePath)
		return
	}

	m.logger = zerolog.New(os.Stdout).
		With().
		Str("from", "mapper").
		Logger()

	return
}

func (m Mapper) GetMappings() (res map[uint32]uint16, err error) {
	currentNs, err := netns.Get()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to retrieve current namespace")
		return
	}

	runtime.LockOSThread()
	defer func() {
		netns.Set(currentNs)
		if err != nil {
			err = errors.Wrapf(err,
				"failed to get back to original netns")
		}
		runtime.UnlockOSThread()
	}()

	err = netns.Set(m.nsHandle)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to set network namespace")
		return
	}

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
