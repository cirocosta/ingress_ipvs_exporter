package exporter

import (
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type ExporterConfig struct {
	ListenAddress string
	TelemetryPath string
}

type Exporter struct {
	listenAddress string
	telemetryPath string
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

	exporter.listenAddress = cfg.ListenAddress
	exporter.telemetryPath = cfg.TelemetryPath
	exporter.logger = zerolog.New(os.Stdout).
		With().
		Str("from", "exporter").
		Logger()

	return
}

func (c Exporter) Listen() (err error) {
	c.logger.Debug().
		Str("listen-address", c.listenAddress).
		Str("telemetry-path", c.telemetryPath).
		Msg("starting http server")

	http.Handle(c.telemetryPath, promhttp.Handler())
	err = http.ListenAndServe(c.listenAddress, nil)
	if err != nil {
		err = errors.Wrapf(err,
			"failed listening on address %s",
			c.listenAddress)
		return
	}

	return
}
