package vm

import (
	"strconv"
	"testing"

	"github.com/gnolang/gno/gnovm/stdlibs/chain"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// The block keeps, encodes and stores every emitted byte, so a byte must cost
// at least a store write's per-byte price, in typ as much as in attrs.
func TestEmitChargesPerByte(t *testing.T) {
	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	const pkgPath = "gno.land/r/test/emitgas"
	{
		ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
		env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
		env.bankk.SetCoins(ctx, addr, initialBalance)
		deployJSONTestPkg(t, env, ctx, addr, pkgPath, `package emitgas

import (
	"chain"
	"strings"
)

var (
	short = "v"
	long  = strings.Repeat("v", 4096)
)

func pick(big bool) string {
	if big {
		return long
	}
	return short
}

func Value(cur realm, big bool, n int) {
	v := pick(big)
	for i := 0; i < n; i++ {
		chain.Emit("t", "k", v)
	}
}

func Type(cur realm, big bool, n int) {
	typ := pick(big)
	for i := 0; i < n; i++ {
		chain.Emit(typ)
	}
}

var emptyPairs = make([]string, 128)

func Pairs(cur realm, full bool, n int) {
	var attrs []string
	if full {
		attrs = emptyPairs
	}
	for i := 0; i < n; i++ {
		chain.Emit("t", attrs...)
	}
}
`)
	}

	gasFor := func(fn string, big bool, n int) int64 {
		t.Helper()
		gm := types.NewInfiniteGasMeter()
		ctx := env.vmk.MakeGnoTransactionStore(env.ctx.WithGasMeter(gm).WithEventLogger(sdk.NewEventLogger()))
		args := []string{strconv.FormatBool(big), strconv.Itoa(n)}
		_, err := env.vmk.Call(ctx, NewMsgCall(addr, nil, pkgPath, fn, args))
		require.NoError(t, err)
		return gm.GasConsumed()
	}
	// Gas of one more Emit call: the difference between two calls and one
	// cancels the transaction's fixed costs, loading the string included.
	perCall := func(fn string, big bool) int64 {
		return gasFor(fn, big, 2) - gasFor(fn, big, 1)
	}

	minPerByte := store.DefaultGasConfig().WriteCostPerByte
	for _, fn := range []string{"Value", "Type"} {
		perByte := (perCall(fn, true) - perCall(fn, false)) / (4096 - 1)
		t.Logf("%s: %d gas per emitted byte", fn, perByte)
		require.GreaterOrEqual(t, perByte, minPerByte,
			"%s: an emitted byte must cost at least a stored byte", fn)
	}

	// An empty pair carries no bytes of its own, but the block stores the
	// framing it adds to the event's encoding, so the charge follows the
	// encoding byte for byte.
	for _, full := range []bool{false, true} {
		evt := chain.Event{Type: "t", PkgPath: pkgPath}
		if full {
			evt.Attributes = make([]chain.EventAttribute, 64)
		}
		gm := &emitMeter{GasMeter: types.NewInfiniteGasMeter()}
		ctx := env.vmk.MakeGnoTransactionStore(env.ctx.WithGasMeter(gm).WithEventLogger(sdk.NewEventLogger()))
		_, err := env.vmk.Call(ctx, NewMsgCall(addr, nil, pkgPath, "Pairs", []string{strconv.FormatBool(full), "1"}))
		require.NoError(t, err)
		want := minPerByte * int64(len(amino.MustMarshalAny(evt)))
		require.Equal(t, want, gm.emitted, "full=%v: the charge must be the event's encoded bytes", full)
	}
}

// emitMeter records what X_emit charges for an event's bytes.
type emitMeter struct {
	types.GasMeter
	emitted types.Gas
}

func (m *emitMeter) ConsumeGas(amount types.Gas, descriptor string) {
	if descriptor == "EmitEventPerByte" {
		m.emitted += amount
	}
	m.GasMeter.ConsumeGas(amount, descriptor)
}
