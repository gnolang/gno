package batch

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
)

type Client interface {
	SendBatch(context.Context, spec.BaseJSONRequests) (spec.BaseJSONResponses, error)
}

// Batch allows us to buffer multiple request/response structures
// into a single batch request.
// NOT thread safe
type Batch struct {
	client   Client
	requests spec.BaseJSONRequests
}

// NewBatch creates a new batch object
func NewBatch(client Client) *Batch {
	return &Batch{
		client:   client,
		requests: make(spec.BaseJSONRequests, 0),
	}
}

// Count returns the number of enqueued requests waiting to be sent
func (b *Batch) Count() int {
	return len(b.requests)
}

// Clear empties out the request batch
func (b *Batch) Clear() int {
	return b.clear()
}

func (b *Batch) clear() int {
	count := len(b.requests)
	b.requests = make(spec.BaseJSONRequests, 0)

	return count
}

// Send will attempt to send the current batch of enqueued requests, and then
// will clear out the requests once done
func (b *Batch) Send(ctx context.Context) (spec.BaseJSONResponses, error) {
	defer func() {
		b.clear()
	}()

	responses, err := b.client.SendBatch(ctx, b.requests)
	if err != nil {
		return nil, err
	}

	return responses, nil
}

// AddRequest adds a new request onto the batch
func (b *Batch) AddRequest(request *spec.BaseJSONRequest) {
	b.requests = append(b.requests, request)
}
