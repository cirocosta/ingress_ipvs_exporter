package main

import (
	"github.com/alexflint/go-arg"
	_ "github.com/cirocosta/ipvs_exporter/collector"
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
)

func main() {
	arg.MustParse(args)
}
