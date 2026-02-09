package lib

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client/batch"
	httpclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client/http"
	wsclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client/ws"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/metadata"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testServer wraps the JSONRPC server for E2E testing
type testServer struct {
	listener net.Listener
	mux      *chi.Mux
	jsonrpc  *server.JSONRPC
}

// newTestServer creates a new test server instance
func newTestServer(t *testing.T) *testServer {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	jsonrpc := server.NewJSONRPC(server.WithLogger(log.NewNoopLogger()))

	mux := chi.NewMux()
	mux.Mount("/", jsonrpc.SetupRoutes(chi.NewMux()))

	return &testServer{
		listener: listener,
		mux:      mux,
		jsonrpc:  jsonrpc,
	}
}

func (s *testServer) start() {
	go func() {
		_ = http.Serve(s.listener, s.mux)
	}()
}

func (s *testServer) stop() {
	_ = s.listener.Close()
}

func (s *testServer) httpAddress() string {
	return fmt.Sprintf("http://%s", s.listener.Addr().String())
}

func (s *testServer) wsAddress() string {
	return fmt.Sprintf("ws://%s/websocket", s.listener.Addr().String())
}

func (s *testServer) registerHandler(method string, handler server.Handler, paramNames ...string) {
	s.jsonrpc.RegisterHandler(method, handler, paramNames...)
}

func TestE2E_HTTP(t *testing.T) {
	t.Parallel()

	t.Run("single request", func(t *testing.T) {
		t.Parallel()

		testTable := []struct {
			name           string
			method         string
			params         []any
			expectedResult string
			expectedError  *spec.BaseJSONError
		}{
			{
				name:           "simple method no params",
				method:         "echo",
				params:         nil,
				expectedResult: "hello",
				expectedError:  nil,
			},
			{
				name:           "method with string param",
				method:         "greet",
				params:         []any{"world"},
				expectedResult: "hello world",
				expectedError:  nil,
			},
			{
				name:           "method not found",
				method:         "nonexistent",
				params:         nil,
				expectedResult: "",
				expectedError: &spec.BaseJSONError{
					Code:    spec.MethodNotFoundErrorCode,
					Message: "Method handler not set",
				},
			},
		}

		for _, tc := range testTable {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				// Setup server
				srv := newTestServer(t)
				srv.registerHandler(
					"echo",
					func(_ *metadata.Metadata, _ []any) (any, *spec.BaseJSONError) {
						return "hello", nil
					},
				)

				srv.registerHandler(
					"greet",
					func(_ *metadata.Metadata, params []any) (any, *spec.BaseJSONError) {
						if len(params) == 0 {
							return nil, spec.GenerateInvalidParamError(0)
						}

						name, ok := params[0].(string)
						if !ok {
							return nil, spec.GenerateInvalidParamError(0)
						}

						return fmt.Sprintf("hello %s", name), nil
					},
					"name",
				)

				srv.start()
				defer srv.stop()

				// Create client
				client, err := httpclient.NewClient(srv.httpAddress())
				require.NoError(t, err)
				defer client.Close()

				// Send request
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				request := spec.NewJSONRequest(spec.JSONRPCNumberID(1), tc.method, tc.params)

				response, err := client.SendRequest(ctx, request)
				require.NoError(t, err)

				// Verify response
				assert.Equal(t, request.ID, response.ID)

				if tc.expectedError != nil {
					require.NotNil(t, response.Error)

					assert.Equal(t, tc.expectedError.Code, response.Error.Code)
					assert.Equal(t, tc.expectedError.Message, response.Error.Message)
				} else {
					assert.Nil(t, response.Error)
					assert.Contains(t, string(response.Result), tc.expectedResult)
				}
			})
		}
	})

	t.Run("batch request", func(t *testing.T) {
		t.Parallel()

		// Setup server
		srv := newTestServer(t)

		srv.registerHandler(
			"add",
			func(_ *metadata.Metadata, params []any) (any, *spec.BaseJSONError) {
				if len(params) < 2 {
					return nil, spec.GenerateInvalidParamError(0)
				}

				// JSON numbers come as float64, convert to int for amino compatibility
				a, ok1 := params[0].(float64)
				b, ok2 := params[1].(float64)

				if !ok1 || !ok2 {
					return nil, spec.GenerateInvalidParamError(0)
				}

				return int(a) + int(b), nil
			},
			"a", "b",
		)

		srv.start()
		defer srv.stop()

		// Create client
		client, err := httpclient.NewClient(srv.httpAddress())
		require.NoError(t, err)
		defer client.Close()

		// Create batch
		b := batch.NewBatch(client)
		b.AddRequest(spec.NewJSONRequest(spec.JSONRPCNumberID(1), "add", []any{1, 2}))
		b.AddRequest(spec.NewJSONRequest(spec.JSONRPCNumberID(2), "add", []any{3, 4}))
		b.AddRequest(spec.NewJSONRequest(spec.JSONRPCNumberID(3), "add", []any{5, 6}))

		require.Equal(t, 3, b.Count())

		// Send batch
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		responses, err := b.Send(ctx)
		require.NoError(t, err)
		require.Len(t, responses, 3)

		// Verify batch is cleared after send
		assert.Equal(t, 0, b.Count())

		// Verify responses
		expectedResults := []string{"3", "7", "11"}

		for i, response := range responses {
			assert.Equal(t, spec.JSONRPCNumberID(i+1), response.ID)
			assert.Nil(t, response.Error)
			assert.Contains(t, string(response.Result), expectedResults[i])
		}
	})

	t.Run("string ID", func(t *testing.T) {
		t.Parallel()

		// Setup server
		srv := newTestServer(t)

		srv.registerHandler(
			"ping",
			func(_ *metadata.Metadata, _ []any) (any, *spec.BaseJSONError) {
				return "pong", nil
			},
		)

		srv.start()
		defer srv.stop()

		// Create client
		client, err := httpclient.NewClient(srv.httpAddress())
		require.NoError(t, err)
		defer client.Close()

		// Send request with string ID
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id := spec.JSONRPCStringID("unique-id")
		request := spec.NewJSONRequest(id, "ping", nil)
		response, err := client.SendRequest(ctx, request)
		require.NoError(t, err)

		assert.Equal(t, id, response.ID)
		assert.Nil(t, response.Error)
		assert.Contains(t, string(response.Result), "pong")
	})
}

func TestE2E_WebSocket(t *testing.T) {
	t.Parallel()

	t.Run("single request", func(t *testing.T) {
		t.Parallel()

		testTable := []struct {
			name           string
			method         string
			params         []any
			expectedResult string
			expectedError  *spec.BaseJSONError
		}{
			{
				name:           "simple method no params",
				method:         "echo",
				params:         nil,
				expectedResult: "hello",
				expectedError:  nil,
			},
			{
				name:           "method with param",
				method:         "greet",
				params:         []any{"websocket"},
				expectedResult: "hello websocket",
				expectedError:  nil,
			},
		}

		for _, tc := range testTable {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				// Setup server
				srv := newTestServer(t)

				srv.registerHandler(
					"echo",
					func(_ *metadata.Metadata, _ []any) (any, *spec.BaseJSONError) {
						return "hello", nil
					},
				)

				srv.registerHandler(
					"greet",
					func(_ *metadata.Metadata, params []any) (any, *spec.BaseJSONError) {
						if len(params) == 0 {
							return nil, spec.GenerateInvalidParamError(0)
						}

						name, ok := params[0].(string)
						if !ok {
							return nil, spec.GenerateInvalidParamError(0)
						}

						return fmt.Sprintf("hello %s", name), nil
					},
					"name",
				)

				srv.start()
				defer srv.stop()

				// Create WS client
				client, err := wsclient.NewClient(srv.wsAddress())
				require.NoError(t, err)
				defer client.Close()

				// Send request
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				request := spec.NewJSONRequest(spec.JSONRPCNumberID(1), tc.method, tc.params)
				response, err := client.SendRequest(ctx, request)
				require.NoError(t, err)

				// Verify response
				assert.Equal(t, request.ID, response.ID)

				if tc.expectedError != nil {
					require.NotNil(t, response.Error)
					assert.Equal(t, tc.expectedError.Code, response.Error.Code)
				} else {
					assert.Nil(t, response.Error)
					assert.Contains(t, string(response.Result), tc.expectedResult)
				}
			})
		}
	})

	t.Run("batch request", func(t *testing.T) {
		t.Parallel()

		// Setup server
		srv := newTestServer(t)

		srv.registerHandler(
			"multiply",
			func(_ *metadata.Metadata, params []any) (any, *spec.BaseJSONError) {
				if len(params) < 2 {
					return nil, spec.GenerateInvalidParamError(0)
				}

				// JSON numbers come as float64, convert to int for amino compatibility
				a, ok1 := params[0].(float64)
				b, ok2 := params[1].(float64)

				if !ok1 || !ok2 {
					return nil, spec.GenerateInvalidParamError(0)
				}

				return int(a) * int(b), nil
			},
			"a", "b",
		)

		srv.start()
		defer srv.stop()

		// Create WS client
		client, err := wsclient.NewClient(srv.wsAddress())
		require.NoError(t, err)
		defer client.Close()

		// Create batch
		b := batch.NewBatch(client)
		b.AddRequest(spec.NewJSONRequest(spec.JSONRPCNumberID(1), "multiply", []any{2, 3}))
		b.AddRequest(spec.NewJSONRequest(spec.JSONRPCNumberID(2), "multiply", []any{4, 5}))

		// Send batch
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		responses, err := b.Send(ctx)
		require.NoError(t, err)
		require.Len(t, responses, 2)

		// Verify responses (results should be 6, 20)
		expectedResults := []string{"6", "20"}
		for i, response := range responses {
			assert.Equal(t, spec.JSONRPCNumberID(i+1), response.ID)
			assert.Nil(t, response.Error)
			assert.Contains(t, string(response.Result), expectedResults[i])
		}
	})

	t.Run("multiple sequential requests", func(t *testing.T) {
		t.Parallel()

		// Setup server
		srv := newTestServer(t)

		srv.registerHandler(
			"counter",
			func(_ *metadata.Metadata, params []any) (any, *spec.BaseJSONError) {
				if len(params) == 0 {
					return nil, spec.GenerateInvalidParamError(0)
				}

				// JSON numbers come as float64
				n, ok := params[0].(float64)
				if !ok {
					return nil, spec.GenerateInvalidParamError(0)
				}

				return int(n) + 1, nil
			},
			"n",
		)

		srv.start()
		defer srv.stop()

		// Create WS client
		client, err := wsclient.NewClient(srv.wsAddress())
		require.NoError(t, err)
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Send multiple sequential requests
		for i := 0; i < 5; i++ {
			request := spec.NewJSONRequest(spec.JSONRPCNumberID(i+1), "counter", []any{i})
			response, err := client.SendRequest(ctx, request)
			require.NoError(t, err)

			assert.Equal(t, spec.JSONRPCNumberID(i+1), response.ID)
			assert.Nil(t, response.Error)
			assert.Contains(t, string(response.Result), fmt.Sprintf("%d", i+1))
		}
	})
}

func TestE2E_RequestErrors(t *testing.T) {
	t.Parallel()

	t.Run("invalid JSON-RPC version", func(t *testing.T) {
		t.Parallel()

		// Setup server
		srv := newTestServer(t)

		srv.registerHandler(
			"test",
			func(_ *metadata.Metadata, _ []any) (any, *spec.BaseJSONError) {
				return "ok", nil
			},
		)

		srv.start()
		defer srv.stop()

		// Create client
		client, err := httpclient.NewClient(srv.httpAddress())
		require.NoError(t, err)
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Create request with wrong JSON-RPC version
		request := &spec.BaseJSONRequest{
			BaseJSON: spec.BaseJSON{
				JSONRPC: "1.0", // Invalid version
				ID:      spec.JSONRPCNumberID(1),
			},
			Method: "test",
			Params: nil,
		}

		response, err := client.SendRequest(ctx, request)
		require.NoError(t, err)

		assert.NotNil(t, response.Error)
		assert.Equal(t, spec.InvalidRequestErrorCode, response.Error.Code)
	})

	t.Run("handler returns error", func(t *testing.T) {
		t.Parallel()

		// Setup server
		srv := newTestServer(t)

		srv.registerHandler(
			"fail",
			func(_ *metadata.Metadata, _ []any) (any, *spec.BaseJSONError) {
				return nil, spec.NewJSONError("kaboom", spec.ServerErrorCode)
			},
		)

		srv.start()
		defer srv.stop()

		// Create client
		client, err := httpclient.NewClient(srv.httpAddress())
		require.NoError(t, err)
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		request := spec.NewJSONRequest(spec.JSONRPCNumberID(1), "fail", nil)
		response, err := client.SendRequest(ctx, request)
		require.NoError(t, err)

		assert.NotNil(t, response.Error)
		assert.Equal(t, spec.ServerErrorCode, response.Error.Code)
		assert.Equal(t, "kaboom", response.Error.Message)
	})
}

func TestE2E_Metadata(t *testing.T) {
	t.Parallel()

	t.Run("HTTP metadata contains remote address", func(t *testing.T) {
		t.Parallel()

		var capturedMetadata *metadata.Metadata

		// Setup server
		srv := newTestServer(t)
		srv.registerHandler(
			"capture",
			func(m *metadata.Metadata, _ []any) (any, *spec.BaseJSONError) {
				capturedMetadata = m

				return "captured", nil
			},
		)

		srv.start()
		defer srv.stop()

		// Create client
		client, err := httpclient.NewClient(srv.httpAddress())
		require.NoError(t, err)
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		request := spec.NewJSONRequest(spec.JSONRPCNumberID(1), "capture", nil)
		_, err = client.SendRequest(ctx, request)
		require.NoError(t, err)

		require.NotNil(t, capturedMetadata)

		assert.NotEmpty(t, capturedMetadata.RemoteAddr)
		assert.Nil(t, capturedMetadata.WebSocketID)
	})

	t.Run("WebSocket metadata contains connection ID", func(t *testing.T) {
		t.Parallel()

		var capturedMetadata *metadata.Metadata

		// Setup server
		srv := newTestServer(t)

		srv.registerHandler(
			"capture",
			func(m *metadata.Metadata, _ []any) (any, *spec.BaseJSONError) {
				capturedMetadata = m

				return "captured", nil
			},
		)

		srv.start()
		defer srv.stop()

		// Create WS client
		client, err := wsclient.NewClient(srv.wsAddress())
		require.NoError(t, err)
		defer client.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		request := spec.NewJSONRequest(spec.JSONRPCNumberID(1), "capture", nil)
		_, err = client.SendRequest(ctx, request)
		require.NoError(t, err)

		require.NotNil(t, capturedMetadata)
		assert.NotEmpty(t, capturedMetadata.RemoteAddr)

		// WebSocket connections should have a connection ID
		require.NotNil(t, capturedMetadata.WebSocketID)
		assert.NotEmpty(t, *capturedMetadata.WebSocketID)
	})
}
