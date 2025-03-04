package telemetry

// Inspired by the example here:
// https://github.com/open-telemetry/opentelemetry-go/blob/main/example/prometheus/main.go

import (
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/gnolang/gno/tm2/pkg/telemetry/config"
	"github.com/gnolang/gno/tm2/pkg/telemetry/metrics"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
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
func Init(c config.Config, logger *slog.Logger) (*sdkTrace.TracerProvider, *sdkMetric.MeterProvider, error) {
	anyTelemetryEnabled := c.MetricsEnabled || c.TracesEnabled
	if !anyTelemetryEnabled {
		return nil, nil, nil
	}

	// Validate the configuration
	if err := c.ValidateBasic(); err != nil {
		return nil, nil, fmt.Errorf("unable to validate config, %w", err)
	}

	// Check if it's been enabled already
	if !telemetryInitialized.CompareAndSwap(false, true) {
		return nil, nil, nil
	}

	// Update the global configuration
	globalConfig = c

	// Check if the metrics are enabled at all
	var metricsProvider *sdkMetric.MeterProvider
	var tracesProvider *sdkTrace.TracerProvider
	var err error
	if c.MetricsEnabled {
		metricsProvider, err = metrics.Init(c)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to initialize metrics, %w", err)
		}
		logger.Info("Metrics initialized")
	}

	if c.TracesEnabled {
		tracesProvider, err = traces.Init(c)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to initialize traces, %w", err)
		}
		logger.Info("Traces initialized")
	}

	return tracesProvider, metricsProvider, nil
}
