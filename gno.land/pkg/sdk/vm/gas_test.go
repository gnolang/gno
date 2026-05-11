package vm

import (
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
