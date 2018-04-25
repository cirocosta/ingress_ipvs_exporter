package exporter

import (
	"os"

	"github.com/docker/libnetwork/ipvs"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Collector struct {
	logger zerolog.Logger
	ipvs   *ipvs.Handle
}

type CollectorConfig struct {
	NamespacePath string
}

func NewCollector(cfg CollectorConfig) (c Collector, err error) {
	c.ipvs, err = ipvs.New(cfg.NamespacePath)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to create ipvs handle for namespace path '%s'",
			cfg.NamespacePath)
		return
	}

	c.logger = zerolog.New(os.Stdout).
		With().
		Str("from", "collector").
		Logger()

	return
}
