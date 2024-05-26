package telemetry

// Inspired by the example here:
// https://github.com/open-telemetry/opentelemetry-go/blob/main/example/prometheus/main.go

import (
	"fmt"
	"sync/atomic"

	"github.com/gnolang/gno/tm2/pkg/telemetry/config"
	"github.com/gnolang/gno/tm2/pkg/telemetry/metrics"
)

var (
	globalConfig         config.Config
	telemetryInitialized atomic.Bool
)

// MetricsEnabled returns true if metrics have been initialized
func MetricsEnabled() bool {
	return globalConfig.MetricsEnabled
}

// Init initializes the global telemetry
func Init(c config.Config) error {
	// Check if the metrics are enabled at all
	if !c.MetricsEnabled {
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

	return metrics.Init(c)
}
