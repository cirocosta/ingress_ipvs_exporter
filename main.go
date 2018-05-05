package main

import (
	"os"

	"github.com/alexflint/go-arg"
	"github.com/cirocosta/ipvs_exporter/collector"
	"github.com/cirocosta/ipvs_exporter/exporter"
	"github.com/rs/zerolog"
)

type config struct {
	ListenAddress string `arg:"--listen-address,help:address to set the http server to listen to"`
	TelemetryPath string `arg:"--telemetry-path,help:endpoint to receive scrape requests from prometheus"`
	NamespacePath string `arg:"--namespace-path,required,help:absolute path to the network namespace where ipv is configured"`
}

var (
	args = &config{
		ListenAddress: ":9100",
		TelemetryPath: "/metrics",
		NamespacePath: "",
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

	collector, err := collector.NewCollector(collector.CollectorConfig{
		NamespacePath: args.NamespacePath,
	})
	must(err)

	exporter, err := exporter.NewExporter(exporter.ExporterConfig{
		ListenAddress: args.ListenAddress,
		TelemetryPath: args.TelemetryPath,
		Collector:     &collector,
	})
	must(err)

	err = exporter.Listen()
	must(err)
}
