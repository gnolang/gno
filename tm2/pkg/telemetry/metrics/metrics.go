package metrics

import (
	"context"
	"fmt"
	"net/url"

	"github.com/gnolang/gno/tm2/pkg/telemetry/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	buildBlockTimerKey = "build_block_hist"

	inboundPeersKey  = "inbound_peers_gauge"
	outboundPeersKey = "outbound_peers_gauge"

	numMempoolTxsKey = "num_mempool_txs_hist"
	numCachedTxsKey  = "num_cached_txs_hist"

	vmExecMsgKey   = "vm_exec_msg_counter"
	vmGasUsedKey   = "vm_gas_used_hist"
	vmCPUCyclesKey = "vm_cpu_cycles_hist"

	validatorCountKey       = "validator_count_hist"
	validatorVotingPowerKey = "validator_vp_hist"
	blockIntervalKey        = "block_interval_hist"
	blockTxsKey             = "block_txs_hist"
	blockSizeKey            = "block_size_hist"
	gasPriceKey             = "block_gas_price_hist"

	httpRequestTimeKey = "http_request_time_hist"
	wsRequestTimeKey   = "ws_request_time_hist"
)

var (
	// Networking //

	// InboundPeers measures the active number of inbound peers
	InboundPeers metric.Int64Gauge

	// OutboundPeers measures the active number of outbound peers
	OutboundPeers metric.Int64Gauge

	// Mempool //

	// NumMempoolTxs measures the number of transaction inside the mempool
	NumMempoolTxs metric.Int64Histogram

	// NumCachedTxs measures the number of transaction inside the mempool cache
	NumCachedTxs metric.Int64Histogram

	// Runtime //

	// VMExecMsgFrequency measures the frequency of VM operations
	VMExecMsgFrequency metric.Int64Counter

	// VMGasUsed measures the VM gas usage
	VMGasUsed metric.Int64Histogram

	// VMCPUCycles measures the VM CPU cycles
	VMCPUCycles metric.Int64Histogram

	// Consensus //

	// BuildBlockTimer measures the block build duration
	BuildBlockTimer metric.Int64Histogram

	// ValidatorsCount measures the size of the active validator set
	ValidatorsCount metric.Int64Histogram

	// ValidatorsVotingPower measures the total voting power of the active validator set
	ValidatorsVotingPower metric.Int64Histogram

	// BlockInterval measures the interval between 2 subsequent blocks
	BlockInterval metric.Int64Histogram

	// BlockTxs measures the number of transactions within the latest block
	BlockTxs metric.Int64Histogram

	// BlockSizeBytes measures the size of the latest block in bytes
	BlockSizeBytes metric.Int64Histogram

	// BlockGasPriceAmount measures the block gas price of the last block
	BlockGasPriceAmount metric.Int64Histogram

	// RPC //

	// HTTPRequestTime measures the HTTP request response time
	HTTPRequestTime metric.Int64Histogram

	// WSRequestTime measures the WS request response time
	WSRequestTime metric.Int64Histogram

	provider *sdkMetric.MeterProvider
)

func Init(config config.Config) error {
	var (
		ctx = context.Background()
		exp sdkMetric.Exporter
	)

	u, err := url.Parse(config.ExporterEndpoint)
	if err != nil {
		return fmt.Errorf("error parsing exporter endpoint: %s, %w", config.ExporterEndpoint, err)
	}

	// Use oltp metric exporter with http/https or grpc
	switch u.Scheme {
	case "http", "https":
		exp, err = otlpmetrichttp.New(
			ctx,
			otlpmetrichttp.WithEndpointURL(config.ExporterEndpoint),
		)
		if err != nil {
			return fmt.Errorf("unable to create http metrics exporter, %w", err)
		}
	default:
		exp, err = otlpmetricgrpc.New(
			ctx,
			otlpmetricgrpc.WithEndpoint(config.ExporterEndpoint),
			otlpmetricgrpc.WithInsecure(),
		)
		if err != nil {
			return fmt.Errorf("unable to create grpc metrics exporter, %w", err)
		}
	}

	provider = sdkMetric.NewMeterProvider(
		// Default period is 1m
		sdkMetric.WithReader(sdkMetric.NewPeriodicReader(exp)),
		sdkMetric.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(config.ServiceName),
				semconv.ServiceVersionKey.String("1.0.0"),
				semconv.ServiceInstanceIDKey.String(config.ServiceInstanceID),
			),
		),
	)
	otel.SetMeterProvider(provider)
	meter := provider.Meter(config.MeterName)

	if BuildBlockTimer, err = meter.Int64Histogram(
		buildBlockTimerKey,
		metric.WithDescription("block build duration"),
		metric.WithUnit("ms"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}

	// Networking //
	if InboundPeers, err = meter.Int64Gauge(
		inboundPeersKey,
		metric.WithDescription("inbound peer count"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}
	// Initialize InboundPeers Gauge
	InboundPeers.Record(ctx, 0)

	if OutboundPeers, err = meter.Int64Gauge(
		outboundPeersKey,
		metric.WithDescription("outbound peer count"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}

	// Initialize OutboundPeers Gauge
	OutboundPeers.Record(ctx, 0)

	// Mempool //
	if NumMempoolTxs, err = meter.Int64Histogram(
		numMempoolTxsKey,
		metric.WithDescription("valid mempool transaction count"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}

	if NumCachedTxs, err = meter.Int64Histogram(
		numCachedTxsKey,
		metric.WithDescription("cached mempool transaction count"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}

	// Runtime //
	if VMExecMsgFrequency, err = meter.Int64Counter(
		vmExecMsgKey,
		metric.WithDescription("vm msg operation call frequency"),
	); err != nil {
		return fmt.Errorf("unable to create counter, %w", err)
	}

	if VMGasUsed, err = meter.Int64Histogram(
		vmGasUsedKey,
		metric.WithDescription("VM gas used"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}

	if VMCPUCycles, err = meter.Int64Histogram(
		vmCPUCyclesKey,
		metric.WithDescription("VM CPU cycles"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}

	// Consensus //
	if ValidatorsCount, err = meter.Int64Histogram(
		validatorCountKey,
		metric.WithDescription("size of the active validator set"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}

	if ValidatorsVotingPower, err = meter.Int64Histogram(
		validatorVotingPowerKey,
		metric.WithDescription("total voting power of the active validator set"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}

	if BlockInterval, err = meter.Int64Histogram(
		blockIntervalKey,
		metric.WithDescription("interval between 2 subsequent blocks"),
		metric.WithUnit("s"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}

	if BlockTxs, err = meter.Int64Histogram(
		blockTxsKey,
		metric.WithDescription("number of transactions within the latest block"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}

	if BlockSizeBytes, err = meter.Int64Histogram(
		blockSizeKey,
		metric.WithDescription("size of the latest block in bytes"),
		metric.WithUnit("B"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}

	if BlockGasPriceAmount, err = meter.Int64Histogram(
		gasPriceKey,
		metric.WithDescription("block gas price"),
		metric.WithUnit("token"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}
	// RPC //

	if HTTPRequestTime, err = meter.Int64Histogram(
		httpRequestTimeKey,
		metric.WithDescription("http request response time"),
		metric.WithUnit("ms"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}

	if WSRequestTime, err = meter.Int64Histogram(
		wsRequestTimeKey,
		metric.WithDescription("ws request response time"),
		metric.WithUnit("ms"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}

	return nil
}

func Shutdown() {
	if provider != nil {
		provider.Shutdown(context.Background())
	}
}
