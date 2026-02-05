package client

import (
	"context"
	"fmt"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/abci"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/blocks"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/consensus"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/health"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/mempool"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/net"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/status"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/tx"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client/batch"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client/http"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client/ws"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/rs/xid"
)

const defaultTimeout = 60 * time.Second

const (
	statusMethod             = "status"
	abciInfoMethod           = "abci_info"
	abciQueryMethod          = "abci_query"
	broadcastTxCommitMethod  = "broadcast_tx_commit"
	broadcastTxAsyncMethod   = "broadcast_tx_async"
	broadcastTxSyncMethod    = "broadcast_tx_sync"
	unconfirmedTxsMethod     = "unconfirmed_txs"
	numUnconfirmedTxsMethod  = "num_unconfirmed_txs"
	netInfoMethod            = "net_info"
	dumpConsensusStateMethod = "dump_consensus_state"
	consensusStateMethod     = "consensus_state"
	consensusParamsMethod    = "consensus_params"
	healthMethod             = "health"
	blockchainMethod         = "blockchain"
	genesisMethod            = "genesis"
	blockMethod              = "block"
	blockResultsMethod       = "block_results"
	commitMethod             = "commit"
	txMethod                 = "tx"
	validatorsMethod         = "validators"
)

// RPCClient encompasses common RPC client methods
type RPCClient struct {
	requestTimeout time.Duration

	caller rpcclient.Client
}

// NewRPCClient creates a new RPC client instance with the given caller
func NewRPCClient(caller rpcclient.Client, opts ...Option) *RPCClient {
	c := &RPCClient{
		requestTimeout: defaultTimeout,
		caller:         caller,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// NewHTTPClient takes a remote endpoint in the form <protocol>://<host>:<port>,
// and returns an HTTP client that communicates with a Tendermint node over
// JSON RPC.
//
// Request batching is available for JSON RPC requests over HTTP, which conforms to
// the JSON RPC specification (https://www.jsonrpc.org/specification#batch). See
// the example for more details
func NewHTTPClient(rpcURL string, opts ...Option) (*RPCClient, error) {
	httpClient, err := http.NewClient(rpcURL)
	if err != nil {
		return nil, err
	}

	return NewRPCClient(httpClient, opts...), nil
}

// NewWSClient takes a remote endpoint in the form <protocol>://<host>:<port>,
// and returns a WS client that communicates with a Tendermint node over
// WS connection.
//
// Request batching is available for JSON RPC requests over WS, which conforms to
// the JSON RPC specification (https://www.jsonrpc.org/specification#batch). See
// the example for more details
func NewWSClient(rpcURL string, opts ...Option) (*RPCClient, error) {
	wsClient, err := ws.NewClient(rpcURL)
	if err != nil {
		return nil, err
	}

	return NewRPCClient(wsClient, opts...), nil
}

// Close attempts to gracefully close the RPC client
func (c *RPCClient) Close() error {
	return c.caller.Close()
}

// NewBatch creates a new RPC batch
func (c *RPCClient) NewBatch() *RPCBatch {
	return &RPCBatch{
		batch:     batch.NewBatch(c.caller),
		resultMap: make(map[string]any),
	}
}

func (c *RPCClient) Status(ctx context.Context, heightGte *int64) (*status.ResultStatus, error) {
	var v int64
	if heightGte != nil {
		v = *heightGte
	}

	return sendRequestCommon[status.ResultStatus](
		ctx,
		c.requestTimeout,
		c.caller,
		statusMethod,
		[]any{
			v,
		},
	)
}

func (c *RPCClient) ABCIInfo(ctx context.Context) (*abci.ResultABCIInfo, error) {
	return sendRequestCommon[abci.ResultABCIInfo](
		ctx,
		c.requestTimeout,
		c.caller,
		abciInfoMethod,
		nil,
	)
}

func (c *RPCClient) ABCIQuery(ctx context.Context, path string, data []byte) (*abci.ResultABCIQuery, error) {
	return c.ABCIQueryWithOptions(ctx, path, data, DefaultABCIQueryOptions)
}

func (c *RPCClient) ABCIQueryWithOptions(ctx context.Context, path string, data []byte, opts ABCIQueryOptions) (*abci.ResultABCIQuery, error) {
	return sendRequestCommon[abci.ResultABCIQuery](
		ctx,
		c.requestTimeout,
		c.caller,
		abciQueryMethod,
		[]any{
			path,
			data,
			opts.Height,
			opts.Prove,
		},
	)
}

func (c *RPCClient) BroadcastTxCommit(ctx context.Context, tx types.Tx) (*mempool.ResultBroadcastTxCommit, error) {
	return sendRequestCommon[mempool.ResultBroadcastTxCommit](
		ctx,
		c.requestTimeout,
		c.caller,
		broadcastTxCommitMethod,
		[]any{
			tx,
		},
	)
}

func (c *RPCClient) BroadcastTxAsync(ctx context.Context, tx types.Tx) (*mempool.ResultBroadcastTx, error) {
	return c.broadcastTX(ctx, broadcastTxAsyncMethod, tx)
}

func (c *RPCClient) BroadcastTxSync(ctx context.Context, tx types.Tx) (*mempool.ResultBroadcastTx, error) {
	return c.broadcastTX(ctx, broadcastTxSyncMethod, tx)
}

func (c *RPCClient) broadcastTX(ctx context.Context, route string, tx types.Tx) (*mempool.ResultBroadcastTx, error) {
	return sendRequestCommon[mempool.ResultBroadcastTx](
		ctx,
		c.requestTimeout,
		c.caller,
		route,
		[]any{
			tx,
		},
	)
}

func (c *RPCClient) UnconfirmedTxs(ctx context.Context, limit int) (*mempool.ResultUnconfirmedTxs, error) {
	return sendRequestCommon[mempool.ResultUnconfirmedTxs](
		ctx,
		c.requestTimeout,
		c.caller,
		unconfirmedTxsMethod,
		[]any{
			limit,
		},
	)
}

func (c *RPCClient) NumUnconfirmedTxs(ctx context.Context) (*mempool.ResultUnconfirmedTxs, error) {
	return sendRequestCommon[mempool.ResultUnconfirmedTxs](
		ctx,
		c.requestTimeout,
		c.caller,
		numUnconfirmedTxsMethod,
		nil,
	)
}

func (c *RPCClient) NetInfo(ctx context.Context) (*net.ResultNetInfo, error) {
	return sendRequestCommon[net.ResultNetInfo](
		ctx,
		c.requestTimeout,
		c.caller,
		netInfoMethod,
		nil,
	)
}

func (c *RPCClient) DumpConsensusState(ctx context.Context) (*consensus.ResultDumpConsensusState, error) {
	return sendRequestCommon[consensus.ResultDumpConsensusState](
		ctx,
		c.requestTimeout,
		c.caller,
		dumpConsensusStateMethod,
		nil,
	)
}

func (c *RPCClient) ConsensusState(ctx context.Context) (*consensus.ResultConsensusState, error) {
	return sendRequestCommon[consensus.ResultConsensusState](
		ctx,
		c.requestTimeout,
		c.caller,
		consensusStateMethod,
		nil,
	)
}

func (c *RPCClient) ConsensusParams(ctx context.Context, height *int64) (*consensus.ResultConsensusParams, error) {
	var v int64
	if height != nil {
		v = *height
	}

	return sendRequestCommon[consensus.ResultConsensusParams](
		ctx,
		c.requestTimeout,
		c.caller,
		consensusParamsMethod,
		[]any{
			v,
		},
	)
}

func (c *RPCClient) Health(ctx context.Context) (*health.ResultHealth, error) {
	return sendRequestCommon[health.ResultHealth](
		ctx,
		c.requestTimeout,
		c.caller,
		healthMethod,
		nil,
	)
}

func (c *RPCClient) BlockchainInfo(ctx context.Context, minHeight, maxHeight int64) (*blocks.ResultBlockchainInfo, error) {
	return sendRequestCommon[blocks.ResultBlockchainInfo](
		ctx,
		c.requestTimeout,
		c.caller,
		blockchainMethod,
		[]any{
			minHeight,
			maxHeight,
		},
	)
}

func (c *RPCClient) Genesis(ctx context.Context) (*net.ResultGenesis, error) {
	return sendRequestCommon[net.ResultGenesis](
		ctx,
		c.requestTimeout,
		c.caller,
		genesisMethod,
		nil,
	)
}

func (c *RPCClient) Block(ctx context.Context, height *int64) (*blocks.ResultBlock, error) {
	var v int64
	if height != nil {
		v = *height
	}

	return sendRequestCommon[blocks.ResultBlock](
		ctx,
		c.requestTimeout,
		c.caller,
		blockMethod,
		[]any{
			v,
		},
	)
}

func (c *RPCClient) BlockResults(ctx context.Context, height *int64) (*blocks.ResultBlockResults, error) {
	var v int64
	if height != nil {
		v = *height
	}

	return sendRequestCommon[blocks.ResultBlockResults](
		ctx,
		c.requestTimeout,
		c.caller,
		blockResultsMethod,
		[]any{
			v,
		},
	)
}

func (c *RPCClient) Commit(ctx context.Context, height *int64) (*blocks.ResultCommit, error) {
	var v int64
	if height != nil {
		v = *height
	}

	return sendRequestCommon[blocks.ResultCommit](
		ctx,
		c.requestTimeout,
		c.caller,
		commitMethod,
		[]any{
			v,
		},
	)
}

func (c *RPCClient) Tx(ctx context.Context, hash []byte) (*tx.ResultTx, error) {
	return sendRequestCommon[tx.ResultTx](
		ctx,
		c.requestTimeout,
		c.caller,
		txMethod,
		[]any{
			hash,
		},
	)
}

func (c *RPCClient) Validators(ctx context.Context, height *int64) (*consensus.ResultValidators, error) {
	var v int64
	if height != nil {
		v = *height
	}

	return sendRequestCommon[consensus.ResultValidators](
		ctx,
		c.requestTimeout,
		c.caller,
		validatorsMethod,
		[]any{
			v,
		},
	)
}

// newRequest creates a new request based on the method
// and given params
func newRequest(method string, params []any) *spec.BaseJSONRequest {
	return spec.NewJSONRequest(
		spec.JSONRPCStringID(xid.New().String()),
		method,
		params,
	)
}

// sendRequestCommon is the common request creation, sending, and parsing middleware
func sendRequestCommon[T any](
	ctx context.Context,
	timeout time.Duration,
	caller rpcclient.Client,
	method string,
	params []any,
) (*T, error) {
	// Prepare the RPC request
	request := newRequest(method, params)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Send the request with the provided context
	response, err := caller.SendRequest(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("unable to call RPC method %s, %w", method, err)
	}

	// Parse the response
	if response.Error != nil {
		return nil, response.Error
	}

	// Unmarshal the RPC response
	return unmarshalResponseBytes[T](response.Result)
}

// unmarshalResponseBytes Amino JSON-unmarshals the RPC response data
func unmarshalResponseBytes[T any](responseBytes []byte) (*T, error) {
	var result T

	// Amino JSON-unmarshal the RPC response data
	if err := amino.UnmarshalJSON(responseBytes, &result); err != nil {
		return nil, fmt.Errorf("unable to unmarshal response bytes, %w", err)
	}

	return &result, nil
}
