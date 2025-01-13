package proxy_test

import (
	"net"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/contribs/gnodev/pkg/proxy"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

func generateTestinGenesisState(creator crypto.PrivKey, paths ...string) gnoland.GnoGenesisState {
	txs := make([]gnoland.TxWithMetadata, len(paths))
	// creator := privKey.PubKey().Address()
	for i, path := range paths {
		pkg := gnovm.MemPackage{
			Name: "foo",
			Path: path,
			Files: []*gnovm.MemFile{
				&gnovm.MemFile{
					Name: "foo.gno",
					Body: `package foo; func Render(_ string) string { return "bar" }`,
				},
				&gnovm.MemFile{Name: "gno.mod", Body: `module ` + path},
			},
		}

		// Create transaction
		var tx std.Tx
		tx.Fee = std.Fee{GasWanted: 1e6, GasFee: std.Coin{Amount: 1e6, Denom: "ugnot"}}
		tx.Msgs = []std.Msg{
			vmm.MsgAddPackage{
				Creator: creator.PubKey().Address(),
				Package: &pkg,
			},
		}

		tx.Signatures = make([]std.Signature, len(tx.GetSigners()))
		txs[i] = gnoland.TxWithMetadata{Tx: tx}
	}

	gnoland.SignGenesisTxs(txs, creator, "tendermint_test")

	return gnoland.GnoGenesisState{
		Txs: txs,
		Balances: []gnoland.Balance{{
			Address: creator.PubKey().Address(),
			Amount:  std.MustParseCoins(ugnot.ValueString(10_000_000_000_000)),
		}},
	}
}

func TestProxy(t *testing.T) {
	rootdir := gnoenv.RootDir()
	cfg := integration.TestingMinimalNodeConfig(rootdir)
	logger := log.NewTestingLogger(t)

	// Create temporary place for unix sock
	tmp := t.TempDir()
	sock := filepath.Join(tmp, "node.sock")
	addr, err := net.ResolveUnixAddr("unix", sock)
	require.NoError(t, err)

	// Create proxy
	path_interceptor, err := proxy.NewPathInterceptor(logger, addr)
	defer path_interceptor.Close()
	require.NoError(t, err)
	cfg.TMConfig.RPC.ListenAddress = path_interceptor.ProxyAddress()

	const targetPath = "gno.land/r/target/path"

	// Setup genesis state
	privKey := secp256k1.GenPrivKey()
	cfg.Genesis.AppState = generateTestinGenesisState(privKey, targetPath)

	integration.TestingInMemoryNode(t, logger, cfg)

	cc := make(chan []string, 1)
	path_interceptor.HandlePath(func(paths ...string) {
		cc <- paths
	})

	t.Run("http/valid_query", func(t *testing.T) {
		const qrender = "vm/qrender"

		cli, err := client.NewHTTPClient(path_interceptor.TargetAddress())
		require.NoError(t, err)
		defer cli.Close()

		res, err := cli.ABCIQuery(qrender, []byte(targetPath+":\n"))
		require.NoError(t, err)
		require.NoError(t, res.Response.Error)

		select {
		case paths := <-cc:
			require.Len(t, paths, 1)
			require.Equal(t, targetPath, paths[0])
		default:
			require.FailNow(t, "path should have been catch")
		}
	})

	t.Run("http/invalid_query", func(t *testing.T) {
		const wrongQuery = "does_not_exist_query"

		cli, err := client.NewHTTPClient(path_interceptor.TargetAddress())
		require.NoError(t, err)
		defer cli.Close()

		res, err := cli.ABCIQuery(wrongQuery, []byte(targetPath+":\n"))
		require.NoError(t, err)
		require.Error(t, res.Response.Error)

		select {
		case <-cc:
			require.FailNow(t, "should not catch a path")
		default:
		}
	})

	t.Run("ws/not_supported", func(t *testing.T) {
		// XXX:
	})

	err = path_interceptor.Close()
	require.NoError(t, err)
}
