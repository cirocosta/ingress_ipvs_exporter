package collector

import (
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type CollectorConfig struct {
	ListenAddress string
	TelemetryPath string
}

type Collector struct {
	listenAddress string
	telemetryPath string
	logger        zerolog.Logger
}

func New(cfg CollectorConfig) (collector Collector, err error) {
	if cfg.ListenAddress == "" {
		err = errors.Errorf("ListenAddress must be specified")
		return
	}

	if cfg.TelemetryPath == "" {
		err = errors.Errorf("TelemetryPath must be specified")
		return
	}

	collector.listenAddress = cfg.ListenAddress
	collector.telemetryPath = cfg.TelemetryPath
	collector.logger = zerolog.New(os.Stdout).
		With().
		Str("from", "collector").
		Logger()

	return
}

func (c Collector) Start() (err error) {
	http.Handle(c.telemetryPath, promhttp.Handler())
	return
}
