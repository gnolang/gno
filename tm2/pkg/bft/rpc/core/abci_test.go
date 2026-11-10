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
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	cfg "github.com/gnolang/gno/tm2/pkg/bft/rpc/config"
	rpcserver "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// recordingQueryConn answers every query and records the paths that reached it.
type recordingQueryConn struct {
	appconn.Query // only QuerySync is called
	mu            sync.Mutex
	paths         []string
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

// codeRunningPaths are the paths newCodeExecutionEnv's app runs code on.
var codeRunningPaths = []string{".app/simulate", "vm/qeval", "vm/qeval_json"}

// newCodeExecutionEnv returns an Environment whose app runs code on
// codeRunningPaths.
func newCodeExecutionEnv(allow bool) (*Environment, *recordingQueryConn) {
	conn := &recordingQueryConn{}
	rpcConfig := cfg.DefaultRPCConfig()
	rpcConfig.AllowCodeExecution = allow
	return &Environment{
		ProxyAppQuery: conn,
		Logger:        log.NewNoopLogger(),
		Config:        *rpcConfig,
		QueryRunsCode: func(path string) bool { return slices.Contains(codeRunningPaths, path) },
	}, conn
}

func TestABCIQuery_AllowCodeExecution(t *testing.T) {
	t.Parallel()

	runsCode := codeRunningPaths
	runsNoCode := []string{".app/version", "vm/qfile", "auth/accounts/g1x"}

	for name, setup := range map[string]func(*Environment){
		"true serves every path": func(*Environment) {},
		"false serves every path for an app that cannot tell": func(env *Environment) {
			env.Config.AllowCodeExecution = false
			env.QueryRunsCode = nil
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			env, conn := newCodeExecutionEnv(true)
			setup(env)
			all := slices.Concat(runsCode, runsNoCode)
			for _, path := range all {
				_, err := env.ABCIQuery(&rpctypes.Context{}, path, nil, 0, false)
				assert.NoError(t, err, "path %q", path)
			}
			assert.Equal(t, all, conn.received())
		})
	}

	t.Run("false refuses the paths that run code", func(t *testing.T) {
		t.Parallel()

		env, conn := newCodeExecutionEnv(false)
		for _, path := range runsCode {
			_, err := env.ABCIQuery(&rpctypes.Context{}, path, nil, 0, false)
			assert.Equal(t, codeExecutionRefusedError(path), err, "path %q", path)
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
	require.Len(t, answers, 2) // answered in request order
	require.NotNil(t, answers[0].Error)
	assert.Contains(t, answers[0].Error.Data, refusedMsg)
	assert.Nil(t, answers[1].Error)

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
