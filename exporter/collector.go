package exporter

import (
	"os"
	"strconv"

	"github.com/docker/libnetwork/ipvs"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type Collector struct {
	logger        zerolog.Logger
	ipvs          *ipvs.Handle
	namespacePath string

	connectionsTotalDesc *prometheus.Desc
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

	c.connectionsTotalDesc = prometheus.NewDesc(
		"ipvs_connections_total",
		"The total number of connections made",
		[]string{"address"},
		prometheus.Labels{"namespace": cfg.NamespacePath},
	)

	return
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.connectionsTotalDesc
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	var (
		services []*ipvs.Service
		err      error
	)

	services, err = c.ipvs.GetServices()
	if err != nil {
		c.logger.Error().
			Err(err).
			Str("namespace", c.namespacePath).
			Msg("failed to retrieve ipvs services")
		return
	}

	if len(services) == 0 {
		return
	}

	for _, service := range services {
		c.logger.Debug().
			Interface("service", service).
			Msg("reporting service")

		ch <- prometheus.MustNewConstMetric(
			c.connectionsTotalDesc,
			prometheus.CounterValue,
			float64(service.Stats.Connections),
			strconv.Itoa(int(service.FWMark)),
		)
	}

	return

}
