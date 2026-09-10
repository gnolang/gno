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
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// inertChain is an in-memory node running the "inert" policy, trusting one
// funded key to approve, with a client signing as that key.
type inertChain struct {
	cfg    *gnoland.InMemoryNodeConfig
	remote string
	client gnoclient.Client
	who    crypto.Address
}

func newInertChain(t *testing.T) *inertChain {
	t.Helper()
	cfg := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
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
	ggs.VM.Params.CodeSubmissionPolicy = vm.CodeSubmissionPolicyInert
	ggs.VM.Params.PkgApprovers = []crypto.Address{who}
	cfg.Genesis.AppState = ggs

	node, remote := integration.TestingInMemoryNode(t, log.NewNoopLogger(), cfg)
	t.Cleanup(func() { node.Stop() })

	rpc, err := rpcclient.NewHTTPClient(remote)
	require.NoError(t, err)
	return &inertChain{
		cfg:    cfg,
		remote: remote,
		client: gnoclient.Client{Signer: signer, RPCClient: rpc},
		who:    who,
	}
}

// park submits mpkg, which the policy stores without enabling.
func (c *inertChain) park(t *testing.T, mpkg *std.MemPackage) {
	t.Helper()
	tx := std.Tx{
		Msgs: []std.Msg{vm.MsgAddPackage{Creator: c.who, Package: mpkg}},
		Fee:  std.NewFee(defaultGasWanted, std.MustParseCoin(defaultGasFee)),
	}
	signed, err := c.client.SignTx(tx, 0, 0)
	require.NoError(t, err)
	res, err := c.client.BroadcastTxCommit(signed)
	require.NoError(t, err)
	require.True(t, res.CheckTx.IsOK(), "park checkTx: %v", res.CheckTx.Error)
	require.True(t, res.DeliverTx.IsOK(), "park deliverTx: %v", res.DeliverTx.Error)
}

// oracle builds an oracle approving on the chain, with the gas ceiling settled
// the way run settles it before any work.
func (c *inertChain) oracle(t *testing.T) *oracle {
	t.Helper()
	// Quiet IO, but with real writers: NewTestIO leaves Err nil, and verify()
	// then skips the child-stderr teeing the daemon actually runs.
	tio := commands.NewTestIO()
	tio.SetOut(commands.WriteNopCloser(io.Discard))
	tio.SetErr(commands.WriteNopCloser(io.Discard))
	o, err := newOracle(config{
		remote:        c.remote,
		chainID:       c.cfg.Genesis.ChainID,
		mnemonic:      integration.DefaultAccount_Seed,
		gnoRoot:       gnoenv.RootDir(),
		gasFee:        defaultGasFee,
		gasWanted:     defaultGasWanted,
		verifyBudget:  time.Minute,
		prepareBudget: defaultPrepareBudget,
	}, tio)
	require.NoError(t, err)
	maxGas, answered := o.queryBlockMaxGas(t.Context())
	require.True(t, answered, "a zero ceiling clamps every gas figure to zero and the enable is refused")
	o.blockMaxGas = maxGas
	return o
}

// TestParkedDependencyIsNotAPermanentVerdict encodes one property: a dependency
// awaiting its own approval settles NOTHING about the package that imports it.
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
	chain := newInertChain(t)

	const depPath = "gno.land/p/test/parkeddep"
	dep := chainPackage(depPath, "package parkeddep\n\nfunc Answer() int { return 42 }\n")
	const appPath = "gno.land/r/test/wantsdep"
	app := chainPackage(appPath,
		"package wantsdep\n\nimport \"gno.land/p/test/parkeddep\"\n\nfunc N(cur realm) int { return parkeddep.Answer() }\n")
	// C is the control: B's shape, but its import was never submitted to this
	// chain at all. vm/qfile answers "not found" for a path that is merely
	// parked and for one that does not exist, so B and C reach the typecheck
	// with the same error text and differ only in a fact the chain holds.
	const missingPath = "gno.land/p/test/neversubmitted"
	const controlPath = "gno.land/r/test/wantsmissing"
	control := chainPackage(controlPath,
		"package wantsmissing\n\nimport \""+missingPath+"\"\n\nfunc N(cur realm) int { return neversubmitted.Answer() }\n")
	chain.park(t, dep) // parked, NOT enabled
	chain.park(t, app)
	chain.park(t, control)

	// Premise: A and B are both parked, awaiting approval. Pinned because the
	// question this test asks only exists in that state -- with A live, B
	// typechecks and is approved, which is a different run entirely. C's
	// import is pinned absent for the same reason: were it parked, C would
	// stop being a control and become a second copy of B.
	parked, err := chain.client.Query(gnoclient.QueryCfg{
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

	o := chain.oracle(t)

	// The control runs first, because it is what stops the fix from being a
	// string match on the typecheck error. C's import does not exist anywhere,
	// which IS the submitter's mistake: rejected, and settled for those bytes
	// until they change. Treating every unresolved import as pending would
	// satisfy B's claims below and break these, so a fix has to ask the chain
	// which paths are parked -- vm/qpkgmeta_json -- rather than read the error.
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

// TestLiveImportTheNodeWillNotServeIsNotAVerdict: a live dependency vm/qfile
// lists but cannot serve leaves the importer pending as the oracle's limit,
// named as such.
//
// A dotless LICENCE file makes such a package: ValidateMemPackage accepts the
// name, but std.SplitFilepath treats only LICENSE and README as files, so
// vm/qfile lists it and then answers "not found" for it. The getter records
// the dependency absent, vm/qpkgmeta_json says live, and one more fetch says
// the same. That is neither a verdict nor a parked import: the chain enables
// the importer as it is, and no approval of the dependency is pending, so
// "resubmit once the import is live" would wait on an event that has already
// happened.
func TestLiveImportTheNodeWillNotServeIsNotAVerdict(t *testing.T) {
	chain := newInertChain(t)
	const depPath = "gno.land/p/test/licenced"
	dep := chainPackage(depPath, "package licenced\n\nfunc Answer() int { return 42 }\n")
	dep.Files = append(dep.Files, &std.MemFile{Name: "LICENCE", Body: "do as you please\n"})
	dep.Sort()
	chain.park(t, dep)

	// The oracle enables the dependency itself, so it is live and the fixture
	// is a package the approver has already accepted.
	o := chain.oracle(t)
	o.handleCandidate(t.Context(), dep)
	depSt := o.status.get(depPath)
	require.Equal(t, statusApproved, depSt.Status,
		"premise: the dependency must be live -- recorded reason: %s", depSt.Reason)

	// Premise: the node lists the file and will not serve it.
	listed, err := chain.client.Query(gnoclient.QueryCfg{Path: "vm/qfile", Data: []byte(depPath)})
	require.NoError(t, err)
	require.Contains(t, string(listed.Response.Data), "LICENCE")
	_, err = chain.client.Query(gnoclient.QueryCfg{Path: "vm/qfile", Data: []byte(depPath + "/LICENCE")})
	require.Error(t, err, "premise: vm/qfile must refuse the file it listed, or this tests nothing")

	const appPath = "gno.land/r/test/wantslicenced"
	app := chainPackage(appPath,
		"package wantslicenced\n\nimport \""+depPath+"\"\n\nfunc N(cur realm) int { return licenced.Answer() }\n")
	chain.park(t, app)

	o.handleCandidate(t.Context(), app)

	st := o.status.get(appPath)
	require.Equal(t, statusPending, st.Status,
		"a dependency this oracle cannot fetch is its own limit, not a verdict -- recorded reason: %s", st.Reason)
	assert.Contains(t, st.Reason, depPath, "the reason must name the import")
	assert.Contains(t, st.Reason, "would not serve", "the reason must say the node is what failed")
	assert.NotContains(t, st.Reason, "parked", "the dependency is live; nothing about it is awaited")
	assert.NotContains(t, o.seen, candidateKey(app),
		"the bytes were never judged, so a restart or resubmission must get a fresh look")
}

// TestParkedImportDoesNotExcuseOtherErrors: a parked import makes a package
// pending only when it is the ONLY thing wrong with it.
//
// Every package here imports parked A, and every one has a second fault the
// submitter owns: a type error of its own, or an import that was never
// submitted anywhere. Those are verdicts, and A being parked must not defer
// them. The absent import comes in both sort orders around A's path, because
// the typecheck reports imports in path order and a classifier that looks at
// the first unresolved one alone passes one order and fails the other.
func TestParkedImportDoesNotExcuseOtherErrors(t *testing.T) {
	chain := newInertChain(t)
	const depPath = "gno.land/p/test/parkeddep"
	chain.park(t, chainPackage(depPath, "package parkeddep\n\nfunc Answer() int { return 42 }\n"))

	cases := map[string]struct {
		path, body string
		blames     string // what the rejection has to name
	}{
		"own type error": {
			path:   "gno.land/r/test/alsobroken",
			body:   "package alsobroken\n\nimport \"" + depPath + "\"\n\nfunc N(cur realm) int { return parkeddep.Answer() + undefinedThing }\n",
			blames: "undefinedThing",
		},
		"absent import sorting before the parked one": {
			path:   "gno.land/r/test/alsomissing",
			body:   "package alsomissing\n\nimport (\n\t\"gno.land/p/test/neversubmitted\"\n\t\"" + depPath + "\"\n)\n\nfunc N(cur realm) int { return parkeddep.Answer() + neversubmitted.Answer() }\n",
			blames: "gno.land/p/test/neversubmitted",
		},
		"absent import sorting after the parked one": {
			path:   "gno.land/r/test/alsomissinglater",
			body:   "package alsomissinglater\n\nimport (\n\t\"" + depPath + "\"\n\t\"gno.land/p/test/zzneversubmitted\"\n)\n\nfunc N(cur realm) int { return parkeddep.Answer() + zzneversubmitted.Answer() }\n",
			blames: "gno.land/p/test/zzneversubmitted",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mpkg := chainPackage(tc.path, tc.body)
			chain.park(t, mpkg)
			o := chain.oracle(t)

			o.handleCandidate(t.Context(), mpkg)

			st := o.status.get(tc.path)
			require.Equal(t, statusRejected, st.Status,
				"a fault the submitter owns is a verdict whatever else is parked -- recorded reason: %s", st.Reason)
			assert.Contains(t, st.Reason, tc.blames,
				"the rejection must name the fault the submitter can act on")
			assert.Contains(t, o.seen, candidateKey(mpkg),
				"a verdict about the bytes settles them until the submitter changes them")
		})
	}
}
