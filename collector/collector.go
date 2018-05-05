package collector

import (
	"os"
	"runtime"
	"strconv"

	"github.com/cirocosta/ipvs_exporter/mapper"
	"github.com/mqliang/libipvs"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/vishvananda/netns"
)

// Collector implements the Prometheus Collector interface
// to provide metrics regarding IPVS in a specified network
// namespace.
type Collector struct {
	logger   zerolog.Logger
	mapper   *mapper.Mapper
	ipvs     libipvs.IPVSHandle
	nsHandle *netns.NsHandle

	connectionsTotalDesc *prometheus.Desc
	bytesInTotalDesc     *prometheus.Desc
	bytesOutTotalDesc    *prometheus.Desc
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
	var nsHandle netns.NsHandle

	if cfg.NamespacePath != "" {
		nsHandle, err = netns.GetFromPath(cfg.NamespacePath)
		if err != nil {
			err = errors.Wrapf(err,
				"failed to retrieve ns from path %s",
				cfg.NamespacePath)
			return
		}

		c.nsHandle = &nsHandle
	}

	getIpvsHandle := func() (err error) {
		c.ipvs, err = libipvs.New()
		if err != nil {
			err = errors.Wrapf(err,
				"failed to create ipvs handle for namespace path")
			return
		}

		return
	}

	if c.nsHandle != nil {
		err = c.RunInNetns(getIpvsHandle)
	} else {
		err = getIpvsHandle()
	}
	if err != nil {
		err = errors.Wrapf(err,
			"failed to retrieve ipvs handle")
	}

	fwmarkMapper, err := mapper.NewMapper()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to create fwmarkmapper")
		return
	}

	c.mapper = &fwmarkMapper
	c.logger = zerolog.New(os.Stdout).
		With().
		Str("from", "collector").
		Logger()

	c.connectionsTotalDesc = prometheus.NewDesc(
		"ipvs_connections_total",
		"The total number of connections made to a virtual server",
		[]string{"fwmark", "port"},
		prometheus.Labels{"namespace": cfg.NamespacePath},
	)

	c.bytesInTotalDesc = prometheus.NewDesc(
		"ipvs_bytes_in_total",
		"The total number of incoming bytes a virtual server",
		[]string{"fwmark", "port"},
		prometheus.Labels{"namespace": cfg.NamespacePath},
	)

	c.bytesOutTotalDesc = prometheus.NewDesc(
		"ipvs_bytes_out_total",
		"The total number of outgoing bytes from a virtual server",
		[]string{"fwmark", "port"},
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

// RunInNetns executes a given function `f` in the network
// namespace as configured via `NamespacePath` in
// `CollectorConfig`.
func (c *Collector) RunInNetns(f func() (err error)) (err error) {
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

	err = netns.Set(*c.nsHandle)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to set network namespace")
		return
	}

	err = f()
	return
}

// Describe sends to the provided channel the set of all configured
// metric descriptions at the moment of collector registration.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.servicesTotalDesc
	ch <- c.connectionsTotalDesc
	ch <- c.bytesInTotalDesc
	ch <- c.bytesOutTotalDesc
}

type ServiceInfo struct {
	destinationServers []*libipvs.Destination
	destinationPort    uint16
	*libipvs.Service
}

func (c *Collector) GetServicesInfos() (infos []*ServiceInfo, err error) {
	var (
		destinations []*libipvs.Destination
		services     []*libipvs.Service
		mappings     map[uint32]uint16
	)

	services, err = c.ipvs.ListServices()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to retrieve ipvs services")
		return
	}

	mappings, err = c.mapper.GetMappings()
	if err != nil {
		err = errors.Wrapf(err,
			"failed to retrieve iptables fwmark mappings")
		return
	}

	infos = make([]*ServiceInfo, len(services))
	for ndx, service := range services {
		destPort, ok := mappings[service.FWMark]
		if !ok {
			err = errors.Errorf(
				"couldn't find destination port for fwmark %d",
				service.FWMark)
			return
		}

		destinations, err = c.ipvs.ListDestinations(service)
		if err != nil {
			err = errors.Wrapf(err,
				"failed to retrieve destinations from service")
			return
		}

		infos[ndx] = &ServiceInfo{
			Service:            service,
			destinationPort:    destPort,
			destinationServers: destinations,
		}
	}

	return
}

// Collect is called by Prometheus when collecting metrics.
// It's meant to list all of the services registered in IPVS in a
// given namespace and the corresponding metrics to the supplied
// channel.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	var (
		err   error
		infos []*ServiceInfo
	)

	f := func() (err error) {
		infos, err = c.GetServicesInfos()
		return
	}

	if c.nsHandle != nil {
		err = c.RunInNetns(f)
	} else {
		err = f()
	}
	if err != nil {
		c.logger.Error().
			Err(err).
			Msg("failed to retrieve ipvs info")
		return
	}

	ch <- prometheus.MustNewConstMetric(
		c.servicesTotalDesc,
		prometheus.GaugeValue,
		float64(len(infos)),
	)

	if len(infos) == 0 {
		return
	}

	for _, info := range infos {
		c.logger.Debug().
			Interface("info", info).
			Msg("reporting service")

		ch <- prometheus.MustNewConstMetric(
			c.connectionsTotalDesc,
			prometheus.CounterValue,
			float64(info.Stats.Connections),
			strconv.Itoa(int(info.FWMark)),
			strconv.Itoa(int(info.destinationPort)),
		)

		ch <- prometheus.MustNewConstMetric(
			c.bytesInTotalDesc,
			prometheus.CounterValue,
			float64(info.Stats.BytesIn),
			strconv.Itoa(int(info.FWMark)),
			strconv.Itoa(int(info.destinationPort)),
		)

		ch <- prometheus.MustNewConstMetric(
			c.bytesOutTotalDesc,
			prometheus.CounterValue,
			float64(info.Stats.BytesOut),
			strconv.Itoa(int(info.FWMark)),
			strconv.Itoa(int(info.destinationPort)),
		)
	}

	return
}
