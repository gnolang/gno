package vm

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
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
	tx.Fee.GasWanted = 3000
	gctx := auth.SetGasMeter(ctx, tx.Fee.GasWanted)
	// Has to be set up after gas meter in the context; so the stores are
	// correctly wrapped in gas stores.
	gctx = vmHandler.vm.MakeGnoTransactionStore(gctx)

	var res sdk.Result
	abort := false

	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case store.OutOfGasError:
				res.Error = sdk.ABCIError(std.ErrOutOfGas(""))
				abort = true
			default:
				t.Errorf("should panic on OutOfGasException only")
			}
			assert.True(t, abort)
			assert.False(t, res.IsOK())
			gasCheck := gctx.GasMeter().GasConsumed()
			assert.Equal(t, int64(3462), gasCheck)
		} else {
			t.Errorf("should panic")
		}
	}()
	msgs := tx.GetMsgs()
	res = vmHandler.Process(gctx, msgs[0])
}

// Enough gas for a successful message.
func TestAddPkgDeliverTx(t *testing.T) {
	isValidTx := true
	ctx, tx, vmHandler := setupAddPkg(isValidTx)

	ctx = ctx.WithMode(sdk.RunTxModeDeliver)
	tx.Fee.GasWanted = 500000
	gctx := auth.SetGasMeter(ctx, tx.Fee.GasWanted)
	gctx = vmHandler.vm.MakeGnoTransactionStore(gctx)
	msgs := tx.GetMsgs()
	res := vmHandler.Process(gctx, msgs[0])
	gasDeliver := gctx.GasMeter().GasConsumed()

	assert.True(t, res.IsOK())

	// NOTE: let's try to keep this bellow 250_000 :)
	assert.Equal(t, int64(226778), gasDeliver)
}

// Enough gas for a failed transaction.
func TestAddPkgDeliverTxFailed(t *testing.T) {
	isValidTx := false
	ctx, tx, vmHandler := setupAddPkg(isValidTx)

	ctx = ctx.WithMode(sdk.RunTxModeDeliver)
	tx.Fee.GasWanted = 500000
	gctx := auth.SetGasMeter(ctx, tx.Fee.GasWanted)
	gctx = vmHandler.vm.MakeGnoTransactionStore(gctx)
	msgs := tx.GetMsgs()
	res := vmHandler.Process(gctx, msgs[0])
	gasDeliver := gctx.GasMeter().GasConsumed()

	assert.False(t, res.IsOK())
	assert.Equal(t, int64(1231), gasDeliver)
}

// Not enough gas for a failed transaction.
func TestAddPkgDeliverTxFailedNoGas(t *testing.T) {
	isValidTx := false
	ctx, tx, vmHandler := setupAddPkg(isValidTx)

	ctx = ctx.WithMode(sdk.RunTxModeDeliver)
	tx.Fee.GasWanted = 1230
	gctx := auth.SetGasMeter(ctx, tx.Fee.GasWanted)
	gctx = vmHandler.vm.MakeGnoTransactionStore(gctx)

	var res sdk.Result
	abort := false

	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case store.OutOfGasError:
				res.Error = sdk.ABCIError(std.ErrOutOfGas(""))
				abort = true
			default:
				t.Errorf("should panic on OutOfGasException only")
			}
			assert.True(t, abort)
			assert.False(t, res.IsOK())
			gasCheck := gctx.GasMeter().GasConsumed()
			assert.Equal(t, int64(1231), gasCheck)
		} else {
			t.Errorf("should panic")
		}
	}()

	msgs := tx.GetMsgs()
	res = vmHandler.Process(gctx, msgs[0])
}

// TestAddPkgGasWithTypeCheckCache tests whether the typeCheckCache state (empty
// vs stdlib-populated) affects gas consumption for an addpkg tx that imports
// strconv.
//
// This reproduces the production scenario where:
//   - A genesis-fresh node (setupTestEnvCold) has an empty vm.typeCheckCache,
//     so every stdlib import during type-checking triggers a GetMemPackage store
//     read, which charges gas.
//   - A restarted node (setupTestEnv) has vm.typeCheckCache pre-populated with
//     stdlib, so stdlib imports are served from cache with no gas charged.
//
// If gas diverges between the two, that is the root cause of the non-determinism
// observed in the gnoland1 chain halt at block 352922.
func TestAddPkgGasWithTypeCheckCache(t *testing.T) {
	const pkgPath = "gno.land/r/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/counter"
	files := []*std.MemFile{
		{
			Name: "gnomod.toml",
			Body: `module = "gno.land/r/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/helloworld"
gno = "0.9"

[dependencies]
`,
		},
		{
			Name: "helloworld.gno",
			Body: `package helloworld

import "strconv"

var counter int

func init() {
	counter = 0
}

func Increment(cur realm) int {
	counter++
	return counter
}

func GetCounter() int {
	return counter
}

func Render(_ string) string {
	return "# Hello from Gno!\n\nCounter: " + strconv.Itoa(counter) + "\n"
}
`,
		},
	}

	addr := crypto.AddressFromPreimage([]byte("test1"))

	runAddPkg := func(env testEnv) int64 {
		ctx := env.ctx.WithBlockHeader(&bft.Header{ChainID: "test-chain-id", Height: 1})
		acc := env.acck.NewAccountWithAddress(ctx, addr)
		env.acck.SetAccount(ctx, acc)
		env.bankk.SetCoins(ctx, addr, std.MustParseCoins(ugnot.ValueString(10_000_000)))
		gctx := auth.SetGasMeter(ctx, 10_000_000)
		gctx = env.vmk.MakeGnoTransactionStore(gctx)
		msg := NewMsgAddPackage(addr, pkgPath, files)
		err := env.vmk.AddPackage(gctx, msg)
		if err != nil {
			t.Fatalf("AddPackage error: %v", err)
		}
		return gctx.GasMeter().GasConsumed()
	}

	// setupTestEnvCold uses LoadStdlib (no cache): vm.typeCheckCache stays empty.
	// This simulates a production node that started from genesis and was never restarted.
	gasCold := runAddPkg(setupTestEnvCold())

	// setupTestEnv uses LoadStdlibCached: vm.typeCheckCache is populated with stdlib.
	// This simulates a production node that was restarted (Initialize populates the cache).
	gasWarm := runAddPkg(setupTestEnv())

	t.Logf("gas cold (empty typeCheckCache): %d", gasCold)
	t.Logf("gas warm (stdlib typeCheckCache): %d", gasWarm)

	assert.Equal(t, gasWarm, gasCold, "gas must be identical regardless of whether typeCheckCache was pre-populated; a difference means genesis-fresh nodes and restarted nodes will disagree on gas, causing a consensus halt")
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

// TestTypeCheckCacheContainsAllStdlibs verifies that every stdlib package in
// InitOrder is present in vm.typeCheckCache after initialization.
//
// The failure mode: TypeCheckMemPackage writes a package's result to permCache
// only when it is imported as a dependency (ImportFrom with canPerm=true).
// The root package of each call is never written there.  This means any stdlib
// that is not imported by a subsequent stdlib in the loop ends up missing.
func TestTypeCheckCacheContainsAllStdlibs(t *testing.T) {
	for _, name := range []string{"cold (LoadStdlib)", "warm (LoadStdlibCached)"} {
		var env testEnv
		if name == "cold (LoadStdlib)" {
			env = setupTestEnvCold()
		} else {
			env = setupTestEnv()
		}
		cache := env.vmk.typeCheckCache
		var missing []string
		for _, lib := range stdlibs.InitOrder() {
			if cache[lib] == nil {
				missing = append(missing, lib)
			}
		}
		if len(missing) > 0 {
			t.Errorf("%s: %d stdlib(s) missing from typeCheckCache: %v", name, len(missing), missing)
		}
	}
}
