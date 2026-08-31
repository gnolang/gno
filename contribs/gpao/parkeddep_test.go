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

// TestParkedDependencyIsNotAPermanentVerdict encodes one property: a dependency
// awaiting its own approval settles NOTHING about the package that imports it.
//
// This test currently fails; delete this line with the change that fixes it.
//
// B imports A. A is on the chain, parked, awaiting its own approval -- the
// composable case the inert policy exists to allow (verifier.go's preprocess
// stage says so in as many words, and tolerates exactly this input). But the
// typecheck stage resolves imports through vm/qfile, which cannot see a parked
// package, so B is rejected as bad code and its content hash marked seen:
// resubmitting identical bytes is a no-op for the lifetime of the process, and
// a restart re-verifies the same bytes and rejects them again. Submission order
// alone decides whether a valid closure can ever deploy.
//
// The property is stated without a mechanism on purpose. vm/qinertpaths could
// tell an oracle the path is parked; the verdict could stop being recorded as
// seen; something else again. Which one is right is left open.
func TestParkedDependencyIsNotAPermanentVerdict(t *testing.T) {
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
	ggs.VM.Params.CodeSubmissionPolicy = vm.CodeSubmissionPolicyInert
	ggs.VM.Params.PkgApprovers = []crypto.Address{who}
	cfg.Genesis.AppState = ggs

	node, remote := integration.TestingInMemoryNode(t, log.NewNoopLogger(), cfg)
	defer node.Stop()

	rpc, err := rpcclient.NewHTTPClient(remote)
	require.NoError(t, err)
	client := gnoclient.Client{Signer: signer, RPCClient: rpc}

	park := func(t *testing.T, mpkg *std.MemPackage) {
		t.Helper()
		tx := std.Tx{
			Msgs: []std.Msg{vm.MsgAddPackage{Creator: who, Package: mpkg}},
			Fee:  std.NewFee(defaultGasWanted, std.MustParseCoin(defaultGasFee)),
		}
		signed, err := client.SignTx(tx, 0, 0)
		require.NoError(t, err)
		res, err := client.BroadcastTxCommit(signed)
		require.NoError(t, err)
		require.True(t, res.CheckTx.IsOK(), "park checkTx: %v", res.CheckTx.Error)
		require.True(t, res.DeliverTx.IsOK(), "park deliverTx: %v", res.DeliverTx.Error)
	}

	const depPath = "gno.land/p/test/parkeddep"
	dep := &std.MemPackage{
		Name: "parkeddep",
		Path: depPath,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(depPath)},
			{Name: "parkeddep.gno", Body: "package parkeddep\n\nfunc Answer() int { return 42 }\n"},
		},
	}
	const appPath = "gno.land/r/test/wantsdep"
	app := &std.MemPackage{
		Name: "wantsdep",
		Path: appPath,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(appPath)},
			{Name: "wantsdep.gno", Body: "package wantsdep\n\nimport \"gno.land/p/test/parkeddep\"\n\nfunc N(cur realm) int { return parkeddep.Answer() }\n"},
		},
	}
	// C is the control: B's shape, but its import was never submitted to this
	// chain at all. vm/qfile answers "not found" for a path that is merely
	// parked and for one that does not exist, so B and C reach the typecheck
	// with the same error text and differ only in a fact the chain holds.
	const missingPath = "gno.land/p/test/neversubmitted"
	const controlPath = "gno.land/r/test/wantsmissing"
	control := &std.MemPackage{
		Name: "wantsmissing",
		Path: controlPath,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(controlPath)},
			{Name: "wantsmissing.gno", Body: "package wantsmissing\n\nimport \"" + missingPath + "\"\n\nfunc N(cur realm) int { return neversubmitted.Answer() }\n"},
		},
	}
	park(t, dep) // parked, NOT enabled
	park(t, app)
	park(t, control)

	// Premise: A and B are both parked, awaiting approval. Pinned because the
	// question this test asks only exists in that state -- with A live, B
	// typechecks and is approved, which is a different run entirely. C's
	// import is pinned absent for the same reason: were it parked, C would
	// stop being a control and become a second copy of B.
	parked, err := client.Query(gnoclient.QueryCfg{
		Path: "vm/qinertpaths",
		Data: []byte("gno.land/"),
	})
	require.NoError(t, err)
	require.Nil(t, parked.Response.Error)
	require.Contains(t, string(parked.Response.Data), depPath,
		"premise: A must be awaiting approval, not live")
	require.Contains(t, string(parked.Response.Data), appPath,
		"premise: B must be awaiting approval, not live")
	require.NotContains(t, string(parked.Response.Data), missingPath,
		"premise: C's import must be absent from the chain, not merely parked")

	// Quiet IO, but with real writers: NewTestIO leaves Err nil, and verify()
	// then skips the child-stderr teeing the daemon actually runs.
	tio := commands.NewTestIO()
	tio.SetOut(commands.WriteNopCloser(io.Discard))
	tio.SetErr(commands.WriteNopCloser(io.Discard))
	ocfg := config{
		remote:       remote,
		chainID:      cfg.Genesis.ChainID,
		mnemonic:     integration.DefaultAccount_Seed,
		gnoRoot:      gnoroot,
		gasFee:       defaultGasFee,
		gasWanted:    defaultGasWanted,
		verifyBudget: time.Minute,
	}
	o, err := newOracle(ocfg, tio)
	require.NoError(t, err)
	o.blockMaxGas = o.queryBlockMaxGas(t.Context())

	// The control runs first, because it is what stops the fix from being a
	// string match on the typecheck error. C's import does not exist anywhere,
	// which IS the submitter's mistake: rejected, and settled for those bytes
	// until they change. Treating every unresolved import as pending would
	// satisfy B's claims below and break these, so a fix has to ask the chain
	// which paths are parked -- vm/qinertpaths -- rather than read the error.
	o.handleCandidate(t.Context(), control)

	controlSt := o.status.get(controlPath)
	require.Equal(t, statusRejected, controlSt.Status,
		"an import that exists nowhere is the submitter's mistake to fix -- recorded reason: %s", controlSt.Reason)
	require.Contains(t, controlSt.Reason, missingPath,
		"the rejection must be about the absent import, not about C's own code")
	require.Contains(t, o.seen, candidateKey(control),
		"re-verifying these bytes reaches the same answer, so they are settled until the submitter changes them")

	o.handleCandidate(t.Context(), app)

	// Guards first. Both claims below are satisfied by an oracle that did
	// nothing at all -- `seen` is empty and status.get returns statusUnknown
	// for a path it never processed -- and by one that failed for reasons this
	// test is not asking about. They pin the run to the question actually
	// posed, so the eventual green cannot be vacuous. None of them says WHICH
	// non-rejected verdict is right: that is left open.
	st := o.status.get(appPath)
	require.NotEqual(t, statusUnknown, st.Status,
		"the oracle must have processed B for the assertions below to mean anything")
	require.NotContains(t, st.Reason, "could not run verification",
		"the verifier must have run for the assertions below to mean anything")
	require.NotContains(t, st.Reason, "ran out of time",
		"a budget overrun is not an answer to the question this test asks")
	if st.Status == statusRejected {
		require.Contains(t, st.Reason, depPath,
			"the rejection must be about the parked dependency, not about B's own code")
	}

	// The two claims, in order of consequence.
	assert.NotContains(t, o.seen, candidateKey(app),
		"B's bytes are fine and unchanged; a dependency awaiting approval must not settle them for the rest of the run")
	assert.NotEqual(t, statusRejected, st.Status,
		"'rejected' is reserved for code the submitter can act on; this is the oracle's queue order, "+
			"not their fault -- recorded reason: %s", st.Reason)
}
