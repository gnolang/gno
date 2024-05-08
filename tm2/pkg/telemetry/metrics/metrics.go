package metrics

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/telemetry/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	broadcastTxTimerKey = "broadcast_tx_hist"
	buildBlockTimerKey  = "build_block_hist"
)

var (
	// BroadcastTxTimer measures the transaction broadcast duration
	BroadcastTxTimer metric.Int64Histogram

	// BuildBlockTimer measures the block build duration
	BuildBlockTimer metric.Int64Histogram
)

func Init(config config.Config) error {
	// Use oltp metric exporter
	exp, err := otlpmetricgrpc.New(
		context.Background(),
		otlpmetricgrpc.WithEndpoint(config.ExporterEndpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("unable to create metrics exporter, %w", err)
	}

	provider := sdkMetric.NewMeterProvider(
		// Default period is 1m
		sdkMetric.WithReader(sdkMetric.NewPeriodicReader(exp)),
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
		broadcastTxTimerKey,
		metric.WithDescription("broadcast tx duration"),
		metric.WithUnit("ms"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}

	if BuildBlockTimer, err = meter.Int64Histogram(
		buildBlockTimerKey,
		metric.WithDescription("block build duration"),
		metric.WithUnit("ms"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}

	return nil
}
