package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestBackup(t *testing.T) {
	mockedServer := mockWebsocketServer(t)
	defer mockedServer.Close()

	io := commands.NewTestIO()
	io.SetOut(os.Stdout)
	io.SetErr(os.Stderr)
	outDir := t.TempDir()
	url := strings.ReplaceAll(mockedServer.URL, "http:", "ws:")

	err := newRootCmd(io).ParseAndRun(context.Background(), []string{
		"--remote", url,
		"--o", outDir,
	})
	require.NoError(t, err)
}

// mockWebsocketServer will create a mock websocket server for testing
// this will only return 5 blocks and then done
func mockWebsocketServer(t *testing.T) *httptest.Server {
	t.Helper()

	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/websocket", r.URL.Path)

		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		defer conn.Close()

		_, requestMsg, err := conn.ReadMessage()
		require.NoError(t, err)

		var request rpctypes.RPCRequest
		require.NoError(t, json.Unmarshal(requestMsg, &request))
		require.Equal(t, "backup", request.Method)

		for height := int64(1); height <= 5; height++ {
			resp := rpctypes.NewRPCSuccessResponse(request.ID, &ctypes.ResultBackupBlock{
				Height: height,
				Block:  &types.Block{Header: types.Header{Height: height}},
			})

			respBz, err := json.Marshal(resp)
			require.NoError(t, err)
			require.NoError(t, conn.WriteMessage(websocket.TextMessage, respBz))
		}

		done := rpctypes.NewRPCSuccessResponse(request.ID, &ctypes.ResultBackupBlock{
			Done: true,
		})

		doneBz, err := json.Marshal(done)
		require.NoError(t, err)
		require.NoError(t, conn.WriteMessage(websocket.TextMessage, doneBz))
	}))

	return server
}
