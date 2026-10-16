package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPCGateway(t *testing.T) {
	// The fake node records what got through: a refused request must never
	// reach it, which is the whole point of filtering before proxying.
	var reached []string
	node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		reached = append(reached, r.Method+" "+r.URL.RequestURI()+" "+string(body))
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(node.Close)

	gw, err := newRPCGateway("tcp://"+strings.TrimPrefix(node.URL, "http://"), 1024)
	require.NoError(t, err)
	handler := http.StripPrefix("/rpc", gw)

	cases := []struct {
		name   string
		method string
		path   string
		body   string
		status int
		reach  string // what the node sees; empty when refused
	}{
		{"uri route", "GET", "/rpc/status", "", 200, "GET /status "},
		{"uri route with query", "GET", `/rpc/abci_query?path="auth/gasprice"`, "", 200, `GET /abci_query?path="auth/gasprice" `},
		{"route listing", "GET", "/rpc", "", 200, "GET / "},
		{"route listing, slash", "GET", "/rpc/", "", 200, "GET / "},
		{"uri unsafe", "GET", "/rpc/unsafe_write_heap_profile?filename=/etc/x", "", 403, ""},
		{"uri cpu profiler", "GET", "/rpc/unsafe_start_cpu_profiler", "", 403, ""},
		{"uri dial", "GET", "/rpc/dial_seeds", "", 403, ""},
		{"websocket", "GET", "/rpc/websocket", "", 403, ""},
		{"unknown route", "GET", "/rpc/whatever", "", 403, ""},
		{"jsonrpc", "POST", "/rpc", `{"method":"status"}`, 200, `POST / {"method":"status"}`},
		{"jsonrpc broadcast", "POST", "/rpc/", `{"method":"broadcast_tx_commit","params":{"tx":"AA=="}}`, 200, `POST / {"method":"broadcast_tx_commit","params":{"tx":"AA=="}}`},
		{"jsonrpc unsafe", "POST", "/rpc", `{"method":"unsafe_write_heap_profile","params":{"filename":"/etc/x"}}`, 403, ""},
		{"jsonrpc batch", "POST", "/rpc", ` [{"method":"status"},{"method":"health"}]`, 200, `POST /  [{"method":"status"},{"method":"health"}]`},
		{"jsonrpc batch smuggling unsafe", "POST", "/rpc", `[{"method":"status"},{"method":"unsafe_flush_mempool"}]`, 403, ""},
		{"jsonrpc empty batch", "POST", "/rpc", `[]`, 400, ""},
		{"jsonrpc no method", "POST", "/rpc", `{}`, 403, ""},
		{"jsonrpc garbage", "POST", "/rpc", `not json`, 400, ""},
		{"jsonrpc too large", "POST", "/rpc", `{"method":"status","pad":"` + strings.Repeat("x", 2048) + `"}`, 413, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reached = nil
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.status, rec.Code, rec.Body.String())
			if tc.reach == "" {
				assert.Empty(t, reached, "a refused request reached the node")
			} else {
				assert.Equal(t, []string{tc.reach}, reached)
			}
		})
	}
}

func TestIsJSONRPCPost(t *testing.T) {
	cases := []struct {
		name        string
		method      string
		path        string
		contentType string
		want        bool
	}{
		{"gnokey", "POST", "/", "application/json", true},
		{"with charset", "POST", "/", "application/json; charset=utf-8", true},
		{"gnoweb form", "POST", "/", "application/x-www-form-urlencoded", false},
		{"realm action form", "POST", "/r/demo/foo", "application/x-www-form-urlencoded", false},
		{"json off the root", "POST", "/r/demo/foo", "application/json", false},
		{"get", "GET", "/", "application/json", false},
		{"no content type", "POST", "/", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			assert.Equal(t, tc.want, isJSONRPCPost(req))
		})
	}
}

// TestStagingDefaults pins what a staging node exposes. Each of these was
// once wrong in a way that looked fine from the logs.
func TestStagingDefaults(t *testing.T) {
	cfg := defaultStagingOptions

	assert.Equal(t, "0.0.0.0:8888", cfg.webListenerAddr,
		"staging must serve outside a container")
	assert.Equal(t, "127.0.0.1:26657", cfg.nodeRPCListenerAddr,
		"the node's RPC stays on the loopback; the web listener serves it, filtered")
	assert.False(t, cfg.unsafeAPI, "/reset and /reload must stay off")
	assert.NotEmpty(t, cfg.faucetAmount, "staging ships a faucet")
	assert.Empty(t, defaultLocalAppConfig.faucetAmount, "local mode has no faucet")
}
