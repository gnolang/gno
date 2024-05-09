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

	vmQueryCallsKey  = "vm_query_calls_counter"
	vmQueryErrorsKey = "vm_query_errors_counter"
	vmGasUsedKey     = "vm_gas_used_hist"
	vmCPUCyclesKey   = "vm_cpu_cycles_hist"
	vmExecMsgKey     = "vm_exec_msg_hist"

	validatorCountKey       = "validator_count_gauge"
	validatorVotingPowerKey = "validator_vp_gauge"
	blockIntervalKey        = "block_interval_hist"
	blockTxsKey             = "block_txs_gauge"
	blockSizeKey            = "block_size_gauge"
	totalTxsKey             = "total_txs_counter"
	latestHeightKey         = "latest_height_counter"

	httpRequestTimeKey = "http_request_time_hist"
	wsRequestTimeKey   = "ws_request_time_hist"
)

var (
	// Misc //

	// BroadcastTxTimer measures the transaction broadcast duration
	BroadcastTxTimer metric.Int64Histogram

	// Networking //

	// InboundPeers measures the active number of inbound peers
	InboundPeers *Int64Gauge

	// OutboundPeers measures the active number of outbound peers
	OutboundPeers *Int64Gauge

	// DialingPeers measures the active number of peers in the dialing state
	DialingPeers *Int64Gauge

	// Mempool //

	// NumMempoolTxs measures the number of transaction inside the mempool
	NumMempoolTxs *Int64Gauge

	// NumCachedTxs measures the number of transaction inside the mempool cache
	NumCachedTxs *Int64Gauge

	// Runtime //

	// VMQueryCalls measures the frequency of VM query calls
	VMQueryCalls metric.Int64Counter

	// VMQueryErrors measures the frequency of VM query errors
	VMQueryErrors metric.Int64Counter

	// VMGasUsed measures the VM gas usage
	VMGasUsed metric.Int64Histogram

	// VMCPUCycles measures the VM CPU cycles
	VMCPUCycles metric.Int64Histogram

	// VMExecMsgFrequency measures the frequency of VM operations
	VMExecMsgFrequency metric.Int64Counter

	// Consensus //

	// BuildBlockTimer measures the block build duration
	BuildBlockTimer metric.Int64Histogram

	// ValidatorsCount measures the size of the active validator set
	ValidatorsCount *Int64Gauge

	// ValidatorsVotingPower measures the total voting power of the active validator set
	ValidatorsVotingPower *Int64Gauge

	// BlockInterval measures the interval between 2 subsequent blocks
	BlockInterval metric.Int64Histogram

	// BlockTxs measures the number of transactions within the latest block
	BlockTxs *Int64Gauge

	// BlockSizeBytes measures the size of the latest block in bytes
	BlockSizeBytes *Int64Gauge

	// TotalTxs measures the total number of transactions on the network
	TotalTxs metric.Int64Counter

	// LatestHeight measures the total number of committed network blocks
	LatestHeight *Int64Gauge

	// RPC //

	// HTTPRequestTime measures the HTTP request response time
	HTTPRequestTime metric.Int64Histogram

	// WSRequestTime measures the WS request response time
	WSRequestTime metric.Int64Histogram
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

	// Runtime //
	if VMQueryCalls, err = meter.Int64Counter(
		vmQueryCallsKey,
		metric.WithDescription("vm query call frequency"),
	); err != nil {
		return fmt.Errorf("unable to create counter, %w", err)
	}

	if VMQueryErrors, err = meter.Int64Counter(
		vmQueryErrorsKey,
		metric.WithDescription("vm query errors call frequency"),
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

	if VMExecMsgFrequency, err = meter.Int64Counter(
		vmExecMsgKey,
		metric.WithDescription("vm msg operation call frequency"),
	); err != nil {
		return fmt.Errorf("unable to create counter, %w", err)
	}

	// Consensus //
	if ValidatorsCount, err = NewInt64Gauge(
		validatorCountKey,
		"size of the active validator set",
		meter,
	); err != nil {
		return fmt.Errorf("unable to create gauge, %w", err)
	}

	if ValidatorsVotingPower, err = NewInt64Gauge(
		validatorVotingPowerKey,
		"total voting power of the active validator set",
		meter,
	); err != nil {
		return fmt.Errorf("unable to create gauge, %w", err)
	}

	if BlockInterval, err = meter.Int64Histogram(
		blockIntervalKey,
		metric.WithDescription("interval between 2 subsequent blocks"),
		metric.WithUnit("s"),
	); err != nil {
		return fmt.Errorf("unable to create histogram, %w", err)
	}

	if BlockTxs, err = NewInt64Gauge(
		blockTxsKey,
		"number of transactions within the latest block",
		meter,
	); err != nil {
		return fmt.Errorf("unable to create gauge, %w", err)
	}

	if BlockSizeBytes, err = NewInt64Gauge(
		blockSizeKey,
		"size of the latest block in bytes",
		meter,
	); err != nil {
		return fmt.Errorf("unable to create gauge, %w", err)
	}

	if TotalTxs, err = meter.Int64Counter(
		totalTxsKey,
		metric.WithDescription("total number of transactions on the network"),
	); err != nil {
		return fmt.Errorf("unable to create counter, %w", err)
	}

	if LatestHeight, err = NewInt64Gauge(
		latestHeightKey,
		"total number of committed network blocks",
		meter,
	); err != nil {
		return fmt.Errorf("unable to create gauge, %w", err)
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
