package exporter

import (
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type ExporterConfig struct {
	ListenAddress string
	TelemetryPath string
	Collector     *Collector
}

type Exporter struct {
	listenAddress string
	telemetryPath string
	collector     *Collector
	logger        zerolog.Logger
}

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

func (e Exporter) Stop() (err error) {
	// TODO implement
	return
}
