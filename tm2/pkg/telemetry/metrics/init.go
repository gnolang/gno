package metrics

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/telemetry/config"
	"github.com/gnolang/gno/tm2/pkg/telemetry/exporter"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var (
	// Metrics.
	BroadcastTxTimer metric.Int64Histogram
	BuildBlockTimer  metric.Int64Histogram
)

func Init(config config.Config) error {
	if config.ExporterEndpoint == "" {
		return exporter.ErrEndpointNotSet
	}

	// Use oltp metric exporter.
	exporter, err := otlpmetricgrpc.New(
		context.Background(),
		otlpmetricgrpc.WithEndpoint(config.ExporterEndpoint),
		otlpmetricgrpc.WithInsecure(), // TODO: enable security
	)
	if err != nil {
		return err
	}

	provider := sdkMetric.NewMeterProvider(
		// Default period is 1m.
		sdkMetric.WithReader(sdkMetric.NewPeriodicReader(exporter)),
		sdkMetric.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(config.ServiceName),
				semconv.ServiceVersionKey.String("1.0.0"),
				semconv.ServiceInstanceIDKey.String("gno-node-1"),
			),
		),
	)
	otel.SetMeterProvider(provider)
	meter := provider.Meter(config.MeterName)

	if BroadcastTxTimer, err = meter.Int64Histogram(
		"broadcast_tx_hist",
		metric.WithDescription("broadcast tx duration"),
		metric.WithUnit("ms"),
	); err != nil {
		return err
	}

	if BuildBlockTimer, err = meter.Int64Histogram(
		"build_block_hist",
		metric.WithDescription("block build duration"),
		metric.WithUnit("ms"),
	); err != nil {
		return err
	}

	return nil
}
