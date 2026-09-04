package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// TestFetchingSourcesDoesNotSpendTheBudget: the budget measures the compile,
// which is the cost the validator pays at enable. Fetching the import closure
// from the node is the oracle's own cost, and a slow node must not turn a
// package that compiles in milliseconds into an overrun.
func TestFetchingSourcesDoesNotSpendTheBudget(t *testing.T) {
	const dep = "gno.land/p/nt/ufmt/v0"
	remote := nodeServing(t, dep)

	// Every RPC round trip pays this. The closure is one file listing plus
	// one query per file, so fetching it costs several times the budget below.
	const perRequest = 800 * time.Millisecond
	slow := httptest.NewServer(delaying(t, remote, perRequest))
	t.Cleanup(slow.Close)

	o := newTestOracle(t)
	o.cfg.remote = slow.URL
	o.cfg.verifyBudget = 2 * time.Second

	start := time.Now()
	err := o.verify(context.Background(), packageImporting(dep))
	elapsed := time.Since(start)

	require.Greater(t, elapsed, o.cfg.verifyBudget,
		"premise: fetching the closure must take longer than the budget, or this test proves nothing")
	require.NoError(t, err,
		"the package compiles in milliseconds; the time went to fetching its dependency from a slow node, which is not what the budget measures")
}

// nodeServing starts a node with the named examples package live on it, and
// returns its RPC address.
func nodeServing(t *testing.T, pkgPath string) string {
	t.Helper()
	gnoroot := gnoenv.RootDir()
	cfg := integration.TestingMinimalNodeConfig(gnoroot)
	cfg.SkipGenesisSigVerification = true

	signer, err := gnoclient.SignerFromBip39(
		integration.DefaultAccount_Seed, cfg.Genesis.ChainID, "", 0, 0)
	require.NoError(t, err)
	info, err := signer.Info()
	require.NoError(t, err)
	who := info.GetAddress()

	ggs := cfg.Genesis.AppState.(gnoland.GnoGenesisState)
	ggs.Balances = []gnoland.Balance{{
		Address: who,
		Amount:  std.NewCoins(std.NewCoin("ugnot", 100_000_000_000)),
	}}
	cfg.Genesis.AppState = ggs

	node, remote := integration.TestingInMemoryNode(t, log.NewNoopLogger(), cfg)
	t.Cleanup(func() { node.Stop() })

	rpc, err := rpcclient.NewHTTPClient(remote)
	require.NoError(t, err)
	client := gnoclient.Client{Signer: signer, RPCClient: rpc}

	mpkg, err := gno.ReadMemPackage(filepath.Join(gnoroot, "examples", pkgPath), pkgPath, gno.MPUserAll)
	require.NoError(t, err)
	signed, err := client.SignTx(std.Tx{
		Msgs: []std.Msg{vm.MsgAddPackage{Creator: who, Package: mpkg}},
		Fee:  std.NewFee(50_000_000, std.MustParseCoin("1000000ugnot")),
	}, 0, 0)
	require.NoError(t, err)
	res, err := client.BroadcastTxCommit(signed)
	require.NoError(t, err)
	require.True(t, res.CheckTx.IsOK(), "deploy %s: %v", pkgPath, res.CheckTx.Error)
	require.True(t, res.DeliverTx.IsOK(), "deploy %s: %v", pkgPath, res.DeliverTx.Error)
	return remote
}

// delaying proxies every request to remote after sleeping d: a node that
// answers, slowly.
func delaying(t *testing.T, remote string, d time.Duration) http.Handler {
	t.Helper()
	// The node advertises tcp://; the proxy speaks HTTP to it.
	target, err := url.Parse(strings.Replace(remote, "tcp://", "http://", 1))
	require.NoError(t, err)
	proxy := httputil.NewSingleHostReverseProxy(target)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(d):
			proxy.ServeHTTP(w, r)
		case <-r.Context().Done():
		}
	})
}

// packageImporting is a realm whose only chain import is dep.
func packageImporting(dep string) *std.MemPackage {
	const path = "gno.land/r/test/needsdep"
	return &std.MemPackage{
		Name: "needsdep",
		Path: path,
		Type: gno.MPUserAll,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(path)},
			{Name: "needsdep.gno", Body: "package needsdep\n\nimport \"" + dep + "\"\n\nfunc F(cur realm) string { return ufmt.Sprintf(\"%d\", 1) }\n"},
		},
	}
}
