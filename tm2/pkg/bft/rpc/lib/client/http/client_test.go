package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_parseRemoteAddr(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		remoteAddr      string
		expectedNetwork string
		expectedRest    string
	}{
		{
			"127.0.0.1",
			"tcp",
			"127.0.0.1:80",
		},
		{
			"127.0.0.1:5000",
			"tcp",
			"127.0.0.1:5000",
		},
		{
			"http://example.com",
			"http",
			"example.com:80",
		},
		{
			"https://example.com",
			"https",
			"example.com:443",
		},
		{
			"http://example.com:5000",
			"http",
			"example.com:5000",
		},
		{
			"https://example.com:5000",
			"https",
			"example.com:5000",
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.remoteAddr, func(t *testing.T) {
			t.Parallel()

			n, r := parseRemoteAddr(testCase.remoteAddr)

			assert.Equal(t, testCase.expectedNetwork, n)
			assert.Equal(t, testCase.expectedRest, r)
		})
	}
}

// Following tests check that we correctly translate http/https to tcp,
// and other protocols are left intact from parseRemoteAddr()

func TestClient_makeHTTPDialer(t *testing.T) {
	t.Parallel()

	t.Run("http", func(t *testing.T) {
		t.Parallel()

		_, err := makeHTTPDialer("https://.")
		require.Error(t, err)

		assert.Contains(t, err.Error(), "dial tcp:", "should convert https to tcp")
	})

	t.Run("udp", func(t *testing.T) {
		t.Parallel()

		_, err := makeHTTPDialer("udp://.")
		require.Error(t, err)

		assert.Contains(t, err.Error(), "dial udp:", "udp protocol should remain the same")
	})
}

// createTestServer creates a test HTTP server
func createTestServer(
	t *testing.T,
	handler http.Handler,
) *httptest.Server {
	t.Helper()

	s := httptest.NewServer(handler)
	t.Cleanup(s.Close)

	return s
}

func TestClient_SendRequest(t *testing.T) {
	t.Parallel()

	t.Run("valid request, response", func(t *testing.T) {
		t.Parallel()

		var (
			request = types.RPCRequest{
				JSONRPC: "2.0",
				ID:      types.JSONRPCStringID("id"),
			}

			handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "application/json", r.Header.Get("content-type"))

				// Parse the message
				var req types.RPCRequest
				require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
				require.Equal(t, request.ID.String(), req.ID.String())

				// Send an empty response back
				response := types.RPCResponse{
					JSONRPC: "2.0",
					ID:      req.ID,
				}

				// Marshal the response
				marshalledResponse, err := json.Marshal(response)
				require.NoError(t, err)

				_, err = w.Write(marshalledResponse)
				require.NoError(t, err)
			})

			server = createTestServer(t, handler)
		)

		// Create the client
		c, err := NewClient(server.URL)
		require.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*5)
		defer cancelFn()

		// Send the request
		resp, err := c.SendRequest(ctx, request)
		require.NoError(t, err)

		assert.Equal(t, request.ID, resp.ID)
		assert.Equal(t, request.JSONRPC, resp.JSONRPC)
		assert.Nil(t, resp.Result)
		assert.Nil(t, resp.Error)
	})

	t.Run("response ID mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			request = types.RPCRequest{
				JSONRPC: "2.0",
				ID:      types.JSONRPCStringID("id"),
			}

			handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "application/json", r.Header.Get("content-type"))

				// Send an empty response back,
				// with an invalid ID
				response := types.RPCResponse{
					JSONRPC: "2.0",
					ID:      types.JSONRPCStringID("totally random ID"),
				}

				// Marshal the response
				marshalledResponse, err := json.Marshal(response)
				require.NoError(t, err)

				_, err = w.Write(marshalledResponse)
				require.NoError(t, err)
			})

			server = createTestServer(t, handler)
		)

		// Create the client
		c, err := NewClient(server.URL)
		require.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*5)
		defer cancelFn()

		// Send the request
		resp, err := c.SendRequest(ctx, request)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, ErrRequestResponseIDMismatch)
	})

	t.Run("empty ID with error returns error", func(t *testing.T) {
		t.Parallel()

		var (
			request = types.RPCRequest{
				JSONRPC: "2.0",
				ID:      types.JSONRPCStringID("id"),
			}

			expectedError = &types.RPCError{
				Code:    -32600,
				Message: "Invalid Request",
				Data:    "some error details",
			}

			handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "application/json", r.Header.Get("content-type"))

				// Send a response with empty ID and an error
				response := types.RPCResponse{
					JSONRPC: "2.0",
					ID:      nil, // Empty ID
					Error:   expectedError,
				}

				// Marshal the response
				marshalledResponse, err := json.Marshal(response)
				require.NoError(t, err)

				_, err = w.Write(marshalledResponse)
				require.NoError(t, err)
			})

			server = createTestServer(t, handler)
		)

		// Create the client
		c, err := NewClient(server.URL)
		require.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*5)
		defer cancelFn()

		// Send the request
		resp, err := c.SendRequest(ctx, request)
		assert.Nil(t, resp)
		require.Error(t, err)

		// Should return the actual error, not ID mismatch
		assert.NotErrorIs(t, err, ErrRequestResponseIDMismatch)

		// Verify it's the RPC error
		var rpcErr *types.RPCError
		assert.ErrorAs(t, err, &rpcErr)
		assert.Equal(t, expectedError.Code, rpcErr.Code)
		assert.Equal(t, expectedError.Message, rpcErr.Message)
		assert.Equal(t, expectedError.Data, rpcErr.Data)
	})
}

func TestClient_SendBatchRequest(t *testing.T) {
	t.Parallel()

	var (
		request = types.RPCRequest{
			JSONRPC: "2.0",
			ID:      types.JSONRPCStringID("id"),
		}

		requests = types.RPCRequests{
			request,
			request,
		}

		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "application/json", r.Header.Get("content-type"))

			// Parse the message
			var reqs types.RPCRequests
			require.NoError(t, json.NewDecoder(r.Body).Decode(&reqs))
			require.Len(t, reqs, len(requests))

			for _, req := range reqs {
				require.Equal(t, request.ID.String(), req.ID.String())
			}

			// Send an empty response batch back
			response := types.RPCResponse{
				JSONRPC: "2.0",
				ID:      request.ID,
			}

			responses := types.RPCResponses{
				response,
				response,
			}

			// Marshal the responses
			marshalledResponses, err := json.Marshal(responses)
			require.NoError(t, err)

			_, err = w.Write(marshalledResponses)
			require.NoError(t, err)
		})

		server = createTestServer(t, handler)
	)

	// Create the client
	c, err := NewClient(server.URL)
	require.NoError(t, err)

	ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelFn()

	// Send the request
	resps, err := c.SendBatch(ctx, requests)
	require.NoError(t, err)

	require.Len(t, resps, len(requests))

	for _, resp := range resps {
		assert.Equal(t, request.ID, resp.ID)
		assert.Equal(t, request.JSONRPC, resp.JSONRPC)
		assert.Nil(t, resp.Result)
		assert.Nil(t, resp.Error)
	}

	t.Run("empty ID with error in batch returns error", func(t *testing.T) {
		t.Parallel()

		var (
			request1 = types.RPCRequest{
				JSONRPC: "2.0",
				ID:      types.JSONRPCStringID("id1"),
			}
			request2 = types.RPCRequest{
				JSONRPC: "2.0",
				ID:      types.JSONRPCStringID("id2"),
			}

			requests = types.RPCRequests{
				request1,
				request2,
			}

			expectedError = &types.RPCError{
				Code:    -32700,
				Message: "Parse error",
				Data:    "batch processing failed",
			}

			handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "application/json", r.Header.Get("content-type"))

				// Send a batch response where one has empty ID and an error
				responses := types.RPCResponses{
					{
						JSONRPC: "2.0",
						ID:      request1.ID,
					},
					{
						JSONRPC: "2.0",
						ID:      nil, // Empty ID
						Error:   expectedError,
					},
				}

				// Marshal the responses
				marshalledResponses, err := json.Marshal(responses)
				require.NoError(t, err)

				_, err = w.Write(marshalledResponses)
				require.NoError(t, err)
			})

			server = createTestServer(t, handler)
		)

		// Create the client
		c, err := NewClient(server.URL)
		require.NoError(t, err)

		ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*5)
		defer cancelFn()

		// Send the batch request
		resps, err := c.SendBatch(ctx, requests)
		assert.Nil(t, resps)
		require.Error(t, err)

		// Should return the actual error, not ID mismatch
		assert.NotErrorIs(t, err, ErrRequestResponseIDMismatch)

		// Verify it's the RPC error
		var rpcErr *types.RPCError
		assert.ErrorAs(t, err, &rpcErr)
		assert.Equal(t, expectedError.Code, rpcErr.Code)
		assert.Equal(t, expectedError.Message, rpcErr.Message)
		assert.Equal(t, expectedError.Data, rpcErr.Data)
	})
}
