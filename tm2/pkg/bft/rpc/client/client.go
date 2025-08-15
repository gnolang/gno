package client

import (
	"context"
	"fmt"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client/batch"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client/http"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client/ws"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
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

func (c *RPCClient) Status() (*ctypes.ResultStatus, error) {
	return c.StatusWithContext(context.Background())
}

func (c *RPCClient) StatusWithContext(ctx context.Context) (*ctypes.ResultStatus, error) {
	return sendRequestCommon[ctypes.ResultStatus](
		ctx,
		c.requestTimeout,
		c.caller,
		statusMethod,
		map[string]any{},
	)
}

func (c *RPCClient) ABCIInfo() (*ctypes.ResultABCIInfo, error) {
	return c.ABCIInfoWithContext(context.Background())
}

func (c *RPCClient) ABCIInfoWithContext(ctx context.Context) (*ctypes.ResultABCIInfo, error) {
	return sendRequestCommon[ctypes.ResultABCIInfo](
		ctx,
		c.requestTimeout,
		c.caller,
		abciInfoMethod,
		map[string]any{},
	)
}

func (c *RPCClient) ABCIQuery(path string, data []byte) (*ctypes.ResultABCIQuery, error) {
	return c.ABCIQueryWithContext(context.Background(), path, data)
}

func (c *RPCClient) ABCIQueryWithContext(ctx context.Context, path string, data []byte) (*ctypes.ResultABCIQuery, error) {
	return c.ABCIQueryWithOptionsWithContext(ctx, path, data, DefaultABCIQueryOptions)
}

func (c *RPCClient) ABCIQueryWithOptions(path string, data []byte, opts ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	return c.ABCIQueryWithOptionsWithContext(context.Background(), path, data, opts)
}

func (c *RPCClient) ABCIQueryWithOptionsWithContext(ctx context.Context, path string, data []byte, opts ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	return sendRequestCommon[ctypes.ResultABCIQuery](
		ctx,
		c.requestTimeout,
		c.caller,
		abciQueryMethod,
		map[string]any{
			"path":   path,
			"data":   data,
			"height": opts.Height,
			"prove":  opts.Prove,
		},
	)
}

func (c *RPCClient) BroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	return c.BroadcastTxCommitWithContext(context.Background(), tx)
}

func (c *RPCClient) BroadcastTxCommitWithContext(ctx context.Context, tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	return sendRequestCommon[ctypes.ResultBroadcastTxCommit](
		ctx,
		c.requestTimeout,
		c.caller,
		broadcastTxCommitMethod,
		map[string]any{"tx": tx},
	)
}

func (c *RPCClient) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return c.BroadcastTxAsyncWithContext(context.Background(), tx)
}

func (c *RPCClient) BroadcastTxAsyncWithContext(ctx context.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return c.broadcastTXWithContext(ctx, broadcastTxAsyncMethod, tx)
}

func (c *RPCClient) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return c.BroadcastTxSyncWithContext(context.Background(), tx)
}

func (c *RPCClient) BroadcastTxSyncWithContext(ctx context.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return c.broadcastTXWithContext(ctx, broadcastTxSyncMethod, tx)
}

func (c *RPCClient) broadcastTXWithContext(ctx context.Context, route string, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return sendRequestCommon[ctypes.ResultBroadcastTx](
		ctx,
		c.requestTimeout,
		c.caller,
		route,
		map[string]any{"tx": tx},
	)
}

func (c *RPCClient) UnconfirmedTxs(limit int) (*ctypes.ResultUnconfirmedTxs, error) {
	return c.UnconfirmedTxsWithContext(context.Background(), limit)
}

func (c *RPCClient) UnconfirmedTxsWithContext(ctx context.Context, limit int) (*ctypes.ResultUnconfirmedTxs, error) {
	return sendRequestCommon[ctypes.ResultUnconfirmedTxs](
		ctx,
		c.requestTimeout,
		c.caller,
		unconfirmedTxsMethod,
		map[string]any{"limit": limit},
	)
}

func (c *RPCClient) NumUnconfirmedTxs() (*ctypes.ResultUnconfirmedTxs, error) {
	return c.NumUnconfirmedTxsWithContext(context.Background())
}

func (c *RPCClient) NumUnconfirmedTxsWithContext(ctx context.Context) (*ctypes.ResultUnconfirmedTxs, error) {
	return sendRequestCommon[ctypes.ResultUnconfirmedTxs](
		ctx,
		c.requestTimeout,
		c.caller,
		numUnconfirmedTxsMethod,
		map[string]any{},
	)
}

func (c *RPCClient) NetInfo() (*ctypes.ResultNetInfo, error) {
	return c.NetInfoWithContext(context.Background())
}

func (c *RPCClient) NetInfoWithContext(ctx context.Context) (*ctypes.ResultNetInfo, error) {
	return sendRequestCommon[ctypes.ResultNetInfo](
		ctx,
		c.requestTimeout,
		c.caller,
		netInfoMethod,
		map[string]any{},
	)
}

func (c *RPCClient) DumpConsensusState() (*ctypes.ResultDumpConsensusState, error) {
	return c.DumpConsensusStateWithContext(context.Background())
}

func (c *RPCClient) DumpConsensusStateWithContext(ctx context.Context) (*ctypes.ResultDumpConsensusState, error) {
	return sendRequestCommon[ctypes.ResultDumpConsensusState](
		ctx,
		c.requestTimeout,
		c.caller,
		dumpConsensusStateMethod,
		map[string]any{},
	)
}

func (c *RPCClient) ConsensusState() (*ctypes.ResultConsensusState, error) {
	return c.ConsensusStateWithContext(context.Background())
}

func (c *RPCClient) ConsensusStateWithContext(ctx context.Context) (*ctypes.ResultConsensusState, error) {
	return sendRequestCommon[ctypes.ResultConsensusState](
		ctx,
		c.requestTimeout,
		c.caller,
		consensusStateMethod,
		map[string]any{},
	)
}

func (c *RPCClient) ConsensusParams(height *int64) (*ctypes.ResultConsensusParams, error) {
	return c.ConsensusParamsWithContext(context.Background(), height)
}

func (c *RPCClient) ConsensusParamsWithContext(ctx context.Context, height *int64) (*ctypes.ResultConsensusParams, error) {
	params := map[string]any{}
	if height != nil {
		params["height"] = height
	}

	return sendRequestCommon[ctypes.ResultConsensusParams](
		ctx,
		c.requestTimeout,
		c.caller,
		consensusParamsMethod,
		params,
	)
}

func (c *RPCClient) Health() (*ctypes.ResultHealth, error) {
	return c.HealthWithContext(context.Background())
}

func (c *RPCClient) HealthWithContext(ctx context.Context) (*ctypes.ResultHealth, error) {
	return sendRequestCommon[ctypes.ResultHealth](
		ctx,
		c.requestTimeout,
		c.caller,
		healthMethod,
		map[string]any{},
	)
}

func (c *RPCClient) BlockchainInfo(minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error) {
	return c.BlockchainInfoWithContext(context.Background(), minHeight, maxHeight)
}

func (c *RPCClient) BlockchainInfoWithContext(ctx context.Context, minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error) {
	return sendRequestCommon[ctypes.ResultBlockchainInfo](
		ctx,
		c.requestTimeout,
		c.caller,
		blockchainMethod,
		map[string]any{
			"minHeight": minHeight,
			"maxHeight": maxHeight,
		},
	)
}

func (c *RPCClient) Genesis() (*ctypes.ResultGenesis, error) {
	return c.GenesisWithContext(context.Background())
}

func (c *RPCClient) GenesisWithContext(ctx context.Context) (*ctypes.ResultGenesis, error) {
	return sendRequestCommon[ctypes.ResultGenesis](
		ctx,
		c.requestTimeout,
		c.caller,
		genesisMethod,
		map[string]any{},
	)
}

func (c *RPCClient) Block(height *int64) (*ctypes.ResultBlock, error) {
	return c.BlockWithContext(context.Background(), height)
}

func (c *RPCClient) BlockWithContext(ctx context.Context, height *int64) (*ctypes.ResultBlock, error) {
	params := map[string]any{}
	if height != nil {
		params["height"] = height
	}

	return sendRequestCommon[ctypes.ResultBlock](
		ctx,
		c.requestTimeout,
		c.caller,
		blockMethod,
		params,
	)
}

func (c *RPCClient) BlockResults(height *int64) (*ctypes.ResultBlockResults, error) {
	return c.BlockResultsWithContext(context.Background(), height)
}

func (c *RPCClient) BlockResultsWithContext(ctx context.Context, height *int64) (*ctypes.ResultBlockResults, error) {
	params := map[string]any{}
	if height != nil {
		params["height"] = height
	}

	return sendRequestCommon[ctypes.ResultBlockResults](
		ctx,
		c.requestTimeout,
		c.caller,
		blockResultsMethod,
		params,
	)
}

func (c *RPCClient) Commit(height *int64) (*ctypes.ResultCommit, error) {
	return c.CommitWithContext(context.Background(), height)
}

func (c *RPCClient) CommitWithContext(ctx context.Context, height *int64) (*ctypes.ResultCommit, error) {
	params := map[string]any{}
	if height != nil {
		params["height"] = height
	}

	return sendRequestCommon[ctypes.ResultCommit](
		ctx,
		c.requestTimeout,
		c.caller,
		commitMethod,
		params,
	)
}

func (c *RPCClient) Tx(hash []byte) (*ctypes.ResultTx, error) {
	return c.TxWithContext(context.Background(), hash)
}

func (c *RPCClient) TxWithContext(ctx context.Context, hash []byte) (*ctypes.ResultTx, error) {
	return sendRequestCommon[ctypes.ResultTx](
		ctx,
		c.requestTimeout,
		c.caller,
		txMethod,
		map[string]any{
			"hash": hash,
		},
	)
}

func (c *RPCClient) Validators(height *int64) (*ctypes.ResultValidators, error) {
	return c.ValidatorsWithContext(context.Background(), height)
}

func (c *RPCClient) ValidatorsWithContext(ctx context.Context, height *int64) (*ctypes.ResultValidators, error) {
	params := map[string]any{}
	if height != nil {
		params["height"] = height
	}

	return sendRequestCommon[ctypes.ResultValidators](
		ctx,
		c.requestTimeout,
		c.caller,
		validatorsMethod,
		params,
	)
}

// newRequest creates a new request based on the method
// and given params
func newRequest(method string, params map[string]any) (rpctypes.RPCRequest, error) {
	id := rpctypes.JSONRPCStringID(xid.New().String())

	return rpctypes.MapToRequest(id, method, params)
}

// sendRequestCommon is the common request creation, sending, and parsing middleware
func sendRequestCommon[T any](
	ctx context.Context,
	timeout time.Duration,
	caller rpcclient.Client,
	method string,
	params map[string]any,
) (*T, error) {
	// Prepare the RPC request
	request, err := newRequest(method, params)
	if err != nil {
		return nil, err
	}

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
