package main

import (
	"os"

	"github.com/alexflint/go-arg"
	"github.com/rs/zerolog"

	. "github.com/cirocosta/ipvs_exporter/exporter"
)

type config struct {
	ListenAddress string `arg:"--listen-address"`
	TelemetryPath string `arg:"--telemetry-path"`
}

var (
	args = &config{
		ListenAddress: ":9100",
		TelemetryPath: "/metrics",
	}
	logger = zerolog.New(os.Stdout)
)

func must(err error) {
	if err == nil {
		return
	}

	logger.Error().
		Err(err).
		Msg("main execution failed")
	os.Exit(1)
}

func main() {
	arg.MustParse(args)

	exporter, err := NewExporter(ExporterConfig{
		ListenAddress: args.ListenAddress,
		TelemetryPath: args.TelemetryPath,
	})
	must(err)

	err = exporter.Listen()
	must(err)
}
