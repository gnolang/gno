package telemetry

// Inspired by the example here:
// https://github.com/open-telemetry/opentelemetry-go/blob/main/example/prometheus/main.go

import (
	"sync/atomic"

	"github.com/gnolang/gno/tm2/pkg/telemetry/config"
	"github.com/gnolang/gno/tm2/pkg/telemetry/metrics"
)

var (
	globalConfig    config.Config
	globalConfigSet atomic.Bool
)

// MetricsEnabled returns true if metrics have been initialized.
func MetricsEnabled() bool {
	return globalConfig.MetricsEnabled
}

// Init sets the configuration for telemetry to c, and if telemetry is enabled,
// starts tracking.
// Init may only be called once. Multiple calls to Init will panic.
func Init(c config.Config) error {
	if !globalConfigSet.CompareAndSwap(false, true) {
		panic("telemetry configuration has already been set and initialised")
	}
	globalConfig = c
	// Initialize metrics to be collected.
	if c.MetricsEnabled {
		return metrics.Init(c)
	}
	return nil
}
