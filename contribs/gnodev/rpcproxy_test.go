package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewRPCProxyAddress covers the address shapes the node actually reports.
// GetRemoteAddress returns a tm2 listen address ("tcp://host:port"), which no
// HTTP client accepts, and staging binds 0.0.0.0 by default, which is not a
// portable dial target.
func TestNewRPCProxyAddress(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		remote  string
		wantErr bool
	}{
		{name: "tm2 tcp listen address", remote: "tcp://127.0.0.1:26657"},
		{name: "wildcard bind is dialled on loopback", remote: "tcp://0.0.0.0:26657"},
		{name: "bare host:port", remote: "127.0.0.1:26657"},
		{name: "http url", remote: "http://127.0.0.1:26657"},
		{name: "empty is rejected", remote: "", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h, err := newRPCProxy(tc.remote)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, h)
		})
	}
}

// TestNewRPCProxyForwards is the behaviour that matters: /rpc reaches the node
// as /, and /rpc/status as /status, so a client pointed at the web origin
// speaks to the chain unchanged.
func TestNewRPCProxyForwards(t *testing.T) {
	t.Parallel()

	var gotPath string
	node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = io.WriteString(w, "node-reached")
	}))
	t.Cleanup(node.Close)

	u, err := url.Parse(node.URL)
	require.NoError(t, err)

	handler, err := newRPCProxy(u.Host)
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.Handle("/rpc", handler)
	mux.Handle("/rpc/", handler)

	for _, tc := range []struct{ request, wantUpstream string }{
		{"/rpc", "/"},
		{"/rpc/", "/"},
		{"/rpc/status", "/status"},
		{"/rpc/abci_query", "/abci_query"},
	} {
		res := httptest.NewRecorder()
		mux.ServeHTTP(res, httptest.NewRequest(http.MethodGet, tc.request, nil))

		assert.Equal(t, http.StatusOK, res.Code, "request %s", tc.request)
		assert.Equal(t, "node-reached", res.Body.String(), "request %s", tc.request)
		assert.Equal(t, tc.wantUpstream, gotPath,
			"%s must reach the node as %s, not with the prefix still attached", tc.request, gotPath)
	}
}

// TestRPCTargetURL asserts the normalisation directly. An earlier version of
// this test dialled a 0.0.0.0 address and asserted the request succeeded, which
// passed with the rewrite removed: Linux happily routes 0.0.0.0 to loopback, so
// the test could not fail and proved nothing.
func TestRPCTargetURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		remote string
		want   string
	}{
		{"tm2 tcp listen address", "tcp://127.0.0.1:26657", "http://127.0.0.1:26657"},
		{"wildcard is dialled on loopback", "tcp://0.0.0.0:26657", "http://127.0.0.1:26657"},
		{"bare host:port", "127.0.0.1:26657", "http://127.0.0.1:26657"},
		{"http is already a url", "http://127.0.0.1:26657", "http://127.0.0.1:26657"},
		{"https is dialled plain, the proxy hop is local", "https://127.0.0.1:26657", "http://127.0.0.1:26657"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := rpcTargetURL(tc.remote)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got.String())
		})
	}
}

func TestRPCTargetURLRejectsEmpty(t *testing.T) {
	t.Parallel()

	_, err := rpcTargetURL("")
	require.Error(t, err)
}
