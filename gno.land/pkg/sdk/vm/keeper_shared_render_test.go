package vm

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sharedDAGRealmSource returns values that render far larger than the calls
// building them: Build's depth*n values render as about n^depth, and CallWide holds
// one 1,000-byte string n times.
const sharedDAGRealmSource = `package dag

import "strings"

func Build(n, depth int) any {
	if depth == 0 {
		return 1
	}
	child := Build(n, depth-1)
	arr := make([]any, n)
	for i := range arr {
		arr[i] = child
	}
	return arr
}

func CallBuild(cur realm, n, depth int) any { return Build(n, depth) }

func CallWide(cur realm, n int) []string { return wide(n) }

var grown []int

// CallGrowWide stores one more int, which needs a deposit, and returns wide(n).
func CallGrowWide(cur realm, n int) []string {
	grown = append(grown, 1)
	return wide(n)
}

func wide(n int) []string {
	s := strings.Repeat("A", 1000)
	out := make([]string, n)
	for i := range out {
		out[i] = s
	}
	return out
}
`

const sharedDAGPkgPath = "gno.land/r/dag" // must match sharedDAGRealmSource's `package dag`

// TestChargedBuffer_ChargesBeforeKeeping pins a call result's byte at an event
// byte's price, charged before the byte is kept: a write the gas left cannot
// pay for keeps nothing.
func TestChargedBuffer_ChargesBeforeKeeping(t *testing.T) {
	ten := 10 * store.DefaultGasConfig().WriteCostPerByte
	gm := store.NewGasMeter(ten)
	out := chargedBuffer{gm: gm}
	_, _ = out.Write([]byte("0123456789"))
	assert.Equal(t, ten, gm.GasConsumed())
	assert.Panics(t, func() { _, _ = out.Write([]byte("x")) })
	assert.Equal(t, "0123456789", out.buf.String())
}

// TestRun_OutputChargesGas pins a run's output, which the block keeps like a
// call's result, at the same price on top of what println charges to print it.
func TestRun_OutputChargesGas(t *testing.T) {
	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("runoutaddr"))
	run := func(arg string) (string, int64) {
		ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
		env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
		gm := store.NewGasMeter(100_000_000)
		ctx = ctx.WithGasMeter(gm)
		files := []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest("gno.land/r/runout")},
			{Name: "script.gno", Body: fmt.Sprintf("package main\n\nimport \"strings\"\n\nfunc main() {\n\ts, t := strings.Repeat(\"A\", 10_000), \"A\"\n\t_, _ = s, t\n\tprintln(%s)\n}\n", arg)},
		}
		res, err := env.vmk.Run(ctx, NewMsgRun(addr, nil, files))
		require.NoError(t, err)
		return res, gm.GasConsumed()
	}
	// The two scripts differ only in the one-letter variable they print.
	long, gasLong := run("s")
	short, gasShort := run("t")
	t.Logf("%d more bytes cost %d gas", len(long)-len(short), gasLong-gasShort)
	// 4 gas a byte for printing it and 14 for keeping it.
	assert.Equal(t, int64((4+14)*(len(long)-len(short))), gasLong-gasShort)
}

func TestCall_SharedDAGResultChargesGas(t *testing.T) {
	env := setupSizeGuardRealm(t, "dagcalladdr", sharedDAGPkgPath, sharedDAGRealmSource)
	caller := crypto.AddressFromPreimage([]byte("dagcalladdr"))

	callMsg := func(msg MsgCall) (res string, err error, oog bool) {
		ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
		ctx = ctx.WithGasMeter(store.NewGasMeter(600_000))
		defer func() {
			if r := recover(); r != nil {
				_, oog = r.(store.OutOfGasError)
				require.True(t, oog, "unexpected panic: %v", r)
			}
		}()
		res, err = env.vmk.Call(ctx, msg)
		return res, err, false
	}
	call := func(send std.Coins, fn string, args ...string) (string, error, bool) {
		return callMsg(NewMsgCall(caller, send, sharedDAGPkgPath, fn, args))
	}

	// Build(2, 6) renders 2,023 bytes inside the limit, the same bytes qeval returns.
	want, err := env.vmk.QueryEval(env.ctx, sharedDAGPkgPath, "Build(2, 6)")
	require.NoError(t, err)
	res, err, oog := call(nil, "CallBuild", "2", "6")
	require.NoError(t, err)
	require.False(t, oog)
	assert.Equal(t, want+"\n\n", res)

	// Each call below fits the limit with its render left free.
	_, _, oog = call(nil, "CallBuild", "4", "7")
	assert.True(t, oog, "a 262,135-byte result rendered inside 600,000 gas")
	_, _, oog = call(nil, "CallWide", "100")
	assert.True(t, oog, "a 101,217-byte result rendered inside 600,000 gas")

	// A send the callee never reads is refused before the result is rendered.
	_, err, oog = call(std.MustParseCoins("1ugnot"), "CallBuild", "4", "7")
	require.False(t, oog, "the render ran before the unobserved send was refused")
	assert.ErrorIs(t, err, UnobservedSendError{})

	// So is a storage deposit the call's MaxDeposit cannot cover.
	msg := NewMsgCall(caller, nil, sharedDAGPkgPath, "CallGrowWide", []string{"100"})
	msg.MaxDeposit = std.MustParseCoins("1ugnot")
	_, err, oog = callMsg(msg)
	require.False(t, oog, "the render ran before the deposit was refused")
	assert.ErrorContains(t, err, "not enough deposit to cover the storage usage")
}
