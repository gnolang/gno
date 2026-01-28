package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSendRequest_EmptyIDWithError tests that when the server returns a response
// with an empty ID and an error (e.g., body size limit exceeded), the client
// returns the actual error instead of "ID mismatch".
func TestSendRequest_EmptyIDWithError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := types.RPCResponse{
			JSONRPC: "2.0",
			ID:      types.JSONRPCStringID(""), // empty ID
			Error: &types.RPCError{
				Code:    -32600,
				Message: "Invalid Request",
				Data:    "error reading request body: http: request body too large",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	require.NoError(t, err)

	request, err := types.MapToRequest(
		types.JSONRPCStringID("test-id"),
		"test_method",
		map[string]any{},
	)
	require.NoError(t, err)

	_, err = client.SendRequest(context.Background(), request)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "request body too large")
	assert.NotContains(t, err.Error(), "mismatch")
}
