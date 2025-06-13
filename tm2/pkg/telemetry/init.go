package telemetry

// Inspired by the example here:
// https://github.com/open-telemetry/opentelemetry-go/blob/main/example/prometheus/main.go

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/telemetry/config"
	"github.com/gnolang/gno/tm2/pkg/telemetry/metrics"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

// MetricsEnabled returns true if metrics have been initialized
func MetricsEnabled() bool {
	return config.GetGlobalConfig().MetricsEnabled
}

// TracesEnabled returns true if traces have been initialized
func TracesEnabled() bool {
	return config.GetGlobalConfig().TracesEnabled
}

func InitMetrics(c config.Config) (*sdkMetric.MeterProvider, error) {
	if !c.MetricsEnabled {
		return nil, nil
	}

	// Validate the configuration
	if err := c.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("unable to validate config, %w", err)
	}

	// Check if it's been enabled already
	if !config.SetMetricsInitialized() {
		return nil, nil
	}

	config.SetGlobalConfig(c)

	return metrics.Init(c)
}

func InitTraces(c config.Config) (*sdkTrace.TracerProvider, error) {
	if !c.TracesEnabled {
		return nil, nil
	}

	// Validate the configuration
	if err := c.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("unable to validate config, %w", err)
	}

	// Check if it's been enabled already
	if !config.SetTracesInitialized() {
		return nil, nil
	}

	config.SetGlobalConfig(c)

	return traces.Init(c)
}
