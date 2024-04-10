package telemetry

// Inspired by the example here:
// https://github.com/open-telemetry/opentelemetry-go/blob/main/example/prometheus/main.go

import (
	"github.com/gnolang/gno/telemetry/metrics"
	"github.com/gnolang/gno/telemetry/options"
)

const (
	defaultMeterName   = "gno.land"
	defaultServiceName = "gno.land"
)

var config options.Config

// MetricsEnabled returns true if metrics have been initialized.
func MetricsEnabled() bool {
	return config.MetricsEnabled
}

// Init will initialize metrics with the options provided. This function may also initialize tracing when
// this is something that we want to support.
func Init(options ...Option) error {
	config.MeterName = defaultMeterName
	config.ServiceName = defaultServiceName
	for _, opt := range options {
		opt(&config)
	}

	// Initialize metrics to be collected.
	if config.MetricsEnabled {
		if err := metrics.Init(config); err != nil {
			return err
		}
	}

	return nil
}
