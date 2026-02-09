package batch

import (
	"context"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateRequests generates dummy RPC requests
func generateRequests(t *testing.T, count int) spec.BaseJSONRequests {
	t.Helper()

	requests := make(spec.BaseJSONRequests, 0, count)

	for i := range count {
		requests = append(
			requests,
			spec.NewJSONRequest(
				spec.JSONRPCNumberID(i),
				"",
				nil,
			),
		)
	}

	return requests
}

func TestBatch_AddRequest(t *testing.T) {
	t.Parallel()

	var (
		capturedSend spec.BaseJSONRequests
		requests     = generateRequests(t, 100)

		mockClient = &mockClient{
			sendBatchFn: func(_ context.Context, requests spec.BaseJSONRequests) (spec.BaseJSONResponses, error) {
				capturedSend = requests

				responses := make(spec.BaseJSONResponses, len(requests))

				for index, request := range requests {
					responses[index] = spec.NewJSONResponse(request.ID, nil, nil)
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
