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
func NewHTTPClient(rpcURL string) (*RPCClient, error) {
	httpClient, err := http.NewClient(rpcURL)
	if err != nil {
		return nil, err
	}

	return NewRPCClient(httpClient), nil
}

// NewWSClient takes a remote endpoint in the form <protocol>://<host>:<port>,
// and returns a WS client that communicates with a Tendermint node over
// WS connection.
//
// Request batching is available for JSON RPC requests over WS, which conforms to
// the JSON RPC specification (https://www.jsonrpc.org/specification#batch). See
// the example for more details
func NewWSClient(rpcURL string) (*RPCClient, error) {
	wsClient, err := ws.NewClient(rpcURL)
	if err != nil {
		return nil, err
	}

	return NewRPCClient(wsClient), nil
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
	return sendRequestCommon[ctypes.ResultStatus](
		c.caller,
		c.requestTimeout,
		statusMethod,
		map[string]any{},
	)
}

func (c *RPCClient) ABCIInfo() (*ctypes.ResultABCIInfo, error) {
	return sendRequestCommon[ctypes.ResultABCIInfo](
		c.caller,
		c.requestTimeout,
		abciInfoMethod,
		map[string]any{},
	)
}

func (c *RPCClient) ABCIQuery(path string, data []byte) (*ctypes.ResultABCIQuery, error) {
	return c.ABCIQueryWithOptions(path, data, DefaultABCIQueryOptions)
}

func (c *RPCClient) ABCIQueryWithOptions(path string, data []byte, opts ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	return sendRequestCommon[ctypes.ResultABCIQuery](
		c.caller,
		c.requestTimeout,
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
	return sendRequestCommon[ctypes.ResultBroadcastTxCommit](
		c.caller,
		c.requestTimeout,
		broadcastTxCommitMethod,
		map[string]any{"tx": tx},
	)
}

func (c *RPCClient) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return c.broadcastTX(broadcastTxAsyncMethod, tx)
}

func (c *RPCClient) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return c.broadcastTX(broadcastTxSyncMethod, tx)
}

func (c *RPCClient) broadcastTX(route string, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return sendRequestCommon[ctypes.ResultBroadcastTx](
		c.caller,
		c.requestTimeout,
		route,
		map[string]any{"tx": tx},
	)
}

func (c *RPCClient) UnconfirmedTxs(limit int) (*ctypes.ResultUnconfirmedTxs, error) {
	return sendRequestCommon[ctypes.ResultUnconfirmedTxs](
		c.caller,
		c.requestTimeout,
		unconfirmedTxsMethod,
		map[string]any{"limit": limit},
	)
}

func (c *RPCClient) NumUnconfirmedTxs() (*ctypes.ResultUnconfirmedTxs, error) {
	return sendRequestCommon[ctypes.ResultUnconfirmedTxs](
		c.caller,
		c.requestTimeout,
		numUnconfirmedTxsMethod,
		map[string]any{},
	)
}

func (c *RPCClient) NetInfo() (*ctypes.ResultNetInfo, error) {
	return sendRequestCommon[ctypes.ResultNetInfo](
		c.caller,
		c.requestTimeout,
		netInfoMethod,
		map[string]any{},
	)
}

func (c *RPCClient) DumpConsensusState() (*ctypes.ResultDumpConsensusState, error) {
	return sendRequestCommon[ctypes.ResultDumpConsensusState](
		c.caller,
		c.requestTimeout,
		dumpConsensusStateMethod,
		map[string]any{},
	)
}

func (c *RPCClient) ConsensusState() (*ctypes.ResultConsensusState, error) {
	return sendRequestCommon[ctypes.ResultConsensusState](
		c.caller,
		c.requestTimeout,
		consensusStateMethod,
		map[string]any{},
	)
}

func (c *RPCClient) ConsensusParams(height *int64) (*ctypes.ResultConsensusParams, error) {
	params := map[string]any{}
	if height != nil {
		params["height"] = height
	}

	return sendRequestCommon[ctypes.ResultConsensusParams](
		c.caller,
		c.requestTimeout,
		consensusParamsMethod,
		params,
	)
}

func (c *RPCClient) Health() (*ctypes.ResultHealth, error) {
	return sendRequestCommon[ctypes.ResultHealth](
		c.caller,
		c.requestTimeout,
		healthMethod,
		map[string]any{},
	)
}

func (c *RPCClient) BlockchainInfo(minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error) {
	return sendRequestCommon[ctypes.ResultBlockchainInfo](
		c.caller,
		c.requestTimeout,
		blockchainMethod,
		map[string]any{
			"minHeight": minHeight,
			"maxHeight": maxHeight,
		},
	)
}

func (c *RPCClient) Genesis() (*ctypes.ResultGenesis, error) {
	return sendRequestCommon[ctypes.ResultGenesis](
		c.caller,
		c.requestTimeout,
		genesisMethod,
		map[string]any{},
	)
}

func (c *RPCClient) Block(height *int64) (*ctypes.ResultBlock, error) {
	params := map[string]any{}
	if height != nil {
		params["height"] = height
	}

	return sendRequestCommon[ctypes.ResultBlock](
		c.caller,
		c.requestTimeout,
		blockMethod,
		params,
	)
}

func (c *RPCClient) BlockResults(height *int64) (*ctypes.ResultBlockResults, error) {
	params := map[string]any{}
	if height != nil {
		params["height"] = height
	}

	return sendRequestCommon[ctypes.ResultBlockResults](
		c.caller,
		c.requestTimeout,
		blockResultsMethod,
		params,
	)
}

func (c *RPCClient) Commit(height *int64) (*ctypes.ResultCommit, error) {
	params := map[string]any{}
	if height != nil {
		params["height"] = height
	}

	return sendRequestCommon[ctypes.ResultCommit](
		c.caller,
		c.requestTimeout,
		commitMethod,
		params,
	)
}

func (c *RPCClient) Tx(hash []byte) (*ctypes.ResultTx, error) {
	return sendRequestCommon[ctypes.ResultTx](
		c.caller,
		c.requestTimeout,
		txMethod,
		map[string]interface{}{
			"hash": hash,
		},
	)
}

func (c *RPCClient) Validators(height *int64) (*ctypes.ResultValidators, error) {
	params := map[string]any{}
	if height != nil {
		params["height"] = height
	}

	return sendRequestCommon[ctypes.ResultValidators](
		c.caller,
		c.requestTimeout,
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
	caller rpcclient.Client,
	timeout time.Duration,
	method string,
	params map[string]any,
) (*T, error) {
	// Prepare the RPC request
	request, err := newRequest(method, params)
	if err != nil {
		return nil, err
	}

	// Send the request
	ctx, cancelFn := context.WithTimeout(context.Background(), timeout)
	defer cancelFn()

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
