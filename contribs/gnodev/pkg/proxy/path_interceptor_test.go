package proxy_test

import (
	"net"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/contribs/gnodev/pkg/proxy"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxy(t *testing.T) {
	const targetPath = "gno.land/r/target/foo"

	pkg := gnovm.MemPackage{
		Name: "foo",
		Path: targetPath,
		Files: []*gnovm.MemFile{
			{
				Name: "foo.gno",
				Body: `package foo; func Render(_ string) string { return "foo" }`,
			},
			{Name: "gno.mod", Body: `module ` + targetPath},
		},
	}

	rootdir := gnoenv.RootDir()
	cfg := integration.TestingMinimalNodeConfig(rootdir)
	logger := log.NewTestingLogger(t)

	tmp := t.TempDir()
	sock := filepath.Join(tmp, "node.sock")
	addr, err := net.ResolveUnixAddr("unix", sock)
	require.NoError(t, err)

	// Create proxy
	interceptor, err := proxy.NewPathInterceptor(logger, addr)
	require.NoError(t, err)
	defer interceptor.Close()
	cfg.TMConfig.RPC.ListenAddress = interceptor.ProxyAddress()
	cfg.SkipGenesisVerification = true

	// Setup genesis
	privKey := secp256k1.GenPrivKey()
	cfg.Genesis.AppState = integration.GenerateTestingGenesisState(privKey, pkg)
	creator := privKey.PubKey().Address()

	integration.TestingInMemoryNode(t, logger, cfg)
	pathChan := make(chan []string, 1)
	interceptor.HandlePath(func(paths ...string) {
		pathChan <- paths
	})

	// ---- Test Cases ----

	t.Run("valid_vm_query", func(t *testing.T) {
		cli, err := client.NewHTTPClient(interceptor.TargetAddress())
		require.NoError(t, err)

		res, err := cli.ABCIQuery("vm/qrender", []byte(targetPath+":\n"))
		require.NoError(t, err)
		assert.Nil(t, res.Response.Error)

		select {
		case paths := <-pathChan:
			require.Len(t, paths, 1)
			assert.Equal(t, []string{targetPath}, paths)
		default:
			t.Fatal("paths not captured")
		}
	})

	t.Run("valid_vm_query_file", func(t *testing.T) {
		cli, err := client.NewHTTPClient(interceptor.TargetAddress())
		require.NoError(t, err)

		res, err := cli.ABCIQuery("vm/qfile", []byte(filepath.Join(targetPath, "foo.gno")))
		require.NoError(t, err)
		assert.Nil(t, res.Response.Error)

		select {
		case paths := <-pathChan:
			require.Len(t, paths, 1)
			assert.Equal(t, []string{targetPath}, paths)
		default:
			t.Fatal("paths not captured")
		}
	})

	t.Run("simulate_tx_paths", func(t *testing.T) {
		// Build transaction with multiple messages
		var tx std.Tx
		send := std.MustParseCoins(ugnot.ValueString(10_000_000))
		tx.Fee = std.Fee{GasWanted: 1e6, GasFee: std.Coin{Amount: 1e6, Denom: "ugnot"}}
		tx.Msgs = []std.Msg{
			vm.NewMsgCall(creator, send, targetPath, "Render", []string{""}),
			vm.NewMsgCall(creator, send, targetPath, "Render", []string{""}),
			vm.NewMsgCall(creator, send, targetPath, "Render", []string{""}),
		}

		bytes, err := tx.GetSignBytes(cfg.Genesis.ChainID, 0, 0)
		require.NoError(t, err)
		signature, err := privKey.Sign(bytes)
		require.NoError(t, err)
		tx.Signatures = []std.Signature{{PubKey: privKey.PubKey(), Signature: signature}}

		bz, err := amino.Marshal(tx)
		require.NoError(t, err)

		cli, err := client.NewHTTPClient(interceptor.TargetAddress())
		require.NoError(t, err)

		res, err := cli.BroadcastTxCommit(bz)
		require.NoError(t, err)
		assert.NoError(t, res.CheckTx.Error)
		assert.NoError(t, res.DeliverTx.Error)

		select {
		case paths := <-pathChan:
			require.Len(t, paths, 1)
			assert.Equal(t, []string{targetPath}, paths)
		default:
			t.Fatal("paths not captured")
		}
	})

	t.Run("websocket_forward", func(t *testing.T) {
		// For now simply try to connect and upgrade the connection
		// XXX: fully support ws

		conn, err := net.Dial(addr.Network(), addr.String())
		require.NoError(t, err)
		defer conn.Close()

		// Send WebSocket handshake
		req, _ := http.NewRequest("GET", "http://"+interceptor.TargetAddress(), nil)
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Connection", "Upgrade")
		err = req.Write(conn)
		require.NoError(t, err)
	})

	t.Run("invalid_query_data", func(t *testing.T) {
		// Making a valid call but not supported by the proxy
		// should succeed
		query := "auth/accounts/" + creator.String()

		cli, err := client.NewHTTPClient(interceptor.TargetAddress())
		require.NoError(t, err)
		defer cli.Close()

		res, err := cli.ABCIQuery(query, []byte{})
		require.NoError(t, err)
		require.NoError(t, res.Response.Error)

		var qret struct{ BaseAccount std.BaseAccount }
		err = amino.UnmarshalJSON(res.Response.Data, &qret)
		require.NoError(t, err)
		assert.Equal(t, qret.BaseAccount.Address, creator)

		select {
		case paths := <-pathChan:
			require.FailNowf(t, "should not catch a path", "catched: %+v", paths)
		default:
		}
	})
}
