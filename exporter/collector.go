package exporter

import (
	"os"

	"github.com/rs/zerolog"
)

type Collector struct {
	logger zerolog.Logger
}

func NewCollector() (c Collector, err error) {
	c.logger = zerolog.New(os.Stdout).
		With().
		Str("from", "collector").
		Logger()
	return
}
