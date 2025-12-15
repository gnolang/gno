package rpcclient

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
)

// Client is the JSON-RPC client abstraction
type Client interface {
	// SendRequest sends a single RPC request to the JSON-RPC layer
	SendRequest(context.Context, *spec.BaseJSONRequest) (*spec.BaseJSONResponse, error)

	// SendBatch sends a batch of RPC requests to the JSON-RPC layer
	SendBatch(context.Context, spec.BaseJSONRequests) (spec.BaseJSONResponses, error)

	// Close closes the RPC client
	Close() error
}

// Batch is the JSON-RPC batch abstraction
type Batch interface {
	// AddRequest adds a single request to the RPC batch
	AddRequest(*spec.BaseJSONRequest)

	// Send sends the batch to the RPC layer
	Send(context.Context) (spec.BaseJSONResponses, error)

	// Clear clears out the batch
	Clear() int

	// Count returns the number of enqueued requests
	Count() int
}
