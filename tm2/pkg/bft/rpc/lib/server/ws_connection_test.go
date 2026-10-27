package rpcserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/service"
)

// TestReadRoutineReturnsWhenStoppedMidStream checks that readRoutine returns
// once the connection stops, even while it is handing a streaming result to a
// writer that has already exited.
func TestReadRoutineReturnsWhenStoppedMidStream(t *testing.T) {
	t.Parallel()

	called := make(chan struct{})
	funcMap := map[string]*RPCFunc{
		"stream": NewWSRPCFunc(func(*types.Context) (*streamingResult, error) {
			close(called)
			return &streamingResult{}, nil
		}, ""),
	}

	serverConns := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err == nil {
			serverConns <- c
		}
	}))
	defer srv.Close()

	client, _, err := websocket.DefaultDialer.Dial("ws://"+srv.Listener.Addr().String(), nil)
	require.NoError(t, err)
	defer client.Close()

	wsc := NewWSConnection(<-serverConns, funcMap)
	// A service with no OnStart marks wsc running without a writeRoutine, so
	// streamChan stays nil and a send on it blocks, as after a failed write.
	wsc.BaseService = *service.NewBaseService(nil, "wsConnection", service.NewBaseService(nil, "noop", nil))
	require.NoError(t, wsc.Start())

	done := make(chan struct{})
	go func() {
		wsc.readRoutine()
		close(done)
	}()

	require.NoError(t, client.WriteMessage(websocket.TextMessage, []byte(`{"jsonrpc":"2.0","method":"stream","id":"1"}`)))
	select {
	case <-called:
	case <-time.After(5 * time.Second):
		t.Fatal("stream handler was never called")
	}
	require.NoError(t, wsc.Stop())

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("readRoutine still blocked on streamChan after the connection stopped")
	}
}
