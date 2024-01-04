package client

import (
	"net/http"

	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client"
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

// baseRPCClient implements the basic RPC method logic without the actual
// underlying RPC call functionality, which is provided by `caller`.
type baseRPCClient struct {
	caller rpcclient.RPCCaller
}

var (
	_ Client = (*HTTP)(nil)
	_ Client = (*BatchHTTP)(nil)
)

// -----------------------------------------------------------------------------
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

// -----------------------------------------------------------------------------
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

// -----------------------------------------------------------------------------
// baseRPCClient
