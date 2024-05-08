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

	inboundPeersKey  = "inbound_peers_gauge"
	outboundPeersKey = "outbound_peers_gauge"
	dialingPeersKey  = "dialing_peers_gauge"

	numMempoolTxsKey = "num_mempool_txs_gauge"
	numCachedTxsKey  = "num_cached_txs_gauge"
)

var (
	// BroadcastTxTimer measures the transaction broadcast duration
	BroadcastTxTimer metric.Int64Histogram

	// BuildBlockTimer measures the block build duration
	BuildBlockTimer metric.Int64Histogram

	// InboundPeers measures the active number of inbound peers
	InboundPeers *Int64Gauge

	// OutboundPeers measures the active number of outbound peers
	OutboundPeers *Int64Gauge

	// DialingPeers measures the active number of peers in the dialing state
	DialingPeers *Int64Gauge

	// NumMempoolTxs measures the number of transaction inside the mempool
	NumMempoolTxs *Int64Gauge

	// NumCachedTxs measures the number of transaction inside the mempool cache
	NumCachedTxs *Int64Gauge
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

	// Networking //
	if InboundPeers, err = NewInt64Gauge(
		inboundPeersKey,
		"inbound peer count",
		meter,
	); err != nil {
		return fmt.Errorf("unable to create gauge, %w", err)
	}

	if OutboundPeers, err = NewInt64Gauge(
		outboundPeersKey,
		"outbound peer count",
		meter,
	); err != nil {
		return fmt.Errorf("unable to create gauge, %w", err)
	}

	if DialingPeers, err = NewInt64Gauge(
		dialingPeersKey,
		"dialing peer count",
		meter,
	); err != nil {
		return fmt.Errorf("unable to create gauge, %w", err)
	}

	// Mempool //
	if NumMempoolTxs, err = NewInt64Gauge(
		numMempoolTxsKey,
		"valid mempool transaction count",
		meter,
	); err != nil {
		return fmt.Errorf("unable to create gauge, %w", err)
	}

	if NumCachedTxs, err = NewInt64Gauge(
		numCachedTxsKey,
		"cached mempool transaction count",
		meter,
	); err != nil {
		return fmt.Errorf("unable to create gauge, %w", err)
	}

	return nil
}
