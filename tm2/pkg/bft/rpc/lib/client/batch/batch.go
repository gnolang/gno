package batch

import (
	"context"
	"sync"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
)

type Client interface {
	SendBatch(context.Context, types.RPCRequests) (types.RPCResponses, error)
}

// Batch allows us to buffer multiple request/response structures
// into a single batch request. Note that this batch acts like a FIFO queue, and
// is thread-safe
type Batch struct {
	sync.RWMutex

	client   Client
	requests types.RPCRequests
}

// NewBatch creates a new batch object
func NewBatch(client Client) *Batch {
	return &Batch{
		client:   client,
		requests: make(types.RPCRequests, 0),
	}
}

// Count returns the number of enqueued requests waiting to be sent
func (b *Batch) Count() int {
	b.RLock()
	defer b.RUnlock()

	return len(b.requests)
}

// Clear empties out the request batch
func (b *Batch) Clear() int {
	b.Lock()
	defer b.Unlock()

	return b.clear()
}

func (b *Batch) clear() int {
	count := len(b.requests)
	b.requests = make(types.RPCRequests, 0)

	return count
}

// Send will attempt to send the current batch of enqueued requests, and then
// will clear out the requests once done
func (b *Batch) Send(ctx context.Context) (types.RPCResponses, error) {
	b.Lock()
	defer func() {
		b.clear()
		b.Unlock()
	}()

	responses, err := b.client.SendBatch(ctx, b.requests)
	if err != nil {
		return nil, err
	}

	return responses, nil
}

// AddRequest adds a new request onto the batch
func (b *Batch) AddRequest(request types.RPCRequest) {
	b.Lock()
	defer b.Unlock()

	b.requests = append(b.requests, request)
}
