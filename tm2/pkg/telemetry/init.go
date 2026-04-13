package telemetry

// Inspired by the example here:
// https://github.com/open-telemetry/opentelemetry-go/blob/main/example/prometheus/main.go

import (
	"fmt"
	"sync/atomic"

	"github.com/gnolang/gno/tm2/pkg/telemetry/config"
	"github.com/gnolang/gno/tm2/pkg/telemetry/metrics"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

var (
	globalConfig         config.Config
	telemetryInitialized atomic.Bool
)

// MetricsEnabled returns true if metrics have been initialized
func MetricsEnabled() bool {
	return globalConfig.MetricsEnabled
}

// TracesEnabled returns true if traces have been initialized
func TracesEnabled() bool {
	return globalConfig.TracesEnabled
}

// Init initializes the global telemetry
func Init(c config.Config) error {
	// Check if the metrics are enabled at all
	if !c.MetricsEnabled && !c.TracesEnabled {
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

	if c.MetricsEnabled {
		err := metrics.Init(c)
		if err != nil {
			return err
		}
	}
	if c.TracesEnabled {
		err := traces.Init(c)
		if err != nil {
			return err
		}
	}
	return nil
}

// Shutdown shuts down the global telemetry (metrics and traces)
func Shutdown() {
	if globalConfig.MetricsEnabled {
		metrics.Shutdown()
	}
	if globalConfig.TracesEnabled {
		traces.Shutdown()
	}
}
