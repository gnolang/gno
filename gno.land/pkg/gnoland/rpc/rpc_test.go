package rpc

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_RegisterHandlers(t *testing.T) {
	s := &Server{
		app:    &mockApplication{},
		logger: slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
	}

	funcs := s.rpcFuncs()

	expected := []string{
		"vm/render",
		"vm/funcs",
		"vm/eval",
		"vm/file",
		"vm/doc",
		"vm/paths",
		"vm/storage",
	}

	require.Len(t, funcs, len(expected))

	for _, key := range expected {
		assert.Contains(t, funcs, key)
	}
}

func TestServer_WebsocketReachable(t *testing.T) {
	t.Parallel()

	s := &Server{
		app:    &mockApplication{},
		logger: slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})),
	}

	mux := s.newMux()

	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/websocket")
	require.NoError(t, err)

	defer resp.Body.Close()

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
