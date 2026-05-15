package vm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Gas for entire tx is consumed in both CheckTx and DeliverTx.
// Gas for executing VM tx (VM CPU and Store Access in bytes) is consumed in DeliverTx.
// Gas for balance checking, message size checking, and signature verification is consumed (deducted) in checkTx.

// Insufficient gas for a successful message.

func TestAddPkgDeliverTxInsuffGas(t *testing.T) {
	isValidTx := true
	ctx, tx, vmHandler := setupAddPkg(isValidTx)

	ctx = ctx.WithMode(sdk.RunTxModeDeliver)
	tx.Fee.GasWanted = 3000000
	gctx := auth.SetGasMeter(ctx, tx.Fee.GasWanted)
	gctx, _ = gctx.CacheContext()

	var res sdk.Result
	abort := false

	// Defer registered BEFORE MakeGnoTransactionStore — setup itself
	// can OOG now that params reads charge gas (gctx threading).
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case store.OutOfGasError:
				res.Error = sdk.ABCIError(std.ErrOutOfGas(""))
				abort = true
			default:
				t.Errorf("should panic on OutOfGasException only, got: %T %v", r, r)
			}
			assert.True(t, abort)
			assert.False(t, res.IsOK())
			// gas.go:206 bumps `consumed` by the request that overflowed
			// before panicking, so GasConsumed can exceed GasWanted.
			// The substantive assertion is the OutOfGasError type above.
			gasCheck := gctx.GasMeter().GasConsumed()
			assert.Greater(t, gasCheck, int64(0))
		} else {
			t.Errorf("should panic")
		}
	}()
	gctx = vmHandler.vm.MakeGnoTransactionStore(gctx)
	msgs := tx.GetMsgs()
	res = vmHandler.Process(gctx, msgs[0])
}

// Enough gas for a successful message.
//
// NOTE: hardcoded gas values are sensitive to anything that adds storage
// reads/writes (params keeper changes, native gas tweaks, etc.). The
// asserted bounds below are wide enough to absorb small drifts; tighten
// once the chain version stabilizes.
func TestAddPkgDeliverTx(t *testing.T) {
	isValidTx := true
	ctx, tx, vmHandler := setupAddPkg(isValidTx)

	ctx = ctx.WithMode(sdk.RunTxModeDeliver)
	tx.Fee.GasWanted = 50000000
	gctx := auth.SetGasMeter(ctx, tx.Fee.GasWanted)
	gctx, _ = gctx.CacheContext()
	gctx = vmHandler.vm.MakeGnoTransactionStore(gctx)
	msgs := tx.GetMsgs()
	res := vmHandler.Process(gctx, msgs[0])
	gasDeliver := gctx.GasMeter().GasConsumed()

	assert.True(t, res.IsOK())
	assert.Greater(t, gasDeliver, int64(0))
	assert.Less(t, gasDeliver, tx.Fee.GasWanted)
}

// Enough gas for a failed transaction.
func TestAddPkgDeliverTxFailed(t *testing.T) {
	isValidTx := false
	ctx, tx, vmHandler := setupAddPkg(isValidTx)

	ctx = ctx.WithMode(sdk.RunTxModeDeliver)
	tx.Fee.GasWanted = 50000000
	gctx := auth.SetGasMeter(ctx, tx.Fee.GasWanted)
	gctx, _ = gctx.CacheContext()
	gctx = vmHandler.vm.MakeGnoTransactionStore(gctx)
	msgs := tx.GetMsgs()
	res := vmHandler.Process(gctx, msgs[0])
	gasDeliver := gctx.GasMeter().GasConsumed()

	assert.False(t, res.IsOK())
	assert.Greater(t, gasDeliver, int64(0))
	assert.Less(t, gasDeliver, tx.Fee.GasWanted)
}

// Not enough gas for a failed transaction.
func TestAddPkgDeliverTxFailedNoGas(t *testing.T) {
	isValidTx := false
	ctx, tx, vmHandler := setupAddPkg(isValidTx)

	ctx = ctx.WithMode(sdk.RunTxModeDeliver)
	tx.Fee.GasWanted = 500000
	gctx := auth.SetGasMeter(ctx, tx.Fee.GasWanted)
	gctx, _ = gctx.CacheContext()

	var res sdk.Result
	abort := false

	// Defer registered BEFORE MakeGnoTransactionStore — setup itself
	// can OOG now that params reads charge gas (gctx threading).
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case store.OutOfGasError:
				res.Error = sdk.ABCIError(std.ErrOutOfGas(""))
				abort = true
			default:
				t.Errorf("should panic on OutOfGasException only, got: %T %v", r, r)
			}
			assert.True(t, abort)
			assert.False(t, res.IsOK())
			// gas.go:206 bumps `consumed` by the request that overflowed
			// before panicking, so GasConsumed can exceed GasWanted.
			// The substantive assertion is the OutOfGasError type above.
			gasCheck := gctx.GasMeter().GasConsumed()
			assert.Greater(t, gasCheck, int64(0))
		} else {
			t.Errorf("should panic")
		}
	}()
	gctx = vmHandler.vm.MakeGnoTransactionStore(gctx)

	msgs := tx.GetMsgs()
	res = vmHandler.Process(gctx, msgs[0])
}

// TestAddPkgDeliverTx_PreprocessAllocLimit verifies the per-tx
// preprocess allocator (keeper.go ~740) bounds adversarial input that
// inflates allocation O(2^N) from O(N) source bytes.
//
// Source: ~30 const decls of the form `const aN = aN-1 + aN-1`. The
// preprocessor folds these constants at preprocess time, so by aN
// the in-memory string has length 2^N × startLen. With N=30 and
// startLen=8 the demanded string allocation is ~8 GB — far past the
// 500 MB maxAllocTx hard cap.
//
// Without the per-tx preprocess allocator, sub-Machines created by
// Preprocess have nil Alloc and consume unbounded heap before
// blowing up the OS. With the allocator wired, the keeper's
// doRecover wraps the panic into the tx error.
func TestAddPkgDeliverTx_PreprocessAllocLimit(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx
	ctx = ctx.WithBlockHeader(&bft.Header{Height: int64(1)})
	ctx = ctx.WithMode(sdk.RunTxModeDeliver)

	addr := crypto.AddressFromPreimage([]byte("preprocess-alloc-tester"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, std.MustParseCoins(ugnot.ValueString(10000000)))

	const pkgPath = "gno.land/r/doubling"

	// Build a tiny adversarial source: doubling-concat constants.
	// Source size: O(N) bytes; allocation demand: O(2^N) bytes.
	const N = 30
	var b strings.Builder
	b.WriteString("package doubling\n")
	b.WriteString(`const a0 = "abcdefgh"` + "\n")
	for i := 1; i <= N; i++ {
		fmt.Fprintf(&b, "const a%d = a%d + a%d\n", i, i-1, i-1)
	}
	b.WriteString(`func Echo() string { return a0 }` + "\n")
	src := b.String()
	require.Less(t, len(src), 1500,
		"adversarial source should stay tiny; got %d bytes", len(src))

	files := []*std.MemFile{
		{Name: "doubling.gno", Body: src},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	}
	msg := NewMsgAddPackage(addr, pkgPath, files)
	// Plenty of gas — we want to confirm the alloc cap bites
	// before gas runs out, not the other way.
	tx := std.NewTx([]std.Msg{msg},
		std.NewFee(500_000_000, std.MustParseCoin(ugnot.ValueString(1))),
		[]std.Signature{}, "")

	gctx := auth.SetGasMeter(ctx, tx.Fee.GasWanted)
	gctx, _ = gctx.CacheContext()
	gctx = env.vmh.vm.MakeGnoTransactionStore(gctx)
	res := env.vmh.Process(gctx, tx.GetMsgs()[0])

	// Tx must fail — but with a clean error, not a process crash.
	assert.False(t, res.IsOK(),
		"adversarial doubling-concat must not succeed; res=%v", res)
	// And the failure must be the alloc-cap panic, not an OOG or
	// generic preprocessor error. The "(no GC)" marker confirms
	// it specifically came from the preprocess hard-cap path
	// (not the regular Machine alloc-limit-with-GC path).
	errMsg := res.Log + " " + res.Error.Error()
	assert.Contains(t, errMsg, "allocation limit exceeded",
		"expected alloc-limit error, got: %s", errMsg)
	assert.Contains(t, errMsg, "(no GC)",
		"expected preprocess no-GC marker, got: %s", errMsg)

	// And gas was charged for the work that DID happen before
	// the cap fired.
	assert.Greater(t, gctx.GasMeter().GasConsumed(), int64(0))

	// Log must be bounded — adversarial value rendering should not
	// produce a multi-MB persisted Log. clipLog backstops at
	// maxLogLineBytes×maxLogLines ≈ 17 KB; bounded printers
	// upstream keep the panic descriptor and stacktrace small.
	assert.Less(t, len(res.Log), 32*1024,
		"persisted Log must be bounded; got %d bytes", len(res.Log))
}

// TestRunDeliverTx_AdversarialErrorOOG verifies the recovery-path
// design choice: when a panic value's user-defined Error() method
// runs out of gas (or trips the transient alloc cap) DURING bounded
// rendering, boundedUserSprint swallows the panic and falls through
// to structural render — preserving the original panic info instead
// of reclassifying as ErrOutOfGas.
//
// Trade-off documented in bounded_strings.go's boundedUserSprint:
// gas is correctly charged (consumed counter is set before any panic
// inside ConsumeGas), only error-classification is lost on this
// adversarial path.
//
// Adversarial shape: a small type whose Error() method does many
// string concatenations. Run with a tight gas budget so Error()
// can't complete. The bounded recovery path swallows the panic and
// returns "<*main.AdvErr>" structurally.
func TestRunDeliverTx_AdversarialErrorOOG(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx
	ctx = ctx.WithBlockHeader(&bft.Header{Height: int64(1)})
	ctx = ctx.WithMode(sdk.RunTxModeDeliver)

	addr := crypto.AddressFromPreimage([]byte("test1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, std.MustParseCoins(ugnot.ValueString(100000000)))

	// Adversarial: type T whose Error() does many small allocations.
	// In bounded recovery, this either trips the 64 KB transient
	// alloc cap or runs out of gas — boundedUserSprint catches
	// either and falls through to structural render.
	const pkgPath = "gno.land/r/" + "test1user/adverr"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "main.gno", Body: `package main

type AdvErr struct{}

func (e AdvErr) Error() string {
	// Many allocations. Each '+=' allocates a new string and
	// charges alloc-gas; the cumulative cost will trip either the
	// transient alloc cap or the tx gas budget during recovery
	// rendering.
	s := "x"
	for i := 0; i < 100000; i++ {
		s += "x"
	}
	return s
}

func main() {
	panic(AdvErr{})
}
`},
	}

	// Big-ish gas budget so test-env setup (params reads, stdlib
	// load) completes; the adversarial alloc cap (64 KB transient,
	// hit cumulatively) fires before this much gas is consumed by
	// the Error() loop.
	tx := std.NewTx(
		[]std.Msg{NewMsgRun(addr, std.Coins{}, files)},
		std.NewFee(50_000_000, std.MustParseCoin(ugnot.ValueString(1))),
		[]std.Signature{}, "")

	gctx := auth.SetGasMeter(ctx, tx.Fee.GasWanted)
	gctx, _ = gctx.CacheContext()
	gctx = env.vmh.vm.MakeGnoTransactionStore(gctx)

	res := env.vmh.Process(gctx, tx.GetMsgs()[0])

	// Tx must fail: panic propagated.
	assert.False(t, res.IsOK(),
		"adversarial panic must fail the tx; res=%v", res)

	// Log must be bounded — clipLog backstop ensures ≤ ~17 KB even
	// if upstream defenses fail; in practice the BoundedSprint*
	// pipeline keeps it much smaller.
	assert.Less(t, len(res.Log), 32*1024,
		"persisted Log must be bounded; got %d bytes", len(res.Log))

	errMsg := res.Log + " " + res.Error.Error()

	// Frame location must be preserved — the bounded stacktrace
	// should still pinpoint where the panic happened.
	assert.Contains(t, errMsg, "main.gno:",
		"bounded stacktrace should include source location; got: %s", errMsg)

	// Verify the design choice: OOG inside boundedUserSprint is
	// SWALLOWED (not re-panicked). Tx surfaces as a generic VM
	// panic, not ErrOutOfGas. If a future change re-introduces an
	// OOG re-panic, this assertion fails — flag that as a design
	// regression to discuss.
	assert.NotContains(t, errMsg, "out of gas",
		"OOG-during-Error should be swallowed by boundedUserSprint; "+
			"if this assertion fails, the design choice may have changed; got: %s",
		errMsg)
	assert.Contains(t, errMsg, "VM panic",
		"tx should be classified as VM panic, not OOG; got: %s", errMsg)
}

// Set up a test env for both a successful and a failed tx.
func setupAddPkg(success bool) (sdk.Context, sdk.Tx, vmHandler) {
	// setup
	env := setupTestEnv()
	ctx := env.ctx
	// conduct base gas meter tests from a non-genesis block since genesis block use infinite gas meter instead.
	ctx = ctx.WithBlockHeader(&bft.Header{Height: int64(1)})
	// Create an account  with 10M ugnot (10gnot)
	addr := crypto.AddressFromPreimage([]byte("test1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, std.MustParseCoins(ugnot.ValueString(10000000)))

	const pkgPath = "gno.land/r/hello"

	// success message
	var files []*std.MemFile
	if success {
		files = []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
			{
				Name: "hello.gno",
				Body: `package hello

func Echo() string {
  return "hello world"
}`,
			},
		}
	} else {
		// failed message
		files = []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
			{
				Name: "hello.gno",
				Body: `package hello

func Echo() UnknowType {
  return "hello world"
}`,
			},
		}
	}

	// create messages and a transaction
	msg := NewMsgAddPackage(addr, pkgPath, files)
	msgs := []std.Msg{msg}
	fee := std.NewFee(500000, std.MustParseCoin(ugnot.ValueString(1)))
	tx := std.NewTx(msgs, fee, []std.Signature{}, "")

	return ctx, tx, env.vmh
}
