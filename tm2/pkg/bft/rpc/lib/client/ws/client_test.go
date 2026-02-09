package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestServer creates a test WS server
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

	t.Run("request timed out", func(t *testing.T) {
		t.Parallel()

		var (
			upgrader = websocket.Upgrader{}

			request = spec.NewJSONRequest(
				spec.JSONRPCStringID("id"),
				"",
				nil,
			)
		)

		ctx, cancelFn := context.WithCancel(context.Background())
		defer cancelFn()

		// Create the server
		handler := func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)

			defer c.Close()

			for {
				_, message, err := c.ReadMessage()
				if websocket.IsUnexpectedCloseError(err) {
					return
				}

				require.NoError(t, err)

				// Parse the message
				var req *spec.BaseJSONRequest
				require.NoError(t, json.Unmarshal(message, &req))
				require.Equal(t, request.ID.String(), req.ID.String())

				// Simulate context cancellation mid-request parsing
				cancelFn()
			}
		}

		s := createTestServer(t, http.HandlerFunc(handler))
		url := "ws" + strings.TrimPrefix(s.URL, "http")

		// Create the client
		c, err := NewClient(url)
		require.NoError(t, err)

		defer func() {
			assert.NoError(t, c.Close())
		}()

		// Try to send the request, but wait for
		// the context to be cancelled
		response, err := c.SendRequest(ctx, request)
		require.Nil(t, response)

		assert.ErrorIs(t, err, ErrTimedOut)
	})

	t.Run("valid request sent", func(t *testing.T) {
		t.Parallel()

		var (
			upgrader = websocket.Upgrader{}

			request = spec.NewJSONRequest(
				spec.JSONRPCStringID("id"),
				"",
				nil,
			)

			response = spec.NewJSONResponse(
				request.ID,
				nil,
				nil,
			)
		)

		// Create the server
		handler := func(w http.ResponseWriter, r *http.Request) {
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
				var req *spec.BaseJSONRequest
				require.NoError(t, json.Unmarshal(message, &req))
				require.Equal(t, request.ID.String(), req.ID.String())

				marshalledResponse, err := json.Marshal(response)
				require.NoError(t, err)

				require.NoError(t, c.WriteMessage(mt, marshalledResponse))
			}
		}

		s := createTestServer(t, http.HandlerFunc(handler))
		url := "ws" + strings.TrimPrefix(s.URL, "http")

		// Create the client
		c, err := NewClient(url)
		require.NoError(t, err)

		defer func() {
			assert.NoError(t, c.Close())
		}()

		// Try to send the valid request
		ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*5)
		defer cancelFn()

		resp, err := c.SendRequest(ctx, request)
		require.NoError(t, err)

		assert.Equal(t, response.ID, resp.ID)
		assert.Equal(t, response.JSONRPC, resp.JSONRPC)
		assert.Equal(t, response.Result, resp.Result)
		assert.Equal(t, response.Error, resp.Error)
	})
}

func TestClient_SendBatch(t *testing.T) {
	t.Parallel()

	t.Run("batch timed out", func(t *testing.T) {
		t.Parallel()

		var (
			upgrader = websocket.Upgrader{}

			request = spec.NewJSONRequest(
				spec.JSONRPCStringID("id"),
				"",
				nil,
			)

			batch = spec.BaseJSONRequests{request}
		)

		ctx, cancelFn := context.WithCancel(context.Background())
		defer cancelFn()

		// Create the server
		handler := func(w http.ResponseWriter, r *http.Request) {
			c, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)

			defer c.Close()

			for {
				_, message, err := c.ReadMessage()
				if websocket.IsUnexpectedCloseError(err) {
					return
				}

				require.NoError(t, err)

				// Parse the message
				var req spec.BaseJSONRequests
				require.NoError(t, json.Unmarshal(message, &req))

				require.Len(t, req, 1)
				require.Equal(t, request.ID.String(), req[0].ID.String())

				// Simulate context cancellation mid-request parsing
				cancelFn()
			}
		}

		s := createTestServer(t, http.HandlerFunc(handler))
		url := "ws" + strings.TrimPrefix(s.URL, "http")

		// Create the client
		c, err := NewClient(url)
		require.NoError(t, err)

		defer func() {
			assert.NoError(t, c.Close())
		}()

		// Try to send the request, but wait for
		// the context to be cancelled
		response, err := c.SendBatch(ctx, batch)
		require.Nil(t, response)

		assert.ErrorIs(t, err, ErrTimedOut)
	})

	t.Run("valid batch sent", func(t *testing.T) {
		t.Parallel()

		var (
			upgrader = websocket.Upgrader{}

			request = spec.NewJSONRequest(
				spec.JSONRPCStringID("id"),
				"",
				nil,
			)

			response = spec.NewJSONResponse(
				request.ID,
				nil,
				nil,
			)

			batch         = spec.BaseJSONRequests{request}
			batchResponse = spec.BaseJSONResponses{response}
		)

		// Create the server
		handler := func(w http.ResponseWriter, r *http.Request) {
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
				var req spec.BaseJSONRequests
				require.NoError(t, json.Unmarshal(message, &req))

				require.Len(t, req, 1)
				require.Equal(t, request.ID.String(), req[0].ID.String())

				marshalledResponse, err := json.Marshal(batchResponse)
				require.NoError(t, err)

				require.NoError(t, c.WriteMessage(mt, marshalledResponse))
			}
		}

		s := createTestServer(t, http.HandlerFunc(handler))
		url := "ws" + strings.TrimPrefix(s.URL, "http")

		// Create the client
		c, err := NewClient(url)
		require.NoError(t, err)

		defer func() {
			assert.NoError(t, c.Close())
		}()

		// Try to send the valid request
		ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*5)
		defer cancelFn()

		resp, err := c.SendBatch(ctx, batch)
		require.NoError(t, err)

		require.Len(t, resp, 1)

		assert.Equal(t, response.ID, resp[0].ID)
		assert.Equal(t, response.JSONRPC, resp[0].JSONRPC)
		assert.Equal(t, response.Result, resp[0].Result)
		assert.Equal(t, response.Error, resp[0].Error)
	})
}
