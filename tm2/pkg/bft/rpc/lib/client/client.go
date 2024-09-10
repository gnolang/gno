package rpcclient

import (
	"context"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
)

// Client is the JSON-RPC client abstraction
type Client interface {
	// SendRequest sends a single RPC request to the JSON-RPC layer
	SendRequest(context.Context, types.RPCRequest) (*types.RPCResponse, error)

	// SendBatch sends a batch of RPC requests to the JSON-RPC layer
	SendBatch(context.Context, types.RPCRequests) (types.RPCResponses, error)

	// Close closes the RPC client
	Close() error
}

// Batch is the JSON-RPC batch abstraction
type Batch interface {
	// AddRequest adds a single request to the RPC batch
	AddRequest(types.RPCRequest)

	// Send sends the batch to the RPC layer
	Send(context.Context) (types.RPCResponses, error)

	// Clear clears out the batch
	Clear() int

	// Count returns the number of enqueued requests
	Count() int
}
