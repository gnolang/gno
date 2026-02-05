package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/metadata"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func decodeResponse[T spec.BaseJSONResponse | spec.BaseJSONResponses](t *testing.T, responseBody []byte) *T {
	t.Helper()

	var response *T

	require.NoError(t, json.NewDecoder(bytes.NewReader(responseBody)).Decode(&response))

	return response
}

// setupTestWebServer is a helper function for common setup logic
func setupTestWebServer(t *testing.T, callback func(s *JSONRPC)) *testWebServer {
	t.Helper()

	s := newWebServer(t, callback)
	s.start()

	return s
}

// TestHTTP_Handle verifies that the JSON-RPC server:
// - can handle a single HTTP request to a dummy endpoint
// - can handle a batch HTTP request to a dummy endpoint
func TestHTTP_Handle(t *testing.T) {
	t.Parallel()

	var (
		commonResponse = "This is a common response!"
		method         = "dummy"
	)

	singleRequest, err := json.Marshal(
		spec.NewJSONRequest(spec.JSONRPCNumberID(1), method, nil),
	)
	require.NoError(t, err)

	requests := spec.BaseJSONRequests{
		spec.NewJSONRequest(spec.JSONRPCNumberID(1), method, nil),
		spec.NewJSONRequest(spec.JSONRPCNumberID(2), method, nil),
		spec.NewJSONRequest(spec.JSONRPCNumberID(3), method, nil),
	}

	batchRequest, err := json.Marshal(requests)
	require.NoError(t, err)

	testTable := []struct {
		verifyResponse func(response []byte) error
		name           string
		request        []byte
	}{
		{
			func(resp []byte) error {
				response := decodeResponse[spec.BaseJSONResponse](t, resp)

				assert.Equal(t, spec.NewJSONResponse(spec.JSONRPCNumberID(1), commonResponse, nil), response)

				return nil
			},
			"single HTTP request",
			singleRequest,
		},
		{
			func(resp []byte) error {
				responses := decodeResponse[spec.BaseJSONResponses](t, resp)

				for index, response := range *responses {
					assert.Equal(
						t,
						spec.NewJSONResponse(spec.JSONRPCNumberID(index+1), commonResponse, nil),
						response,
					)
				}

				return nil
			},
			"batch HTTP request",
			batchRequest,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create a new JSON-RPC server
			webServer := setupTestWebServer(t, func(s *JSONRPC) {
				s.handlers = make(handlers)

				s.handlers.addHandler(method, func(_ *metadata.Metadata, _ []any) (any, *spec.BaseJSONError) {
					return commonResponse, nil
				})
			})

			defer webServer.stop()

			respRaw, err := http.Post(
				webServer.address(),
				jsonMimeType,
				bytes.NewBuffer(testCase.request),
			)
			if err != nil {
				t.Fatalf("unexpected HTTP error, %v", err)
			}

			resp, err := io.ReadAll(respRaw.Body)
			if err != nil {
				t.Fatalf("unable to read response body, %v", err)
			}

			if err := testCase.verifyResponse(resp); err != nil {
				t.Fatalf("unable to verify response, %v", err)
			}
		})
	}
}

type testWebServer struct {
	mux      *chi.Mux
	listener net.Listener
}

func newWebServer(t *testing.T, callbacks ...func(s *JSONRPC)) *testWebServer {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("unable to start listen, %v", err)
	}

	mux := chi.NewMux()
	webServer := &testWebServer{
		mux:      mux,
		listener: listener,
	}

	s := NewJSONRPC(WithLogger(log.NewNoopLogger()))

	for _, callback := range callbacks {
		callback(s)
	}

	// Hook up the JSON-RPC server to the mux
	mux.Mount("/", s.SetupRoutes(chi.NewMux()))

	return webServer
}

func (ms *testWebServer) start() {
	go func() {
		//nolint:errcheck // No need to check error
		_ = http.Serve(ms.listener, ms.mux)
	}()
}

func (ms *testWebServer) stop() {
	_ = ms.listener.Close()
}

func (ms *testWebServer) address() string {
	return fmt.Sprintf("http://%s", ms.listener.Addr().String())
}
