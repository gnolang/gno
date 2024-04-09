package rpcclient

import (
	"context"
	"fmt"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client/http"
	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/random"
)

type BatchClient interface {
	SendBatch(ctx context.Context, requests http.WrappedRPCRequests) (types.RPCResponses, error)
	GetIDPrefix() types.JSONRPCID
}

var _ http.RPCCaller = (*RPCRequestBatch)(nil)

// RPCRequestBatch allows us to buffer multiple request/response structures
// into a single batch request. Note that this batch acts like a FIFO queue, and
// is thread-safe
type RPCRequestBatch struct {
	sync.Mutex

	client   BatchClient
	requests http.WrappedRPCRequests
}

// NewRPCRequestBatch creates a new
func NewRPCRequestBatch(client BatchClient) *RPCRequestBatch {
	return &RPCRequestBatch{
		client:   client,
		requests: make(http.WrappedRPCRequests, 0),
	}
}

// Count returns the number of enqueued requests waiting to be sent
func (b *RPCRequestBatch) Count() int {
	b.Lock()
	defer b.Unlock()

	return len(b.requests)
}

// Clear empties out the request batch
func (b *RPCRequestBatch) Clear() int {
	b.Lock()
	defer b.Unlock()

	return b.clear()
}

func (b *RPCRequestBatch) clear() int {
	count := len(b.requests)
	b.requests = make(http.WrappedRPCRequests, 0)

	return count
}

// Send will attempt to send the current batch of enqueued requests, and then
// will clear out the requests once done. On success, this returns the
// deserialized list of results from each of the enqueued requests
func (b *RPCRequestBatch) Send(ctx context.Context) ([]any, error) {
	b.Lock()
	defer func() {
		b.clear()
		b.Unlock()
	}()

	requests := make(types.RPCRequests, 0, len(b.requests))
	results := make([]any, 0, len(b.requests))

	for _, req := range b.requests {
		requests = append(requests, req.request)
		results = append(results, req.result)
	}

	responses, err := b.client.SendBatch(ctx, b.requests)
	if err != nil {
		return nil, err
	}

	if err := http.unmarshalResponsesIntoResults(requests, responses, results); err != nil {
		return nil, err
	}

	return results, nil
}

// Call enqueues a request to call the given RPC method with the specified parameters
func (b *RPCRequestBatch) Call(method string, params map[string]any, result any) error {
	// Assuming this is sufficiently random, there shouldn't be any problems.
	// However, using uuid for any kind of ID generation is always preferred
	id := types.JSONRPCStringID(
		fmt.Sprintf("%s-%s", b.client.GetIDPrefix(), random.RandStr(8)),
	)

	request, err := types.MapToRequest(id, method, params)
	if err != nil {
		return err
	}

	b.enqueue(
		&http.WrappedRPCRequest{
			request: request,
			result:  result,
		},
	)

	return nil
}

func (b *RPCRequestBatch) enqueue(req *http.WrappedRPCRequest) {
	b.Lock()
	defer b.Unlock()

	b.requests = append(b.requests, req)
}
