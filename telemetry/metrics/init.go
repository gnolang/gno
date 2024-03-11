package metrics

import (
	"context"
	"math/rand"

	"github.com/gnolang/gno/telemetry/exporter"
	"github.com/gnolang/gno/telemetry/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
)

var (
	ctx context.Context

	// Metrics.
	BroadcastTxTimer Int64Collector
	BuildBlockTimer  Int64Collector
)

func Init(setCtx context.Context, config options.Config) error {
	if config.ExporterEndpoint == "" {
		return exporter.ErrEndpointNotSet
	}

	ctx = setCtx

	// Use oltp metric exporter
	exporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(config.ExporterEndpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return err
	}

	provider := sdkMetric.NewMeterProvider(sdkMetric.WithReader(sdkMetric.NewPeriodicReader(exporter)))
	otel.SetMeterProvider(provider)
	meter := provider.Meter(config.MeterName)

	broadcastTxTimer, err := meter.Int64Histogram(
		"broadcast_tx_hist",
		metric.WithDescription("broadcast tx duration"),
		// metric.WithExplicitBucketBoundaries(0, 16, 32, 64, 128, 256, 512),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}
	BroadcastTxTimer = Int64Histogram{
		Int64Histogram: broadcastTxTimer,
		useFakeMetrics: config.UseFakeMetrics,
		fakeRangeStart: 5,
		fakeRangeEnd:   250,
	}

	buildBlockTimer, err := meter.Int64Histogram(
		"build_block_hist",
		metric.WithDescription("block build duration"),
		// metric.WithExplicitBucketBoundaries(0, 16, 32, 64, 128, 256, 512),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return err
	}
	BuildBlockTimer = Int64Histogram{
		Int64Histogram: buildBlockTimer,
		useFakeMetrics: config.UseFakeMetrics,
		fakeRangeStart: 0,
		fakeRangeEnd:   150,
	}

	return nil
}

type Int64Collector interface {
	Collect(int64)
}

type Int64Histogram struct {
	metric.Int64Histogram

	useFakeMetrics bool
	fakeRangeStart int64
	fakeRangeEnd   int64
}

func (h Int64Histogram) Collect(value int64) {
	if h.useFakeMetrics {
		value = rand.Int63n(h.fakeRangeEnd) + h.fakeRangeStart
	}

	h.Int64Histogram.Record(ctx, value)
}

type Int64Counter struct {
	metric.Int64Counter

	useFakeMetrics bool
	fakeRangeStart int64
	fakeRangeEnd   int64
}

func (c Int64Counter) Collect(value int64) {
	if c.useFakeMetrics {
		value = rand.Int63n(c.fakeRangeEnd) + c.fakeRangeStart
	}

	c.Int64Counter.Add(ctx, value)
}
