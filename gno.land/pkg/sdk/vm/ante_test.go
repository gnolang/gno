package vm

import (
	"testing"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/jaekwon/testify/assert"
)

// The gas consumed is counted only towards the execution of gno messages
// We only abort the tx due to the insufficient gas.

func TestAddPkgSimulateGas(t *testing.T) {
	// setup
	success := true
	ctx, tx, anteHandler := setup(success)
	// simulation should not fail even if gas wanted is low

	ctx = ctx.WithMode(sdk.RunTxModeSimulate)
	simulate := true

	tx.Fee.GasWanted = 1
	gctx := auth.SetGasMeter(simulate, ctx, tx.Fee.GasWanted)
	gctx, res, abort := anteHandler(gctx, tx, simulate)
	gasSimulate := gctx.GasMeter().GasConsumed()

	assert.False(t, abort)
	assert.True(t, res.IsOK())
	assert.Equal(t, gasSimulate, int64(94055))
}

// failed tx will not aborted
func TestAddPkgSimulateFailedGas(t *testing.T) {
	// setup
	success := false
	ctx, tx, anteHandler := setup(success)
	// simulation should not fail even if gas wanted is low

	ctx = ctx.WithMode(sdk.RunTxModeSimulate)
	simulate := true

	tx.Fee.GasWanted = 1
	gctx := auth.SetGasMeter(simulate, ctx, tx.Fee.GasWanted)
	gctx, res, abort := anteHandler(gctx, tx, simulate)
	gasSimulate := gctx.GasMeter().GasConsumed()

	assert.False(t, abort)
	assert.True(t, res.IsOK())
	assert.Equal(t, gasSimulate, int64(18989))
}

func TestAddPkgCheckTxGas(t *testing.T) {
	success := true
	ctx, tx, anteHandler := setup(success)
	// Testing case with enough gas and succcful message execution

	ctx = ctx.WithMode(sdk.RunTxModeCheck)
	simulate := false
	tx.Fee.GasWanted = 500000
	gctx := auth.SetGasMeter(simulate, ctx, tx.Fee.GasWanted)
	gctx, res, abort := anteHandler(gctx, tx, simulate)
	gasCheck := gctx.GasMeter().GasConsumed()

	assert.False(t, abort)
	assert.True(t, res.IsOK())
	assert.Equal(t, gasCheck, int64(94055))
}

// CheckTx only abort when there is no enough gas meter.
func TestAddPkgCheckTxNoGas(t *testing.T) {
	success := true
	ctx, tx, anteHandler := setup(success)
	// Testing case with enough gas and succcful message execution
	ctx = ctx.WithMode(sdk.RunTxModeCheck)
	simulate := false
	tx.Fee.GasWanted = 3000
	gctx := auth.SetGasMeter(simulate, ctx, tx.Fee.GasWanted)

	var res sdk.Result
	abort := false

	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case store.OutOfGasException:
				res.Error = sdk.ABCIError(std.ErrOutOfGas(""))
				abort = true
			default:
				t.Errorf("should panic on OutOfGasException only")
			}
			assert.True(t, abort)
			assert.False(t, res.IsOK())
			gasCheck := gctx.GasMeter().GasConsumed()
			assert.Equal(t, gasCheck, int64(3231))
		} else {
			t.Errorf("should panic")
		}
	}()
	gctx, res, abort = anteHandler(gctx, tx, simulate)
}

// failed tx execution should pass the vm.AnteHandler
func TestAddPkgCheckTxFailedGas(t *testing.T) {
	success := false
	ctx, tx, anteHandler := setup(success)

	ctx = ctx.WithMode(sdk.RunTxModeCheck)
	simulate := false
	tx.Fee.GasWanted = 500000
	gctx := auth.SetGasMeter(simulate, ctx, tx.Fee.GasWanted)
	gctx, res, abort := anteHandler(gctx, tx, simulate)
	gasCheck := gctx.GasMeter().GasConsumed()

	assert.False(t, abort)
	assert.True(t, res.IsOK())
	assert.Equal(t, gasCheck, int64(18989))
}

// For deliver Tx ante handler does not check gas consumption and does not consume gas
func TestAddPkgDeliverTxGas(t *testing.T) {
	success := true
	ctx, tx, anteHandler := setup(success)

	var simulate bool

	ctx = ctx.WithMode(sdk.RunTxModeDeliver)
	simulate = false
	tx.Fee.GasWanted = 1
	gctx := auth.SetGasMeter(simulate, ctx, tx.Fee.GasWanted)
	gasDeliver := gctx.GasMeter().GasConsumed()
	gctx, res, abort := anteHandler(gctx, tx, simulate)
	assert.False(t, abort)
	assert.True(t, res.IsOK())
	assert.Equal(t, gasDeliver, int64(0))
}

// // For deliver Tx, ante handler does not check gas consumption and does not consume gas
func TestAddPkgDeliverTxFailGas(t *testing.T) {
	success := true
	ctx, tx, anteHandler := setup(success)

	var simulate bool

	ctx = ctx.WithMode(sdk.RunTxModeDeliver)
	simulate = false
	tx.Fee.GasWanted = 1
	gctx := auth.SetGasMeter(simulate, ctx, tx.Fee.GasWanted)
	gasDeliver := gctx.GasMeter().GasConsumed()
	gctx, res, abort := anteHandler(gctx, tx, simulate)
	assert.False(t, abort)
	assert.True(t, res.IsOK())
	assert.Equal(t, gasDeliver, int64(0))
}

func setup(success bool) (sdk.Context, sdk.Tx, sdk.AnteHandler) {
	// setup
	env := setupTestEnv()
	ctx := env.ctx
	// conduct base gas meter tests from a non-genesis block since genesis block use infinite gas meter instead.
	ctx = ctx.WithBlockHeader(&bft.Header{Height: int64(1)})
	anteHandler := NewAnteHandler(env.vmk)
	// Createa an account  with 10M gnot (10gnot)
	addr := crypto.AddressFromPreimage([]byte("test1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bank.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))
	// success message
	var files []*std.MemFile
	if success {
		files = []*std.MemFile{
			{
				Name: "hello.gno",
				Body: `package hello

import "std"

func Echo() string {
  return "hello world"
}`,
			},
		}
	} else {
		// falied message
		files = []*std.MemFile{
			{
				Name: "hello.gno",
				Body: `package hello

		import "std"

		func Echo() UnknowType {
			return "hello world"
		}`,
			},
		}
	}

	pkgPath := "gno.land/r/hello"
	// creat messages and a transaction
	msg := NewMsgAddPackage(addr, pkgPath, files)
	msgs := []std.Msg{msg}
	fee := std.NewFee(500000, std.MustParseCoin("1ugnot"))
	tx := std.NewTx(msgs, fee, []std.Signature{}, "")

	return ctx, tx, anteHandler
}
