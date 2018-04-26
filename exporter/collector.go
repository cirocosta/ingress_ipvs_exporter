package exporter

import (
	"os"

	"github.com/docker/libnetwork/ipvs"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Collector struct {
	logger        zerolog.Logger
	ipvs          *ipvs.Handle
	namespacePath string
}

type CollectorConfig struct {
	NamespacePath string
}

type Statistic struct {
	Port uint16
	ipvs.SvcStats
}

func NewCollector(cfg CollectorConfig) (c Collector, err error) {
	c.ipvs, err = ipvs.New(cfg.NamespacePath)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to create ipvs handle for namespace path '%s'",
			cfg.NamespacePath)
		return
	}

	if c.ipvs == nil {
		err = errors.Wrapf(err,
			"created nil ipvs handle for namespace path '%s'",
			cfg.NamespacePath)
		return
	}

	c.namespacePath = cfg.NamespacePath
	c.logger = zerolog.New(os.Stdout).
		With().
		Str("from", "collector").
		Logger()

	return
}

func (c *Collector) GetStatistics() (stats []Statistic, err error) {
	var services []*ipvs.Service

	services, err = c.ipvs.GetServices()
	if err != nil {
		err = errors.Wrapf(err, "failed to retrieve ipvs svcs from ns %s",
			c.namespacePath)
		return
	}

	if len(services) == 0 {
		return
	}

	stats = make([]Statistic, len(services))
	for ndx, service := range services {
		stats[ndx] = Statistic{
			Port:     service.Port,
			SvcStats: service.Stats,
		}
	}

	return
}
