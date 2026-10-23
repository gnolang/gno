package main

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"sync/atomic"
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
		pkgHash, err := vm.PackageContentHash(mpkg)
		require.NoError(t, err)
		signed, err := client.SignTx(std.Tx{
			Msgs: []std.Msg{vm.MsgEnablePackage{
				Approver: who,
				PkgPath:  mpkg.Path,
				PkgHash:  pkgHash,
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
		remote:        remote,
		chainID:       cfg.Genesis.ChainID,
		mnemonic:      integration.DefaultAccount_Seed,
		gnoRoot:       gnoroot,
		gasFee:        defaultGasFee,
		gasWanted:     defaultGasWanted,
		verifyBudget:  time.Minute,
		prepareBudget: defaultPrepareBudget,
	}, tio)
	require.NoError(t, err)
	maxGas, answered := o.queryBlockMaxGas(t.Context())
	require.True(t, answered, "a zero ceiling clamps every gas figure to zero and the enable is refused")
	o.blockMaxGas = maxGas

	o.handleCandidate(t.Context(), candidate{mpkg: v2})

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

// TestSettledPackageIsNotVerified: a package live with nothing pending has
// nothing to enable, so it is recorded as already active without a verifier
// being spawned for it. A restart catching up with -start-height over blocks
// already approved hands the oracle every such package again, and each
// verification costs a child process plus a fetch of the import closure.
//
// The oracle and the child it spawns reach the node through a proxy counting
// vm/qfile queries: fetching sources sends those and the settled check never
// does, so one counted is a verification.
func TestSettledPackageIsNotVerified(t *testing.T) {
	chain := newInertChain(t)
	const depPath = "gno.land/p/test/livedep"
	dep := chainPackage(depPath, "package livedep\n\nfunc Answer() int { return 42 }\n")
	const appPath = "gno.land/r/test/useslivedep"
	app := chainPackage(appPath,
		"package useslivedep\n\nimport \""+depPath+"\"\n\nfunc N(cur realm) int { return livedep.Answer() }\n")

	var qfiles atomic.Int64
	counted := httptest.NewServer(countingQuery(t, chain.remote, "vm/qfile", &qfiles))
	t.Cleanup(counted.Close)
	chain.remote = counted.URL

	first := chain.oracle(t)
	chain.park(t, dep)
	first.handleCandidate(t.Context(), candidate{mpkg: dep})
	chain.park(t, app)
	qfiles.Store(0)
	first.handleCandidate(t.Context(), candidate{mpkg: app})
	firstSt := first.status.get(appPath)
	require.Equal(t, statusApproved, firstSt.Status,
		"premise: the first run must enable the package -- recorded reason: %s", firstSt.Reason)
	require.Positive(t, qfiles.Load(),
		"premise: verifying the package fetches its import, or the zero below proves nothing")

	// A restart: same bytes, fresh oracle, nothing in seen.
	replay := chain.oracle(t)
	qfiles.Store(0)
	replay.handleCandidate(t.Context(), candidate{mpkg: app})

	st := replay.status.get(appPath)
	require.Equal(t, statusApproved, st.Status, "recorded reason: %s", st.Reason)
	assert.Equal(t, "already active on-chain", st.Reason)
	assert.Contains(t, replay.seen, candidateKey(app),
		"a settled package is terminal for these bytes")
	assert.Zero(t, qfiles.Load(),
		"a settled package has nothing to enable, so no verifier may run for it")
}

// TestPackageEnabledDuringVerificationIsNotEnabledAgain: the settled check is
// asked again once verification is done, since the package may have gone live
// meanwhile, and an enable sent then only pays a fee to be refused.
func TestPackageEnabledDuringVerificationIsNotEnabledAgain(t *testing.T) {
	chain := newInertChain(t)
	const path = "gno.land/r/test/racing"
	pkg := chainPackage(path, "package racing\n\nfunc N(cur realm) int { return 1 }\n")
	chain.park(t, pkg)

	var metaQueries, broadcasts atomic.Int64
	enabled := make(chan error, 1)
	proxy := httptest.NewServer(rewriting(t, chain.remote, func(body []byte) []byte {
		if bytes.Contains(body, []byte(`"broadcast_tx`)) {
			broadcasts.Add(1)
		}
		// The first query is the check before verifying; the package goes live
		// just before the second one is answered.
		if bytes.Contains(body, []byte(`"path":"vm/qpkgmeta_json"`)) && metaQueries.Add(1) == 2 {
			enabled <- chain.enable(pkg)
		}
		return body
	}))
	t.Cleanup(proxy.Close)
	chain.remote = proxy.URL

	o := chain.oracle(t)
	o.handleCandidate(t.Context(), candidate{mpkg: pkg})

	require.Equal(t, int64(2), metaQueries.Load(), "premise: the package is checked before and after verifying")
	require.NoError(t, <-enabled)
	st := o.status.get(path)
	assert.Equal(t, statusApproved, st.Status, "recorded reason: %s", st.Reason)
	assert.Equal(t, "already active on-chain", st.Reason)
	assert.Zero(t, broadcasts.Load(), "an enable for a package already live only pays a fee")
}

// TestSettledCheckEndsInTime: the check runs before every verification, so a
// node that accepts the connection and never answers holds the verifier no
// longer than the caller's context or the check's own timeout, whichever ends
// first, never the RPC client's minute. handleCandidate passes the daemon's
// context, which has no deadline.
func TestSettledCheckEndsInTime(t *testing.T) {
	for name, tc := range map[string]struct {
		callerDeadline, settledTimeout time.Duration
	}{
		"the caller's context ends first":    {callerDeadline: 200 * time.Millisecond, settledTimeout: settledCheckTimeout},
		"the check's own timeout ends first": {settledTimeout: 200 * time.Millisecond},
	} {
		t.Run(name, func(t *testing.T) {
			o := newTestOracleAt(t, stalledNode(t))
			o.settledTimeout = tc.settledTimeout
			ctx := t.Context()
			if tc.callerDeadline > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tc.callerDeadline)
				defer cancel()
			}

			answered := make(chan bool, 1)
			go func() { answered <- o.isSettled(ctx, "gno.land/r/test/anything") }()

			select {
			case settled := <-answered:
				assert.False(t, settled, "a node that did not answer settles nothing")
			case <-time.After(5 * time.Second):
				t.Fatal("isSettled outlived both its context and its timeout")
			}
		})
	}
}

// TestPkgMetaSettled covers the three answers vm/qpkgmeta_json can give and
// the failure shapes, without a node.
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
		{"unknown status: not settled", `{"path":"p","status":"frozen"}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			meta, err := decodePkgMeta([]byte(tc.data))
			assert.Equal(t, tc.want, err == nil && metaSettled(meta))
		})
	}
}
