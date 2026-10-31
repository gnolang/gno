package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	cfg "github.com/gnolang/gno/tm2/pkg/bft/rpc/config"
	rpcserver "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// recordingQueryConn answers every query and records the paths that reached it.
type recordingQueryConn struct {
	mu    sync.Mutex
	paths []string
}

func (c *recordingQueryConn) Error() error { return nil }

func (c *recordingQueryConn) EchoSync(msg string) (abci.ResponseEcho, error) {
	return abci.ResponseEcho{Message: msg}, nil
}

func (c *recordingQueryConn) InfoSync(abci.RequestInfo) (abci.ResponseInfo, error) {
	return abci.ResponseInfo{}, nil
}

func (c *recordingQueryConn) QuerySync(req abci.RequestQuery) (abci.ResponseQuery, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.paths = append(c.paths, req.Path)
	return abci.ResponseQuery{}, nil
}

func (c *recordingQueryConn) received() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.paths...)
}

// newCodeExecutionEnv returns an Environment whose app runs code on
// vm/qeval and vm/qeval_json, beside .app/simulate.
func newCodeExecutionEnv(allow bool) (*Environment, *recordingQueryConn) {
	conn := &recordingQueryConn{}
	rpcConfig := cfg.DefaultRPCConfig()
	rpcConfig.AllowCodeExecution = allow
	return &Environment{
		ProxyAppQuery:           conn,
		Logger:                  log.NewNoopLogger(),
		Config:                  *rpcConfig,
		CodeExecutionQueryPaths: []string{"vm/qeval", "vm/qeval_json"},
	}, conn
}

func TestABCIQuery_AllowCodeExecution(t *testing.T) {
	t.Parallel()

	runsCode := []string{
		".app/simulate",
		"/.app/simulate",
		".app/simulate/x",
		"vm/qeval",
		"/vm/qeval",
		"vm/qeval?x=1",
		"vm/qeval/x",
		"vm/qeval_json",
		"vm/qeval_json?x=1",
	}
	runsNoCode := []string{
		".app/version",
		".app/simulatex",
		"vm/qfile",
		"vm/qevalx",
		"vm/qeva",
		"vmx/qeval",
		"vm",
		"auth/accounts/g1x",
	}

	t.Run("true serves every path", func(t *testing.T) {
		t.Parallel()

		env, conn := newCodeExecutionEnv(true)
		all := slices.Concat(runsCode, runsNoCode)
		for _, path := range all {
			_, err := env.ABCIQuery(&rpctypes.Context{}, path, nil, 0, false)
			assert.NoError(t, err, "path %q", path)
		}
		assert.Equal(t, all, conn.received())
	})

	t.Run("false refuses the paths that run code", func(t *testing.T) {
		t.Parallel()

		env, conn := newCodeExecutionEnv(false)
		for _, path := range runsCode {
			_, err := env.ABCIQuery(&rpctypes.Context{}, path, nil, 0, false)
			assert.Equal(t, NewCodeExecutionRefusedError(path), err, "path %q", path)
		}
		for _, path := range runsNoCode {
			_, err := env.ABCIQuery(&rpctypes.Context{}, path, nil, 0, false)
			assert.NoError(t, err, "path %q", path)
		}
		assert.Equal(t, runsNoCode, conn.received())
	})
}

// TestABCIQuery_CodeExecutionOverEveryTransport sends a refused and a served
// path over a URI GET, one JSON-RPC batch and a WebSocket, through the node's
// own routes.
func TestABCIQuery_CodeExecutionOverEveryTransport(t *testing.T) {
	t.Parallel()

	const refusedMsg = "rpc.allow_code_execution = false"

	env, conn := newCodeExecutionEnv(false)
	routes := env.Routes(false)
	mux := http.NewServeMux()
	wm := rpcserver.NewWebsocketManager(routes)
	wm.SetLogger(log.NewNoopLogger())
	mux.HandleFunc("/websocket", wm.WebsocketHandler)
	rpcserver.RegisterRPCFuncs(mux, routes, log.NewNoopLogger())
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// URI GET.
	resp, err := http.Get(srv.URL + `/abci_query?path="vm/qeval"`)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	// One JSON-RPC batch holding a refused and a served call.
	batch := make([]rpctypes.RPCRequest, 0, 2)
	for i, path := range []string{".app/simulate", "vm/qfile"} {
		req, err := rpctypes.MapToRequest(rpctypes.JSONRPCIntID(i), "abci_query", map[string]any{"path": path})
		require.NoError(t, err)
		batch = append(batch, req)
	}
	body, err := json.Marshal(batch)
	require.NoError(t, err)
	resp, err = http.Post(srv.URL, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	var answers rpctypes.RPCResponses
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&answers))
	resp.Body.Close()
	require.Len(t, answers, 2)
	for _, answer := range answers {
		if answer.ID == rpctypes.JSONRPCIntID(0) {
			require.NotNil(t, answer.Error)
			assert.Contains(t, answer.Error.Data, refusedMsg)
		} else {
			assert.Nil(t, answer.Error)
		}
	}

	// WebSocket.
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/websocket", nil)
	require.NoError(t, err)
	defer ws.Close()
	req, err := rpctypes.MapToRequest(rpctypes.JSONRPCStringID("ws"), "abci_query", map[string]any{"path": "vm/qeval_json"})
	require.NoError(t, err)
	require.NoError(t, ws.WriteJSON(req))
	var wsAnswer rpctypes.RPCResponse
	require.NoError(t, ws.ReadJSON(&wsAnswer))
	require.NotNil(t, wsAnswer.Error)
	assert.Contains(t, wsAnswer.Error.Data, refusedMsg)

	assert.Equal(t, []string{"vm/qfile"}, conn.received())
}
