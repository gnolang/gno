package client

import (
	"net/http"

	ctypes "github.com/gnolang/gno/pkgs/bft/rpc/core/types"
	rpcclient "github.com/gnolang/gno/pkgs/bft/rpc/lib/client"
	"github.com/gnolang/gno/pkgs/bft/types"
	"github.com/gnolang/gno/pkgs/errors"
)

/*
HTTP is a Client implementation that communicates with a Tendermint node over
JSON RPC and WebSockets.

This is the main implementation you probably want to use in production code.
There are other implementations when calling the Tendermint node in-process
(Local), or when you want to mock out the server for test code (mock).

Request batching is available for JSON RPC requests over HTTP, which conforms to
the JSON RPC specification (https://www.jsonrpc.org/specification#batch). See
the example for more details.
*/
type HTTP struct {
	remote string
	rpc    *rpcclient.JSONRPCClient

	*baseRPCClient
}

// BatchHTTP provides the same interface as `HTTP`, but allows for batching of
// requests (as per https://www.jsonrpc.org/specification#batch). Do not
// instantiate directly - rather use the HTTP.NewBatch() method to create an
// instance of this struct.
//
// Batching of HTTP requests is thread-safe in the sense that multiple
// goroutines can each create their own batches and send them using the same
// HTTP client. Multiple goroutines could also enqueue transactions in a single
// batch, but ordering of transactions in the batch cannot be guaranteed in such
// an example.
type BatchHTTP struct {
	rpcBatch *rpcclient.JSONRPCRequestBatch
	*baseRPCClient
}

// rpcClient is an internal interface to which our RPC clients (batch and
// non-batch) must conform. Acts as an additional code-level sanity check to
// make sure the implementations stay coherent.
type rpcClient interface {
	ABCIClient
	HistoryClient
	NetworkClient
	SignClient
	StatusClient
	MempoolClient
}

// baseRPCClient implements the basic RPC method logic without the actual
// underlying RPC call functionality, which is provided by `caller`.
type baseRPCClient struct {
	caller rpcclient.JSONRPCCaller
}

var (
	_ rpcClient = (*HTTP)(nil)
	_ rpcClient = (*BatchHTTP)(nil)
	_ rpcClient = (*baseRPCClient)(nil)
)

//-----------------------------------------------------------------------------
// HTTP

// NewHTTP takes a remote endpoint in the form <protocol>://<host>:<port> and
// the websocket path (which always seems to be "/websocket")
// The function panics if the provided remote is invalid.<Paste>
func NewHTTP(remote, wsEndpoint string) *HTTP {
	httpClient := rpcclient.DefaultHTTPClient(remote)

	return NewHTTPWithClient(remote, wsEndpoint, httpClient)
}

// NewHTTPWithClient allows for setting a custom http client. See NewHTTP
// The function panics if the provided client is nil or remote is invalid.
func NewHTTPWithClient(remote, wsEndpoint string, client *http.Client) *HTTP {
	if client == nil {
		panic("nil http.Client provided")
	}
	rc := rpcclient.NewJSONRPCClientWithHTTPClient(remote, client)

	return &HTTP{
		rpc:           rc,
		remote:        remote,
		baseRPCClient: &baseRPCClient{caller: rc},
	}
}

var _ Client = (*HTTP)(nil)

// NewBatch creates a new batch client for this HTTP client.
func (c *HTTP) NewBatch() *BatchHTTP {
	rpcBatch := c.rpc.NewRequestBatch()

	return &BatchHTTP{
		rpcBatch: rpcBatch,
		baseRPCClient: &baseRPCClient{
			caller: rpcBatch,
		},
	}
}

//-----------------------------------------------------------------------------
// BatchHTTP

// Send is a convenience function for an HTTP batch that will trigger the
// compilation of the batched requests and send them off using the client as a
// single request. On success, this returns a list of the deserialized results
// from each request in the sent batch.
func (b *BatchHTTP) Send() ([]interface{}, error) {
	return b.rpcBatch.Send()
}

// Clear will empty out this batch of requests and return the number of requests
// that were cleared out.
func (b *BatchHTTP) Clear() int {
	return b.rpcBatch.Clear()
}

// Count returns the number of enqueued requests waiting to be sent.
func (b *BatchHTTP) Count() int {
	return b.rpcBatch.Count()
}

//-----------------------------------------------------------------------------
// baseRPCClient

func (c *baseRPCClient) Status() (*ctypes.ResultStatus, error) {
	result := new(ctypes.ResultStatus)
	_, err := c.caller.Call("status", map[string]interface{}{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "Status")
	}
	return result, nil
}

func (c *baseRPCClient) ABCIInfo() (*ctypes.ResultABCIInfo, error) {
	result := new(ctypes.ResultABCIInfo)
	_, err := c.caller.Call("abci_info", map[string]interface{}{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "ABCIInfo")
	}
	return result, nil
}

func (c *baseRPCClient) ABCIQuery(path string, data []byte) (*ctypes.ResultABCIQuery, error) {
	return c.ABCIQueryWithOptions(path, data, DefaultABCIQueryOptions)
}

func (c *baseRPCClient) ABCIQueryWithOptions(path string, data []byte, opts ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	result := new(ctypes.ResultABCIQuery)
	_, err := c.caller.Call("abci_query",
		map[string]interface{}{"path": path, "data": data, "height": opts.Height, "prove": opts.Prove},
		result)
	if err != nil {
		return nil, errors.Wrap(err, "ABCIQuery")
	}
	return result, nil
}

func (c *baseRPCClient) BroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	result := new(ctypes.ResultBroadcastTxCommit)
	_, err := c.caller.Call("broadcast_tx_commit", map[string]interface{}{"tx": tx}, result)
	if err != nil {
		return nil, errors.Wrap(err, "broadcast_tx_commit")
	}
	return result, nil
}

func (c *baseRPCClient) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return c.broadcastTX("broadcast_tx_async", tx)
}

func (c *baseRPCClient) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return c.broadcastTX("broadcast_tx_sync", tx)
}

func (c *baseRPCClient) broadcastTX(route string, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	result := new(ctypes.ResultBroadcastTx)
	_, err := c.caller.Call(route, map[string]interface{}{"tx": tx}, result)
	if err != nil {
		return nil, errors.Wrap(err, route)
	}
	return result, nil
}

func (c *baseRPCClient) UnconfirmedTxs(limit int) (*ctypes.ResultUnconfirmedTxs, error) {
	result := new(ctypes.ResultUnconfirmedTxs)
	_, err := c.caller.Call("unconfirmed_txs", map[string]interface{}{"limit": limit}, result)
	if err != nil {
		return nil, errors.Wrap(err, "unconfirmed_txs")
	}
	return result, nil
}

func (c *baseRPCClient) NumUnconfirmedTxs() (*ctypes.ResultUnconfirmedTxs, error) {
	result := new(ctypes.ResultUnconfirmedTxs)
	_, err := c.caller.Call("num_unconfirmed_txs", map[string]interface{}{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "num_unconfirmed_txs")
	}
	return result, nil
}

func (c *baseRPCClient) NetInfo() (*ctypes.ResultNetInfo, error) {
	result := new(ctypes.ResultNetInfo)
	_, err := c.caller.Call("net_info", map[string]interface{}{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "NetInfo")
	}
	return result, nil
}

func (c *baseRPCClient) DumpConsensusState() (*ctypes.ResultDumpConsensusState, error) {
	result := new(ctypes.ResultDumpConsensusState)
	_, err := c.caller.Call("dump_consensus_state", map[string]interface{}{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "DumpConsensusState")
	}
	return result, nil
}

func (c *baseRPCClient) ConsensusState() (*ctypes.ResultConsensusState, error) {
	result := new(ctypes.ResultConsensusState)
	_, err := c.caller.Call("consensus_state", map[string]interface{}{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "ConsensusState")
	}
	return result, nil
}

func (c *baseRPCClient) Health() (*ctypes.ResultHealth, error) {
	result := new(ctypes.ResultHealth)
	_, err := c.caller.Call("health", map[string]interface{}{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "Health")
	}
	return result, nil
}

func (c *baseRPCClient) BlockchainInfo(minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error) {
	result := new(ctypes.ResultBlockchainInfo)
	_, err := c.caller.Call("blockchain",
		map[string]interface{}{"minHeight": minHeight, "maxHeight": maxHeight},
		result)
	if err != nil {
		return nil, errors.Wrap(err, "BlockchainInfo")
	}
	return result, nil
}

func (c *baseRPCClient) Genesis() (*ctypes.ResultGenesis, error) {
	result := new(ctypes.ResultGenesis)
	_, err := c.caller.Call("genesis", map[string]interface{}{}, result)
	if err != nil {
		return nil, errors.Wrap(err, "Genesis")
	}
	return result, nil
}

func (c *baseRPCClient) Block(height *int64) (*ctypes.ResultBlock, error) {
	result := new(ctypes.ResultBlock)
	_, err := c.caller.Call("block", map[string]interface{}{"height": height}, result)
	if err != nil {
		return nil, errors.Wrap(err, "Block")
	}
	return result, nil
}

func (c *baseRPCClient) BlockResults(height *int64) (*ctypes.ResultBlockResults, error) {
	result := new(ctypes.ResultBlockResults)
	_, err := c.caller.Call("block_results", map[string]interface{}{"height": height}, result)
	if err != nil {
		return nil, errors.Wrap(err, "Block Result")
	}
	return result, nil
}

func (c *baseRPCClient) Commit(height *int64) (*ctypes.ResultCommit, error) {
	result := new(ctypes.ResultCommit)
	_, err := c.caller.Call("commit", map[string]interface{}{"height": height}, result)
	if err != nil {
		return nil, errors.Wrap(err, "Commit")
	}
	return result, nil
}

func (c *baseRPCClient) Tx(hash []byte, prove bool) (*ctypes.ResultTx, error) {
	result := new(ctypes.ResultTx)
	params := map[string]interface{}{
		"hash":  hash,
		"prove": prove,
	}
	_, err := c.caller.Call("tx", params, result)
	if err != nil {
		return nil, errors.Wrap(err, "Tx")
	}
	return result, nil
}

func (c *baseRPCClient) TxSearch(query string, prove bool, page, perPage int) (*ctypes.ResultTxSearch, error) {
	result := new(ctypes.ResultTxSearch)
	params := map[string]interface{}{
		"query":    query,
		"prove":    prove,
		"page":     page,
		"per_page": perPage,
	}
	_, err := c.caller.Call("tx_search", params, result)
	if err != nil {
		return nil, errors.Wrap(err, "TxSearch")
	}
	return result, nil
}

func (c *baseRPCClient) Validators(height *int64) (*ctypes.ResultValidators, error) {
	result := new(ctypes.ResultValidators)
	params := map[string]interface{}{}
	if height != nil {
		params["height"] = height
	}
	_, err := c.caller.Call("validators", params, result)
	if err != nil {
		return nil, errors.Wrap(err, "Validators")
	}
	return result, nil
}
