package metrics

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

// Int64Gauge wraps the opentelemetry gauge
type Int64Gauge struct {
	meter metric.Meter
	gauge metric.Int64ObservableGauge
}

// NewInt64Gauge creates a new int64 gauge using the provided meter
func NewInt64Gauge(
	name string,
	description string,
	meter metric.Meter,
) (*Int64Gauge, error) {
	// Create the observable gauge
	gauge, err := meter.Int64ObservableGauge(name, metric.WithDescription(description))
	if err != nil {
		return nil, fmt.Errorf("unable to create gauge, %w", err)
	}

	return &Int64Gauge{
		meter: meter,
		gauge: gauge,
	}, nil
}

func (g *Int64Gauge) Observe(value int64) {
	_, _ = g.meter.RegisterCallback(
		func(_ context.Context, observer metric.Observer) error {
			observer.ObserveInt64(g.gauge, value)

			return nil
		},
	)
}
