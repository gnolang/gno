package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestServer creates a test RPC server
func createTestServer(
	t *testing.T,
	handler http.Handler,
) *httptest.Server {
	t.Helper()

	s := httptest.NewServer(handler)
	t.Cleanup(s.Close)

	return s
}

func defaultHTTPHandler(
	t *testing.T,
	method string,
	responseResult any,
) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("content-type"))

		// Parse the message
		var req types.RPCRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))

		// Basic request validation
		require.Equal(t, req.JSONRPC, "2.0")
		require.Equal(t, req.Method, method)

		// Marshal the result data to Amino JSON
		result, err := amino.MarshalJSON(responseResult)
		require.NoError(t, err)

		// Send a response back
		response := types.RPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		}

		// Marshal the response
		marshalledResponse, err := json.Marshal(response)
		require.NoError(t, err)

		_, err = w.Write(marshalledResponse)
		require.NoError(t, err)
	}
}

func defaultWSHandler(
	t *testing.T,
	method string,
	responseResult any,
) http.HandlerFunc {
	t.Helper()

	upgrader := websocket.Upgrader{}

	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)

		defer c.Close()

		for {
			mt, message, err := c.ReadMessage()
			if websocket.IsUnexpectedCloseError(err) {
				return
			}

			require.NoError(t, err)

			// Parse the message
			var req types.RPCRequest
			require.NoError(t, json.Unmarshal(message, &req))

			// Basic request validation
			require.Equal(t, req.JSONRPC, "2.0")
			require.Equal(t, req.Method, method)

			// Marshal the result data to Amino JSON
			result, err := amino.MarshalJSON(responseResult)
			require.NoError(t, err)

			// Send a response back
			response := types.RPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  result,
			}

			// Marshal the response
			marshalledResponse, err := json.Marshal(response)
			require.NoError(t, err)

			require.NoError(t, c.WriteMessage(mt, marshalledResponse))
		}
	}
}

type e2eTestCase struct {
	name   string
	client *RPCClient
}

func generateE2ETestCases(
	t *testing.T,
	method string,
	responseResult any,
) []e2eTestCase {
	t.Helper()

	// Create the http client
	httpServer := createTestServer(t, defaultHTTPHandler(t, method, responseResult))
	httpClient, err := NewHTTPClient(httpServer.URL)
	require.NoError(t, err)

	// Create the WS client
	wsServer := createTestServer(t, defaultWSHandler(t, method, responseResult))
	wsClient, err := NewWSClient("ws" + strings.TrimPrefix(wsServer.URL, "http"))
	require.NoError(t, err)

	return []e2eTestCase{
		{
			name:   "http",
			client: httpClient,
		},
		{
			name:   "ws",
			client: wsClient,
		},
	}
}

func TestRPCClient_E2E_Status(t *testing.T) {
	t.Parallel()

	var (
		expectedStatus = &ctypes.ResultStatus{
			NodeInfo: p2p.NodeInfo{
				Moniker: "dummy",
			},
		}
	)

	testTable := generateE2ETestCases(t, statusMethod, expectedStatus)

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				require.NoError(t, testCase.client.Close())
			}()

			status, err := testCase.client.Status()
			require.NoError(t, err)

			assert.Equal(t, expectedStatus, status)
		})
	}
}
