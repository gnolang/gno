package telemetry

// Inspired by the example here:
// https://github.com/open-telemetry/opentelemetry-go/blob/main/example/prometheus/main.go

import (
	"context"

	"github.com/gnolang/gno/telemetry/metrics"
	"github.com/gnolang/gno/telemetry/options"
)

const (
	defaultMeterName          = "gno.land"
	defaultServiceName        = "gno.land"
	defaultPort        uint64 = 4591
)

var config options.Config

// MetricsEnabled returns true if metrics have been initialized.
func MetricsEnabled() bool {
	return config.MetricsEnabled
}

// Init can indicate both, either, or none of metrics and tracing depending on the options provided.
func Init(ctx context.Context, options ...Option) error {

	config.Port = defaultPort
	config.MeterName = defaultMeterName
	config.ServiceName = defaultServiceName
	for _, opt := range options {
		opt(&config)
	}

	// Initialize metrics to be collected.
	if config.MetricsEnabled {
		if err := metrics.Init(ctx, config); err != nil {
			return err
		}
	}

	return nil
}
