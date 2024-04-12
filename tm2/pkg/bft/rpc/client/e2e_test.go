package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gorilla/websocket"
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

func defaultHTTPHandler(t *testing.T, responseBytes []byte) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("content-type"))

		// Parse the message
		var req types.RPCRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))

		// Marshal the result data to Amino JSON
		result, err := amino.MarshalJSON(responseBytes)
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

func defaultWSHandler(t *testing.T, responseBytes []byte) http.HandlerFunc {
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

			// Marshal the result data to Amino JSON
			result, err := amino.MarshalJSON(responseBytes)
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

func TestRPCClient_E2E_Status(t *testing.T) {
	t.Parallel() // TODO implement
}
