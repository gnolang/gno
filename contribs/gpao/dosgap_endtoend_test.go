package main

import (
	"fmt"
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
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fanOutSrc builds a value-containment "doubling" chain: each level references the
// previous one twice BY VALUE, which go/types' validType walk expands without
// memoizing (the optimization is commented out as a workaround for
// golang/go#65711). Depth 24 is 234,880,917 node visits, about 8.5s of validType.
//
// Deliberately above what a block can pay for once the walk is priced:
// 2.35e8 nodes x 100 gas/node = 2.35e10 gas against a Block.MaxGas of 3e9.
func fanOutSrc(pkgName string, depth int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\ntype t0 struct{ v int }\n", pkgName)
	for i := 1; i <= depth; i++ {
		fmt.Fprintf(&b, "type t%d struct{ a, b [0]t%d }\n", i, i-1)
	}
	return b.String()
}

const dosDepth = 24

// describeBroadcast flattens the two shapes a rejected transaction arrives in:
// BroadcastTxCommit returns an error for some failures and a non-OK DeliverTx for
// others, and out-of-gas can present either way.
func describeBroadcast(res *ctypes.ResultBroadcastTxCommit, err error) string {
	if err != nil {
		return fmt.Sprintf("broadcast error: %v", err)
	}
	if res == nil {
		return "no result"
	}
	if !res.CheckTx.IsOK() {
		return fmt.Sprintf("checkTx failed: %v", res.CheckTx.Error)
	}
	if !res.DeliverTx.IsOK() {
		return fmt.Sprintf("deliverTx failed after %d gas: %v",
			res.DeliverTx.GasUsed, res.DeliverTx.Error)
	}
	return fmt.Sprintf("SUCCEEDED using %d gas", res.DeliverTx.GasUsed)
}

// TestDoSGapAgainstARealChain is the end-to-end demonstration: a real in-memory
// node, real signed transactions over RPC, the "inert" submission policy, an
// approver, and gpao's own admission procedure — driven against a package whose
// type-check is an unmetered 8.5-second walk.
//
// It exists because every cheaper test skips the part that matters. A keeper-level
// test calls EnablePackage directly, so it never exercises the ante handler, the
// simulate path, or gpao's decision. Here all three run.
//
// Three findings, one per subtest. On a tree where the walk is priced all three
// pass; on one where it is not, the last two fail BY SUCCEEDING, slowly — which is
// the gap.
//
//  1. Submitting the package is cheap and fast: under "inert" AddPackage parks the
//     bytes without type-checking them. #6088 working as designed.
//
//  2. gpao's own probe does not protect the chain — it participates. gpao decides
//     by simulating the enable at the chain's Block.MaxGas (classifySimulate over
//     SimulateResult). That simulate executes the walk ON THE NODE. Unpriced it
//     returns verdictReady after ~8.5s, so gpao approves and broadcasts, and the
//     chain then spends another ~8.5s applying it. Note also what the simulate
//     itself is: an unauthenticated abci_query that occupies a validator for 8.5s
//     with no fee and no transaction. Priced, the same probe returns
//     verdictWillFail and gpao correctly declines.
//
// It proves the work is UNBOUNDED AND UNPAID, reachable by an unprivileged party,
// on a chain configured exactly as #6088 intends. It does NOT prove the node halts,
// and nothing here should be described that way: every stall observed here ends and
// the node recovers. A halt cannot be asserted, only observed — a test that proved
// the walk never finishes would itself never finish — so the closest a test can get
// is "still running when the client gave up", which is what the composed subtest
// reports at its 10s RPC deadline. For an actual halt, raise dosDepth (40 is ~2^41
// node visits, hours) and watch a node produce nothing; that is a manual demo with
// no pass condition, not a test.
//
// Two further gaps, deliberately not closed:
//
//   - the gpao DAEMON is not run. Its decision functions (classifySimulate,
//     queryBlockMaxGas) are exercised against a real node, and the daemon
//     broadcasts on that verdict by construction, so running the process would add
//     fidelity but no finding — at the cost of a block reader, a verifier goroutine
//     and retry logic in a test that currently has none.
//
//   - single node. Every validator executes the same DeliverTx, so this generalises
//     trivially. Multi-node testing earns its cost where results DIFFER between
//     nodes, which is a determinism question, not this one.
//
//     3. MsgRun bypasses the flow entirely. It type-checks submitted source
//     immediately "under every policy including inert" (see the RunSubmitters doc
//     in vm/params.go), its allowlist is empty by default so anyone may send it,
//     and it is the only code-bearing message with no namespace or CLA gate. A
//     second key that is neither the approver nor a listed submitter reaches the
//     same walk on this locked-down chain.
func TestDoSGapAgainstARealChain(t *testing.T) {
	const parkedPath = "gno.land/r/test/parked"

	gnoroot := gnoenv.RootDir()
	cfg := integration.TestingMinimalNodeConfig(gnoroot)
	cfg.SkipGenesisSigVerification = true

	signer, err := gnoclient.SignerFromBip39(
		integration.DefaultAccount_Seed, cfg.Genesis.ChainID, "", 0, 0)
	require.NoError(t, err)
	info, err := signer.Info()
	require.NoError(t, err)
	approver := info.GetAddress()

	// A second key: neither approver nor a listed code submitter.
	strangerSigner, err := gnoclient.SignerFromBip39(
		integration.DefaultAccount_Seed, cfg.Genesis.ChainID, "", 0, 1)
	require.NoError(t, err)
	strangerInfo, err := strangerSigner.Info()
	require.NoError(t, err)
	stranger := strangerInfo.GetAddress()
	require.NotEqual(t, approver, stranger)

	// A third key, so the composition subtest does not share a sequence with the
	// single-message one (whose transaction may or may not be included when the
	// node is wedged, leaving the sequence ambiguous).
	packerSigner, err := gnoclient.SignerFromBip39(
		integration.DefaultAccount_Seed, cfg.Genesis.ChainID, "", 0, 2)
	require.NoError(t, err)
	packerInfo, err := packerSigner.Info()
	require.NoError(t, err)
	packer := packerInfo.GetAddress()

	// The chain exactly as #6088 intends it: submissions park, one approver holds
	// the key, and MsgRun is left at its shipped default.
	ggs := cfg.Genesis.AppState.(gnoland.GnoGenesisState)
	ggs.Balances = []gnoland.Balance{
		{Address: approver, Amount: std.NewCoins(std.NewCoin("ugnot", 100_000_000_000))},
		{Address: stranger, Amount: std.NewCoins(std.NewCoin("ugnot", 100_000_000_000))},
		{Address: packer, Amount: std.NewCoins(std.NewCoin("ugnot", 100_000_000_000))},
	}
	ggs.VM.Params.CodeSubmissionPolicy = "inert"
	ggs.VM.Params.PkgApprovers = []crypto.Address{approver}
	require.Empty(t, ggs.VM.Params.RunSubmitters,
		"this demonstration relies on the SHIPPED default: an empty run_submitters "+
			"means anyone may send MsgRun. #6088 explains why it cannot fail closed "+
			"(GovDAO proposal creation is MsgRun-only)")
	cfg.Genesis.AppState = ggs

	// Pin Block.MaxGas to the PRODUCTION default. TestingMinimalNodeConfig ships
	// 30B ("calibrated for LMDB 59K per read"), 10x tm2's MaxBlockMaxGas, which
	// would make a depth-24 walk affordable and quietly defeat the point.
	cfg.Genesis.ConsensusParams.Block.MaxGas = 3_000_000_000

	node, remote := integration.TestingInMemoryNode(t, log.NewNoopLogger(), cfg)
	defer node.Stop()

	rpc, err := rpcclient.NewHTTPClient(remote)
	require.NoError(t, err)
	client := gnoclient.Client{Signer: signer, RPCClient: rpc}
	strangerClient := gnoclient.Client{Signer: strangerSigner, RPCClient: rpc}
	packerClient := gnoclient.Client{Signer: packerSigner, RPCClient: rpc}

	// feeFor sizes a fee that clears the min-gas-price check at this gas-wanted.
	feeFor := func(gas int64) std.Fee {
		return std.NewFee(gas, std.MustParseCoin(fmt.Sprintf("%dugnot", gas/10+1_000_000)))
	}

	// io is set because queryBlockMaxGas logs when the chain reports gpao's own
	// defaultBlockMaxGas — which is exactly the 3e9 pinned above, a useful
	// confirmation that the pin matches what gpao itself assumes.
	o := &oracle{client: client, io: commands.NewTestIO()}
	maxGas := o.queryBlockMaxGas(t.Context())
	require.Positive(t, maxGas)
	t.Logf("chain Block.MaxGas = %d", maxGas)

	mpkg := &std.MemPackage{
		Name: "parked",
		Path: parkedPath,
		Files: []*std.MemFile{
			// "code.gno" so the sources always sort before gnomod.toml; a MemPackage
			// with unsorted files is rejected as "invalid package path", which says
			// nothing about what actually went wrong.
			{Name: "code.gno", Body: fanOutSrc("parked", dosDepth)},
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(parkedPath)},
		},
	}

	t.Run("submit parks the bytes without type-checking them", func(t *testing.T) {
		tx, err := client.SignTx(std.Tx{
			Msgs: []std.Msg{vm.MsgAddPackage{Creator: approver, Package: mpkg}},
			Fee:  feeFor(maxGas),
		}, 0, 0)
		require.NoError(t, err)

		start := time.Now()
		res, err := client.BroadcastTxCommit(tx)
		elapsed := time.Since(start)
		require.NoError(t, err)
		require.True(t, res.CheckTx.IsOK(), "submit checkTx: %v", res.CheckTx.Error)
		require.True(t, res.DeliverTx.IsOK(), "submit deliverTx: %v", res.DeliverTx.Error)

		t.Logf("submit: %v, %d gas — parked, walk deferred",
			elapsed.Round(time.Millisecond), res.DeliverTx.GasUsed)
		assert.Less(t, elapsed, 3*time.Second,
			"under the inert policy submit must not type-check; if this is slow the "+
				"walk ran at submit")
	})

	t.Run("gpao's own probe runs the walk on the node and would approve it", func(t *testing.T) {
		probe, err := client.SignTx(std.Tx{
			Msgs: []std.Msg{vm.MsgEnablePackage{
				Approver: approver, PkgPath: parkedPath,
				PkgHash: vm.PackageContentHash(mpkg),
			}},
			Fee: feeFor(maxGas),
		}, 0, 0)
		require.NoError(t, err)

		start := time.Now()
		res, simErr := client.SimulateResult(probe)
		elapsed := time.Since(start)
		verdict := classifySimulate(res, simErr)
		t.Logf("gpao simulate: %v, verdict=%v (this abci_query occupied the node for "+
			"that long, unauthenticated and unpaid)", elapsed.Round(time.Millisecond), verdict)

		require.Equal(t, verdictWillFail, verdict,
			"gpao decides by simulating at Block.MaxGas; the walk must be priced "+
				"beyond that budget so the probe reports willFail and gpao declines. "+
				"A verdictReady here means gpao would APPROVE this package, and that "+
				"its own probe just spent %v of validator time", elapsed)
	})

	t.Run("the chain refuses the enable on its own, not only via the oracle", func(t *testing.T) {
		tx, err := client.SignTx(std.Tx{
			Msgs: []std.Msg{vm.MsgEnablePackage{
				Approver: approver, PkgPath: parkedPath,
				PkgHash: vm.PackageContentHash(mpkg),
			}},
			Fee: feeFor(maxGas),
		}, 0, 1)
		require.NoError(t, err)

		start := time.Now()
		res, err := client.BroadcastTxCommit(tx)
		elapsed := time.Since(start)
		outcome := describeBroadcast(res, err)
		t.Logf("enable: %v, %s", elapsed.Round(time.Millisecond), outcome)

		// -verify-budget is a per-operator flag and an approver need not be gpao at
		// all, so the chain cannot delegate this bound off-chain.
		require.Contains(t, outcome, "out of gas",
			"the CHAIN must refuse a package whose walk exceeds a whole block's gas; "+
				"if this succeeded it spent %v of every validator's time", elapsed)
	})

	t.Run("a stranger reaches the same walk through MsgRun", func(t *testing.T) {
		// NewMsgRun leaves Path empty on purpose; the handler forces
		// gno.land/e/<caller>/run, so the caller has no namespace to be checked
		// against — which is why MsgRun has no gate other than run_submitters.
		msg := vm.NewMsgRun(stranger, nil, []*std.MemFile{
			{Name: "code.gno", Body: fanOutSrc("main", dosDepth) + "\nfunc main() {}\n"},
		})
		tx, err := strangerClient.SignTx(std.Tx{
			Msgs: []std.Msg{msg},
			Fee:  feeFor(maxGas),
		}, 0, 0)
		require.NoError(t, err)

		start := time.Now()
		res, err := strangerClient.BroadcastTxCommit(tx)
		elapsed := time.Since(start)
		outcome := describeBroadcast(res, err)
		t.Logf("stranger MsgRun: %v, %s", elapsed.Round(time.Millisecond), outcome)

		require.Contains(t, outcome, "out of gas",
			"MsgRun ignores the inert policy and the approver set, and its allowlist "+
				"is empty by default, so this is the path no configuration closes; if "+
				"it succeeded a stranger just spent %v of every validator's time",
			elapsed)
	})

	t.Run("one transaction composes many walks that each pass on their own", func(t *testing.T) {
		// The composition case, and the one no per-PACKAGE bound can see. Each
		// message here is a depth-20 chain: 14,679,973 nodes, about 0.5s of
		// validType and 1.47e9 gas once priced. Individually that is unremarkable —
		// twenty times inside gpao's 10s budget, and affordable within a single
		// block's 3e9. It is the SUM that is the problem, and neither the oracle's
		// per-package budget nor a per-package ceiling is looking at the sum.
		//
		// Tx.Msgs is unbounded and ValidateBasic caps gas rather than the message
		// count, so composedMsgs is just a knob: raise it and the master-side time
		// grows linearly while the transaction stays as cheap as one message. It is
		// set low enough here that a demonstration run finishes.
		const composedMsgs = 30
		const shallowDepth = 20

		msgs := make([]std.Msg, 0, composedMsgs)
		for range composedMsgs {
			msgs = append(msgs, vm.NewMsgRun(packer, nil, []*std.MemFile{
				{Name: "code.gno", Body: fanOutSrc("main", shallowDepth) + "\nfunc main() {}\n"},
			}))
		}
		tx, err := packerClient.SignTx(std.Tx{Msgs: msgs, Fee: feeFor(maxGas)}, 0, 0)
		require.NoError(t, err)

		start := time.Now()
		res, err := packerClient.BroadcastTxCommit(tx)
		elapsed := time.Since(start)
		outcome := describeBroadcast(res, err)
		t.Logf("%d composed MsgRun: %v, %s", composedMsgs,
			elapsed.Round(time.Millisecond), outcome)

		// Priced, the meter is shared across the messages of one transaction, so the
		// first two exhaust the 3e9 budget and the rest never run: bounded at about
		// GasWanted/100 nodes no matter how many messages are packed in. Unpriced,
		// all thirty walk.
		require.Contains(t, outcome, "out of gas",
			"%d individually-cheap messages in one transaction must be bounded by the "+
				"transaction's gas, not by anything per-package; if this succeeded the "+
				"sum ran unpriced in %v", composedMsgs, elapsed)
	})
}
