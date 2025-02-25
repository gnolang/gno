package telemetry

// Inspired by the example here:
// https://github.com/open-telemetry/opentelemetry-go/blob/main/example/prometheus/main.go

import (
	"fmt"
	"sync/atomic"

	"github.com/gnolang/gno/tm2/pkg/telemetry/config"
	"github.com/gnolang/gno/tm2/pkg/telemetry/metrics"
	"github.com/gnolang/gno/tm2/pkg/telemetry/tracing"
)

var (
	globalConfig         config.Config
	telemetryInitialized atomic.Bool
)

// MetricsEnabled returns true if metrics have been initialized
func MetricsEnabled() bool {
	return globalConfig.MetricsEnabled
}

// TracingEnabled returns true if tracing has been initialized
func TracingEnabled() bool {
	return globalConfig.TracingEnabled
}

// Init initializes the global telemetry
func Init(c config.Config) error {
	anyTelemetryEnabled := c.MetricsEnabled || c.TracingEnabled
	if !anyTelemetryEnabled {
		return nil
	}

	// Validate the configuration
	if err := c.ValidateBasic(); err != nil {
		return fmt.Errorf("unable to validate config, %w", err)
	}

	// Check if it's been enabled already
	if !telemetryInitialized.CompareAndSwap(false, true) {
		return nil
	}

	// Update the global configuration
	globalConfig = c

	// Check if the metrics are enabled at all
	if !c.MetricsEnabled {
		if err := metrics.Init(c); err != nil {
			return fmt.Errorf("unable to initialize metrics, %w", err)
		}
	}

	if !c.TracingEnabled {
		if err := tracing.Init(c); err != nil {
			return fmt.Errorf("unable to initialize tracing, %w", err)
		}
	}

	return nil
}
