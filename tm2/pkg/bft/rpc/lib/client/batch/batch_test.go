package batch

import (
	"context"
	"testing"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateRequests generates dummy RPC requests
func generateRequests(t *testing.T, count int) types.RPCRequests {
	t.Helper()

	requests := make(types.RPCRequests, 0, count)

	for i := 0; i < count; i++ {
		requests = append(requests, types.RPCRequest{
			JSONRPC: "2.0",
			ID:      types.JSONRPCIntID(i),
		})
	}

	return requests
}

func TestBatch_AddRequest(t *testing.T) {
	t.Parallel()

	var (
		capturedSend types.RPCRequests
		requests     = generateRequests(t, 100)

		mockClient = &mockClient{
			sendBatchFn: func(_ context.Context, requests types.RPCRequests) (types.RPCResponses, error) {
				capturedSend = requests

				responses := make(types.RPCResponses, len(requests))

				for index, request := range requests {
					responses[index] = types.RPCResponse{
						JSONRPC: "2.0",
						ID:      request.ID,
					}
				}

				return responses, nil
			},
		}
	)

	// Create the batch
	b := NewBatch(mockClient)

	// Add the requests
	for _, request := range requests {
		b.AddRequest(request)
	}

	// Make sure the count is correct
	require.Equal(t, len(requests), b.Count())

	// Send the requests
	responses, err := b.Send(context.Background())
	require.NoError(t, err)

	// Make sure the correct requests were sent
	assert.Equal(t, requests, capturedSend)

	// Make sure the correct responses were returned
	require.Len(t, responses, len(requests))

	for index, response := range responses {
		assert.Equal(t, requests[index].ID, response.ID)
		assert.Equal(t, requests[index].JSONRPC, response.JSONRPC)
		assert.Nil(t, response.Result)
		assert.Nil(t, response.Error)
	}

	// Make sure the batch has been cleared after sending
	assert.Equal(t, b.Count(), 0)
}

func TestBatch_Clear(t *testing.T) {
	t.Parallel()

	requests := generateRequests(t, 100)

	// Create the batch
	b := NewBatch(nil)

	// Add the requests
	for _, request := range requests {
		b.AddRequest(request)
	}

	// Clear the batch
	require.EqualValues(t, len(requests), b.Clear())

	// Make sure the batch is cleared
	require.Equal(t, b.Count(), 0)
}
