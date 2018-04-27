package exporter

import (
	"os"
	"strconv"

	"github.com/docker/libnetwork/ipvs"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

// Collector implements the Prometheus Collector interface
// to provide metrics regarding IPVS in a specified network
// namespace.
type Collector struct {
	logger        zerolog.Logger
	ipvs          *ipvs.Handle
	namespacePath string

	connectionsTotalDesc *prometheus.Desc
	servicesTotalDesc    *prometheus.Desc
}

// CollectorConfig provides the necessary configuration for
// initializing a Collector.
//
// This configuration is only meant to be used in NewCollector,
// which then validates the provided values.
type CollectorConfig struct {
	// NamespacePath corresponds to the full path to the
	// network namespace that the collector should use
	// when performing the netlink calls for IPVS statistics.
	//
	// Examples:
	// - "/var/run/netns/my-namespace"
	// - "/var/run/docker/netns/ingress_sbox"
	// - "" (nothing - use the current ns)
	NamespacePath string
}

// NewCollector initializes the collector making use of the configuration
// provided via CollectorConfig.
//
// This method has no side effects when it comes to prometheus - metrics
// descriptions are not registered in the global instance here (see
// NewExporter).
func NewCollector(cfg CollectorConfig) (c Collector, err error) {
	c.ipvs, err = ipvs.New(cfg.NamespacePath)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to create ipvs handle for namespace path '%s'",
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
		"The total number of connections made to a virtual server",
		[]string{"fwmark"},
		prometheus.Labels{"namespace": cfg.NamespacePath},
	)

	c.servicesTotalDesc = prometheus.NewDesc(
		"ipvs_services_total",
		"The total number of services registered in ipvs",
		nil,
		prometheus.Labels{"namespace": cfg.NamespacePath},
	)

	return
}

// Describe sends to the provided channel the set of all configured
// metric descriptions at the moment of collector registration.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.connectionsTotalDesc
	ch <- c.servicesTotalDesc
}

// Collect is called by Prometheus when collecting metrics.
// It's meant to list all of the services registered in IPVS in a
// given namespace and the corresponding metrics to the supplied
// channel.
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

	ch <- prometheus.MustNewConstMetric(
		c.servicesTotalDesc,
		prometheus.GaugeValue,
		float64(len(services)),
	)

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
