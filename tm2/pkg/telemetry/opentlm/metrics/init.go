package metrics

import (
	"context"
	"time"

	"github.com/gnolang/gno/tm2/pkg/telemetry/opentlm/config"
	"github.com/gnolang/gno/tm2/pkg/telemetry/opentlm/exporter"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// var (
// 	// Metrics.
// 	BroadcastTxTimer metric.Int64Histogram
// 	BuildBlockTimer  metric.Int64Histogram
// )

// Collector is complient with the telemetry.Collector interface
type Collector struct {
	broadcastTxTimer     metric.Int64Histogram
	buildBlockTimer      metric.Int64Histogram
	blockIntervalSeconds metric.Int64Histogram
}

// Collector is complient with the telemetry.Collector interface
func (c *Collector) RecordBroadcastTxTimer(data time.Duration) {
	c.broadcastTxTimer.Record(context.Background(), data.Milliseconds())
}

func (c *Collector) RecordBuildBlockTimer(data time.Duration) {
	c.buildBlockTimer.Record(context.Background(), data.Milliseconds())
}

func (c *Collector) RecordBlockIntervalSeconds(data time.Duration) {
	c.blockIntervalSeconds.Record(context.Background(), data.Milliseconds())
}

func Init(config *config.Config) (*Collector, error) {
	if config.ExporterEndpoint == "" {
		return nil, exporter.ErrEndpointNotSet
	}

	collector := &Collector{}

	// Use oltp metric exporter.
	exporter, err := otlpmetricgrpc.New(
		context.Background(),
		otlpmetricgrpc.WithEndpoint(config.ExporterEndpoint),
		otlpmetricgrpc.WithInsecure(), // TODO: enable security
	)
	if err != nil {
		return nil, err
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

	if collector.broadcastTxTimer, err = meter.Int64Histogram(
		"broadcast_tx_hist",
		metric.WithDescription("broadcast tx duration"),
		metric.WithUnit("ms"),
	); err != nil {
		return nil, err
	}

	if collector.buildBlockTimer, err = meter.Int64Histogram(
		"build_block_hist",
		metric.WithDescription("block build duration"),
		metric.WithUnit("ms"),
	); err != nil {
		return nil, err
	}

	if collector.blockIntervalSeconds, err = meter.Int64Histogram(
		"block_interval_seconds",
		metric.WithDescription("block interval duration"),
		metric.WithUnit("s"),
	); err != nil {
		return nil, err
	}

	return collector, nil
}
