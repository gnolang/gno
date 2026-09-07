package main

import (
	"io"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// privateGnoMod is a gnomod.toml marking the package a PRIVATE realm.
//
// GenGnoModLatest ends its last line without a newline, so the marker supplies
// its own separator; gnomod.ParseMemPackage in the keeper is what has to accept
// the result.
func privateGnoMod(path string) string {
	return gno.GenGnoModLatest(path) + "\nprivate = true\n"
}

// TestRedeployParkedOverLivePrivateRealmIsEnabled pins the pre-flight to what
// the chain can actually hold at a path.
//
// AddPackage refuses to park over a live PUBLIC package but allows it for a
// PRIVATE one, so "live" and "has a submission pending" coexist. vm/qfile
// reads the live store only; a pre-flight built on it reports the parked
// redeploy as already active, records SUCCESS, and silently drops it. The
// chain ships vm/qpkgmeta_json for exactly this: its Pending field is true for
// a live private realm with a redeploy parked over it.
func TestRedeployParkedOverLivePrivateRealmIsEnabled(t *testing.T) {
	gnoroot := gnoenv.RootDir()
	cfg := integration.TestingMinimalNodeConfig(gnoroot)
	cfg.SkipGenesisSigVerification = true

	signer, err := gnoclient.SignerFromBip39(
		integration.DefaultAccount_Seed, cfg.Genesis.ChainID, "", 0, 0)
	require.NoError(t, err)
	info, err := signer.Info()
	require.NoError(t, err)
	who := info.GetAddress()

	// The chain runs "inert" and trusts this one key to approve.
	ggs := cfg.Genesis.AppState.(gnoland.GnoGenesisState)
	ggs.Balances = []gnoland.Balance{{
		Address: who,
		Amount:  std.NewCoins(std.NewCoin("ugnot", 100_000_000_000)),
	}}
	ggs.VM.Params.CodeSubmissionPolicy = "inert"
	ggs.VM.Params.PkgApprovers = []crypto.Address{who}
	cfg.Genesis.AppState = ggs

	node, remote := integration.TestingInMemoryNode(t, log.NewNoopLogger(), cfg)
	defer node.Stop()

	rpc, err := rpcclient.NewHTTPClient(remote)
	require.NoError(t, err)
	client := gnoclient.Client{Signer: signer, RPCClient: rpc}

	parkPrivate := func(t *testing.T, mpkg *std.MemPackage) {
		t.Helper()
		signed, err := client.SignTx(std.Tx{
			Msgs: []std.Msg{vm.MsgAddPackage{Creator: who, Package: mpkg}},
			Fee:  std.NewFee(20_000_000, std.MustParseCoin("1000000ugnot")),
		}, 0, 0)
		require.NoError(t, err)
		res, err := client.BroadcastTxCommit(signed)
		require.NoError(t, err)
		require.True(t, res.CheckTx.IsOK(), "park checkTx: %v", res.CheckTx.Error)
		require.True(t, res.DeliverTx.IsOK(), "park deliverTx: %v", res.DeliverTx.Error)
	}

	enableAsApprover := func(t *testing.T, mpkg *std.MemPackage) {
		t.Helper()
		signed, err := client.SignTx(std.Tx{
			Msgs: []std.Msg{vm.MsgEnablePackage{
				Approver: who,
				PkgPath:  mpkg.Path,
				PkgHash:  vm.PackageContentHash(mpkg),
			}},
			Fee: std.NewFee(20_000_000, std.MustParseCoin("1000000ugnot")),
		}, 0, 0)
		require.NoError(t, err)
		res, err := client.BroadcastTxCommit(signed)
		require.NoError(t, err)
		require.True(t, res.CheckTx.IsOK(), "enable checkTx: %v", res.CheckTx.Error)
		require.True(t, res.DeliverTx.IsOK(), "enable deliverTx: %v", res.DeliverTx.Error)
	}

	const pkgPath = "gno.land/r/test/privrealm"
	v1 := &std.MemPackage{
		Name: "privrealm",
		Path: pkgPath,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: privateGnoMod(pkgPath)},
			{Name: "privrealm.gno", Body: "package privrealm\n\nfunc V(cur realm) string { return \"v1\" }\n"},
		},
	}
	v2 := &std.MemPackage{
		Name: "privrealm",
		Path: pkgPath,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: privateGnoMod(pkgPath)},
			{Name: "privrealm.gno", Body: "package privrealm\n\nfunc V(cur realm) string { return \"v2\" }\n"},
		},
	}

	parkPrivate(t, v1)
	enableAsApprover(t, v1)
	parkPrivate(t, v2) // the chain permits this: private over private

	// Real writers, not NewTestIO's nil ones: the parent tees a child's stderr
	// through io.Err(), and a nil there is a crash the tests should surface
	// rather than dodge.
	tio := commands.NewTestIO()
	tio.SetOut(commands.WriteNopCloser(io.Discard))
	tio.SetErr(commands.WriteNopCloser(io.Discard))
	o, err := newOracle(config{
		remote:       remote,
		chainID:      cfg.Genesis.ChainID,
		mnemonic:     integration.DefaultAccount_Seed,
		gnoRoot:      gnoroot,
		gasFee:       defaultGasFee,
		gasWanted:    defaultGasWanted,
		verifyBudget: time.Minute,
	}, tio)
	require.NoError(t, err)
	o.blockMaxGas = o.queryBlockMaxGas(t.Context())

	o.handleCandidate(t.Context(), v2)

	st := o.status.get(pkgPath)
	require.NotEqual(t, "already active on-chain", st.Reason,
		"the redeploy was parked, not live; reporting success without enabling silently drops it")
	require.Equal(t, statusApproved, st.Status, "status board: %+v", st)

	// The proof that an enable was actually sent: the chain now serves v2.
	res, err := client.Query(gnoclient.QueryCfg{
		Path: "vm/qfile",
		Data: []byte(pkgPath + "/privrealm.gno"),
	})
	require.NoError(t, err)
	require.Nil(t, res.Response.Error)
	assert.Contains(t, string(res.Response.Data), "v2",
		"the live blob must be the redeploy, not the original")
}

// TestPkgMetaSettled covers the three answers vm/qpkgmeta_json can give and
// the two failure shapes, without a node.
func TestPkgMetaSettled(t *testing.T) {
	for _, tc := range []struct {
		name string
		data string
		want bool
	}{
		{"live with nothing pending: settled", `{"path":"p","status":"live"}`, true},
		{"live with a redeploy parked: NOT settled", `{"path":"p","status":"live","pending":true}`, false},
		{"inert: not settled", `{"path":"p","status":"inert","pending":true}`, false},
		{"absent: not settled", `{"path":"p","status":"absent"}`, false},
		{"garbage: not settled, so the oracle still tries", `{`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, pkgMetaSettled([]byte(tc.data)))
		})
	}
}
