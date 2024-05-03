package telemetry

// Inspired by the example here:
// https://github.com/open-telemetry/opentelemetry-go/blob/main/example/prometheus/main.go

import (
	"sync/atomic"

	opentlmConfig "github.com/gnolang/gno/tm2/pkg/telemetry/opentlm/config"
	"github.com/gnolang/gno/tm2/pkg/telemetry/opentlm/metrics"
	"github.com/gnolang/gno/tm2/pkg/telemetry/prom"
)

var (
	isConfigSet atomic.Bool

	promCollector    Collector
	opentlmCollector Collector
)

type Config struct {
	OpenTelemetry *opentlmConfig.Config `toml:"opentelemetry"`
	Prometheus    *prom.Config          `toml:"prometheus"`
}

func DefaultTelemetryConfig() *Config {
	return &Config{
		OpenTelemetry: opentlmConfig.DefaultConfig(),
		Prometheus:    prom.DefaultConfig(),
	}
}

func TestTelemetryConfig() *Config {
	return DefaultTelemetryConfig()
}

// MetricsEnabled returns true if metrics have been initialized.
func MetricsEnabled() bool {
	return promCollector != nil || opentlmCollector != nil
}

// Init sets the configuration for telemetry to c, and if telemetry is enabled,
// starts tracking.
// Init may only be called once. Multiple calls to Init will panic.
func Init(c Config) error {
	if !isConfigSet.CompareAndSwap(false, true) {
		panic("telemetry configuration has already been set and initialised")
	}

	var err error

	// Initialize opentelemetry metrics to be collected.
	if c.OpenTelemetry.MetricsEnabled {
		opentlmCollector, err = metrics.Init(c.OpenTelemetry)
		if err != nil {
			return err
		}
	}

	// Initialize prometheus metrics to be collected
	if c.Prometheus.MetricsEnabled {
		promCollector, err = prom.Init(c.Prometheus)
		if err != nil {
			return err
		}
	}
	return nil
}
