// exporter defines the implementation of ipvs exporter internals,
// providing means for it to gather IPVS metrics from well defined
// namespaces as well as registering the exporter details with the
// prometheus client.
//
// ps.: The package is meant to be used by the main command only as it
// doesn't provide any interface for generic loggers.
package exporter

import (
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// ExporterConfig provides the configuration necessary to
// instantiate a new Exporter via `NewExporter`.
type ExporterConfig struct {
	// ListenAddress is the address used by prometheus
	// to listen for scraping requests.
	//
	// Examples:
	// - :8080
	// - 127.0.0.2:1313
	ListenAddress string

	// TelemetryPath configures the path under which
	// the prometheus metrics are reported.
	//
	// For instance:
	// - /metrics
	// - /telemetry
	TelemetryPath string

	// Collector is an already instantiated Collector that
	// implements the Prometheus collector interface so
	// prometheus can ask it for metrics and metric descriptions
	// to expose under the configured telemetry path.
	Collector *Collector
}

// Exporter is responsible for initiating the Prometheus HTTP
// server with the IPVS collector registered.
//
// It must be instantiated via the `NewExporter` so the configuration
// can be properly checked.
type Exporter struct {
	listenAddress string
	telemetryPath string
	collector     *Collector
	logger        zerolog.Logger
}

// NewExporter instantiates an Exporter, validating the provided
// configuration and registering the IPVS collector with the
// prometheus client.
func NewExporter(cfg ExporterConfig) (exporter Exporter, err error) {
	if cfg.ListenAddress == "" {
		err = errors.Errorf("ListenAddress must be specified")
		return
	}

	if cfg.TelemetryPath == "" {
		err = errors.Errorf("TelemetryPath must be specified")
		return
	}

	if cfg.Collector == nil {
		err = errors.Errorf("Collector must be specified")
		return
	}

	exporter.collector = cfg.Collector
	exporter.listenAddress = cfg.ListenAddress
	exporter.telemetryPath = cfg.TelemetryPath
	exporter.logger = zerolog.New(os.Stdout).
		With().
		Str("from", "exporter").
		Logger()

	err = prometheus.Register(exporter.collector)
	if err != nil {
		err = errors.Wrapf(err, "failed to register ipvs collector")
		return
	}

	return
}

// Listen initiates the HTTP server using the configurations
// provided via ExporterConfig.
//
// This is a blocking method - make sure you either make use of
// goroutines to not block if needed.
func (e Exporter) Listen() (err error) {
	e.logger.Debug().
		Str("listen-address", e.listenAddress).
		Str("telemetry-path", e.telemetryPath).
		Msg("starting http server")

	http.Handle(e.telemetryPath, promhttp.Handler())
	err = http.ListenAndServe(e.listenAddress, nil)
	if err != nil {
		err = errors.Wrapf(err,
			"failed listening on address %s",
			e.listenAddress)
		return
	}

	return
}
