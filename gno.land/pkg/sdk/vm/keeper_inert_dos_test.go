package vm

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// dosChainDepth picks a value-containment chain whose walk is slow enough to be
// unmistakable but short enough that a run finishes: depth 24 is
// 234,880,917 nodes, about 8.5s of validType on an Apple M5 (measured; see
// BenchmarkValidTypeWalk for the ns/node figure it derives from).
//
// It is deliberately ABOVE what a block can pay for once the walk is priced:
// 2.35e8 nodes x 100 gas/node = 2.35e10 gas against MaxBlockMaxGas of 3e9. So with
// the charge in place the message is unaffordable and dies instantly; without it
// the node walks for ~8.5s per message.
const dosChainDepth = 24

// dosSrc builds a doubling chain: each level references the previous one twice by
// value, which validType expands without memoizing (golang/go#65711).
func dosSrc(pkgName string, depth int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\ntype t0 struct{ v int }\n", pkgName)
	for i := 1; i <= depth; i++ {
		fmt.Fprintf(&b, "type t%d struct{ a, b [0]t%d }\n", i, i-1)
	}
	return b.String()
}

// TestInertFlowDoesNotBoundTheTypeCheckWalk walks the whole #6088 submission flow
// on a chain locked down exactly as that change intends — CodeSubmissionPolicy
// "inert", an approver configured, MsgRun left at its default — and shows where the
// flow does and does not bound the unmetered go/types validType walk.
//
// This is the in-process, keeper-level version. contribs/gpao's
// TestDoSGapAgainstARealChain covers the same three findings end to end — real node,
// real signed transactions, the ante handler, and gpao's own decision procedure —
// and adds the composed-transaction case. This one is kept because it is four times
// faster and lives in the main module, so it guards the same paths on a cheaper gate.
// Neither demonstrates a halt: the stalls here end and the node recovers.
//
// A demonstration suite as much as a regression test. Each subtest asserts what
// THIS branch produces, so on master (walk unpriced) the two DoS arms fail by
// succeeding, slowly. Run it in both trees to see the gap; the t.Logf lines carry
// the wall-clock numbers.
//
// What #6088 does close, and this suite confirms:
//
//   - submit no longer type-checks under the inert policy, so parking a
//     pathological package is cheap and harmless (subtest 1);
//   - gpao verifies each package off-chain against -verify-budget (10s default) and
//     refuses what it cannot finish, and broadcasts ONE MsgEnablePackage per
//     transaction (oracle.go:773) — so an approver-driven flow neither enables a
//     too-slow package nor composes several into one transaction.
//
// What it does not close:
//
//   - the CHAIN has no bound of its own at enable. The oracle's budget is an
//     off-chain admission policy set by a per-operator flag; a validator executing
//     MsgEnablePackage applies no budget at all. An approver that is not gpao, or a
//     gpao with a raised budget, wedges every validator (subtest 2).
//   - MsgRun bypasses the flow entirely: it type-checks and executes submitted
//     source immediately "under every policy including inert" (params.go), its
//     run_submitters allowlist is empty by default so anyone may send it, and it is
//     the only code-bearing message with no namespace or CLA gate. The same walk is
//     reachable by a stranger on a fully locked-down chain, no oracle involved
//     (subtest 3).
func TestInertFlowDoesNotBoundTheTypeCheckWalk(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	submitter := crypto.AddressFromPreimage([]byte("submitter"))
	stranger := crypto.AddressFromPreimage([]byte("stranger"))
	for _, addr := range []crypto.Address{approver, submitter, stranger} {
		acc := env.acck.NewAccountWithAddress(ctx, addr)
		env.acck.SetAccount(ctx, acc)
		env.bankk.SetCoins(ctx, addr, initialBalance)
	}

	// The chain as #6088 intends it: parked submissions, one approver.
	params := DefaultParams()
	params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
	params.PkgApprovers = []crypto.Address{approver}
	env.vmk.SetParams(ctx, params)

	// The gate that would stop subtest 3 is OFF by default, and #6088 explains why
	// it must be: run_submitters cannot fail closed without disabling MsgRun on
	// every chain that upgrades without editing genesis, and GovDAO proposal
	// creation is MsgRun-only. This is the shipped default, not a lax fixture.
	require.Empty(t, DefaultParams().RunSubmitters,
		"fixture assumes the shipped default: empty run_submitters means anyone may MsgRun")

	// The largest budget any transaction may carry — the ante handler caps
	// GasWanted at Block.MaxGas. These are not under-funded transactions; they are
	// as well-funded as consensus allows.
	const maxGas = int64(3_000_000_000)

	pkgFiles := func(name, path string, depth int) []*std.MemFile {
		// The .gno file must sort after gnomod.toml; unsorted files are rejected.
		return []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(path)},
			{Name: name + ".gno", Body: dosSrc(name, depth)},
		}
	}

	const parkedPath = "gno.land/r/test/parked"

	t.Run("submit parks without type-checking, so the bomb is cheap to plant", func(t *testing.T) {
		gm := types.NewGasMeter(maxGas)
		sctx := env.vmk.MakeGnoTransactionStore(env.ctx.WithGasMeter(gm))

		start := time.Now()
		err := env.vmk.AddPackage(sctx, NewMsgAddPackage(submitter, parkedPath,
			pkgFiles("parked", parkedPath, dosChainDepth)))
		elapsed := time.Since(start)
		require.NoError(t, err)
		env.vmk.CommitGnoTransactionStore(sctx)

		assert.NotNil(t, env.vmk.getGnoTransactionStore(ctx).GetInertPackage(parkedPath),
			"package must be parked")
		t.Logf("submit: %v, %d gas — the walk did not run (inert defers it)",
			elapsed.Round(time.Millisecond), gm.GasConsumed())

		// #6088 working as designed. It is also why the hazard moves rather than
		// disappears: the bytes are stored, and whoever enables them pays the walk.
		assert.Less(t, elapsed, 2*time.Second,
			"submit must not type-check under the inert policy; if this is slow the "+
				"walk ran at submit and the inert branch was not taken")
	})

	t.Run("enable has no chain-side bound: only gas stops the walk", func(t *testing.T) {
		mp := env.vmk.getGnoTransactionStore(ctx).GetInertPackage(parkedPath)
		require.NotNil(t, mp, "depends on the submit subtest")

		gm := types.NewGasMeter(maxGas)
		ectx := env.vmk.MakeGnoTransactionStore(env.ctx.WithGasMeter(gm))

		start := time.Now()
		err := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("%v", r)
				}
			}()
			return env.vmk.EnablePackage(ectx, MsgEnablePackage{
				Approver: approver, PkgPath: parkedPath, PkgHash: PackageContentHash(mp),
			})
		}()
		elapsed := time.Since(start)
		t.Logf("enable: %v, %d gas, err=%v", elapsed.Round(time.Millisecond), gm.GasConsumed(), err)

		// gpao would refuse this off-chain, over its 10s budget — but the chain must
		// not depend on that, because -verify-budget is a per-operator flag and an
		// approver need not be gpao. With the walk priced this costs 2.35e10 gas
		// against a 3e9 maximum and dies before walking; without it the enable
		// succeeds after ~8.5s of validator CPU and this assertion fails.
		require.Error(t, err,
			"enabling a package whose walk exceeds a whole block's gas must be refused "+
				"by the CHAIN, not only by the off-chain oracle; if this succeeded the "+
				"walk ran unpriced (elapsed %v)", elapsed)
		assert.Contains(t, err.Error(), "out of gas")
	})

	t.Run("MsgRun bypasses the flow: a stranger reaches the same walk", func(t *testing.T) {
		// No submission, no approver, no oracle. MsgRun carries its own source and
		// type-checks it immediately under every policy, including inert.
		gm := types.NewGasMeter(maxGas)
		rctx := env.vmk.MakeGnoTransactionStore(env.ctx.WithGasMeter(gm))

		start := time.Now()
		_, err := func() (res string, err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("%v", r)
				}
			}()
			return env.vmk.Run(rctx, NewMsgRun(stranger, nil, []*std.MemFile{
				{Name: "main.gno", Body: dosSrc("main", dosChainDepth) + "\nfunc main() {}\n"},
			}))
		}()
		elapsed := time.Since(start)
		t.Logf("run: %v, %d gas, err=%v", elapsed.Round(time.Millisecond), gm.GasConsumed(), err)

		require.Error(t, err,
			"a stranger's MsgRun must be refused by the chain on a locked-down inert "+
				"chain; the inert policy and approver set have no bearing on MsgRun, and "+
				"run_submitters is empty by default (elapsed %v)", elapsed)
		assert.Contains(t, err.Error(), "out of gas")
	})
}
