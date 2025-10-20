package rpcserver

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/log"
)

func TestMaxOpenConnections(t *testing.T) {
	t.Parallel()

	const maxVal = 5 // max simultaneous connections

	// Start the server.
	var open int32
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if n := atomic.AddInt32(&open, 1); n > int32(maxVal) {
			t.Errorf("%d open connections, want <= %d", n, maxVal)
		}
		defer atomic.AddInt32(&open, -1)
		time.Sleep(10 * time.Millisecond)
		fmt.Fprint(w, "some body")
	})
	config := DefaultConfig()
	config.MaxOpenConnections = maxVal
	l, err := Listen("tcp://127.0.0.1:0", config)
	require.NoError(t, err)
	defer l.Close()
	go StartHTTPServer(l, mux, log.NewTestingLogger(t), config)

	// Make N GET calls to the server.
	attempts := maxVal * 2
	var wg sync.WaitGroup
	var failed int32
	for range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := http.Client{Timeout: 3 * time.Second}
			r, err := c.Get("http://" + l.Addr().String())
			if err != nil {
				t.Log(err)
				atomic.AddInt32(&failed, 1)
				return
			}
			defer r.Body.Close()
			io.Copy(io.Discard, r.Body)
		}()
	}
	wg.Wait()

	// We expect some Gets to fail as the server's accept queue is filled,
	// but most should succeed.
	if int(failed) >= attempts/2 {
		t.Errorf("%d requests failed within %d attempts", failed, attempts)
	}
}

func TestStartHTTPAndTLSServer(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer ln.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "some body")
	})

	go StartHTTPAndTLSServer(ln, mux, "test.crt", "test.key", log.NewTestingLogger(t), DefaultConfig())

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c := &http.Client{Transport: tr}
	res, err := c.Get("https://" + ln.Addr().String())
	require.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, []byte("some body"), body)
}

func TestRecoverAndLogHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		panicArg         any
		expectedResponse string
	}{
		{
			name:     "panic with types.RPCResponse",
			panicArg: types.NewRPCErrorResponse(types.JSONRPCStringID("id"), 42, "msg", "data"),
			expectedResponse: `{
  "jsonrpc": "2.0",
  "id": "id",
  "error": {
    "code": 42,
    "message": "msg",
    "data": "data"
  }
}`,
		},
		{
			name:     "panic with error",
			panicArg: fmt.Errorf("I'm an error"),
			expectedResponse: `{
  "jsonrpc": "2.0",
  "id": "",
  "error": {
    "code": -32603,
    "message": "Internal error",
    "data": "I'm an error"
  }
}`,
		},
		{
			name:     "panic with string",
			panicArg: "I'm an string",
			expectedResponse: `{
  "jsonrpc": "2.0",
  "id": "",
  "error": {
    "code": -32603,
    "message": "Internal error",
    "data": "I'm an string"
  }
}`,
		},
		{
			name: "panic with random struct",
			panicArg: struct {
				f int
			}{f: 1},
			expectedResponse: `{
  "jsonrpc": "2.0",
  "id": "",
  "error": {
    "code": -32603,
    "message": "Internal error",
    "data": "{1}"
  }
}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				req, _ = http.NewRequest(http.MethodGet, "", nil)
				resp   = httptest.NewRecorder()
				logger = log.NewNoopLogger()
				// Create a handler that will always panic with argument tt.panicArg
				handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					panic(tt.panicArg)
				})
			)

			RecoverAndLogHandler(handler, logger, noop.NewTracerProvider().Tracer("test")).ServeHTTP(resp, req)

			require.Equal(t, tt.expectedResponse, resp.Body.String())
		})
	}
}
