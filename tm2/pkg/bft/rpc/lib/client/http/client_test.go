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

// This test covers bug https://github.com/gnolang/gno/issues/3676
// TestSendRequestCommon_BatchSingleResponse simulates a batch call where the batch
// contains a single request. This test stays on the client side.
func TestSendRequestCommon_BatchSingleResponse(t *testing.T) {
	t.Parallel()

	// Step 1: Create a dummy batch request with a single item.
	singleRequest := types.RPCRequest{
		JSONRPC: "2.0",
		ID:      types.JSONRPCStringID("1"),
	}
	batchRequest := types.RPCRequests{singleRequest} // types.RPCRequests is defined as []RPCRequest

	// Step 2: Create the expected batch response.
	expectedResp := types.RPCResponse{
		JSONRPC: "2.0",
		ID:      singleRequest.ID,
	}
	expectedBatchResponse := types.RPCResponses{expectedResp} // types.RPCResponses is defined as []*RPCResponse

	// Step 3: Marshal the expected batch response as a JSON array.
	respBytes, err := json.Marshal(expectedBatchResponse)
	require.NoError(t, err)

	// Step 4: Create a test HTTP server that always returns the JSON array.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(respBytes)
		require.NoError(t, err)
	})
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Step 5: Create an HTTP client and a context with timeout.
	httpClient := ts.Client()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 6: Call sendRequestCommon with the batch request.
	// We choose R to be types.RPCResponses so we expect a batch (slice) in return.
	actualBatchResponse, err := sendRequestCommon[types.RPCRequests, types.RPCResponses](ctx, httpClient, ts.URL, batchRequest)
	require.NoError(t, err)

	// Step 7: Verify that the returned value is a slice with one element.
	assert.Len(t, actualBatchResponse, 1, "expected the batch response slice to have length one")

	// Step 8: Verify that the returned response fields match the expected values.
	assert.Equal(t, expectedBatchResponse[0].ID, actualBatchResponse[0].ID)
	assert.Equal(t, expectedBatchResponse[0].JSONRPC, actualBatchResponse[0].JSONRPC)
}

// TestSendRequestCommon_ErrorPaths tests error paths in sendRequestCommon without using a custom RoundTripper.
func TestSendRequestCommon_ErrorPaths(t *testing.T) {
	t.Parallel()

	// 1. HTTP request creation failure: Using an invalid URL.
	t.Run("HTTP request creation failure", func(t *testing.T) {
		t.Parallel()
		req := types.RPCRequest{JSONRPC: "2.0", ID: types.JSONRPCStringID("1")}
		ctx := context.Background()
		// Passing an invalid URL causes http.NewRequest to fail.
		resp, err := sendRequestCommon[types.RPCRequest, *types.RPCResponse](ctx, http.DefaultClient, "://invalid-url", req)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "unable to create request")
	})

	// 2. Client.Do failure: Cancel the context before the request is made.
	t.Run("Client.Do failure", func(t *testing.T) {
		t.Parallel()
		req := types.RPCRequest{JSONRPC: "2.0", ID: types.JSONRPCStringID("1")}
		// Create a context that is already canceled.
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		resp, err := sendRequestCommon[types.RPCRequest, *types.RPCResponse](ctx, http.DefaultClient, "http://example.com", req)
		require.Error(t, err)
		assert.Nil(t, resp)
		// Expect the error to mention the failure to send the request.
		assert.Contains(t, err.Error(), "unable to send request")
	})

	// 3. Non-OK HTTP status: Simulate a server returning a bad status code.
	t.Run("Non-OK HTTP status", func(t *testing.T) {
		t.Parallel()
		req := types.RPCRequest{JSONRPC: "2.0", ID: types.JSONRPCStringID("1")}
		// Create an httptest server that always returns 500.
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		ts := httptest.NewServer(handler)
		defer ts.Close()

		ctx := context.Background()
		resp, err := sendRequestCommon[types.RPCRequest, *types.RPCResponse](ctx, ts.Client(), ts.URL, req)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "invalid status code")
	})

	// 4. JSON unmarshal failure: Simulate invalid JSON in the response.
	t.Run("JSON unmarshal failure", func(t *testing.T) {
		t.Parallel()
		req := types.RPCRequest{JSONRPC: "2.0", ID: types.JSONRPCStringID("1")}
		// Create an httptest server that returns invalid JSON.
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("invalid json"))
			require.NoError(t, err)
		})
		ts := httptest.NewServer(handler)
		defer ts.Close()

		ctx := context.Background()
		resp, err := sendRequestCommon[types.RPCRequest, *types.RPCResponse](ctx, ts.Client(), ts.URL, req)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "unable to unmarshal")
	})
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
}
