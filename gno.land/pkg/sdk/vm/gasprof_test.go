package vm

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/stretchr/testify/require"
)

// Phase 3a: profiling a real on-chain tx through the keeper. Unlike the gno
// test surface (Phase 2), the on-chain path shares one meter across every
// dimension, so wrapping it once at MakeGnoTransactionStore captures cpu, alloc,
// AND store gas — and the profile reconciles exactly with the tx meter.
func TestVMKeeperGasProfile(t *testing.T) {
	env := setupTestEnv()

	// Simulate ante-handler gas already charged before the tx boundary.
	base := env.ctx.WithGasMeter(store.NewInfiniteGasMeter())
	base.GasMeter().ConsumeGas(12_345, "txSize")

	pctx, prof := WithGasProfile(base)
	ctx := env.vmk.MakeGnoTransactionStore(pctx)

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)

	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "test.gno", Body: `package test

var data []int

func Append(cur realm, n int) int {
	for i := 0; i < n; i++ {
		data = append(data, i)
	}
	return len(data)
}`},
	}
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(addr, pkgPath, files)))

	res, err := env.vmk.Call(ctx, NewMsgCall(addr, nil, pkgPath, "Append", []string{"40"}))
	require.NoError(t, err)
	require.Contains(t, res, "40")

	tot := prof.Totals()
	require.Positive(t, tot.CPU, "cpu gas captured")
	require.Positive(t, tot.Alloc, "alloc gas captured")
	require.Positive(t, tot.Store, "store gas captured on-chain (the Phase 3 win)")
	require.Equal(t, int64(12_345), tot.Other, "(ante) snapshot booked into the tree")

	// Reconciliation: the whole tree (ante + all dimensions minus refunds)
	// equals the tx meter's total. Nothing dropped or double-counted.
	net := tot.CPU + tot.Alloc + tot.Store + tot.Other - tot.Refund
	require.Equal(t, ctx.GasMeter().GasConsumed(), net,
		"profile reconciles with the tx meter")

	// Attribution: the called function is a nested node (leading ';' proves it
	// has a parent, trailing ' ' that it's a folded value line), and the (ante)
	// snapshot node is present.
	var fb strings.Builder
	require.NoError(t, prof.WriteFolded(&fb))
	require.Contains(t, fb.String(), ";gno.land/r/test.Append ")
	require.Contains(t, fb.String(), "(root);(ante) ")
}

// Without WithGasProfile, MakeGnoTransactionStore installs no wrapper, so tx
// execution is byte-identical to production (no profiling overhead, no change).
func TestVMKeeperGasProfile_offByDefault(t *testing.T) {
	env := setupTestEnv()
	require.Nil(t, getGasProfiler(env.ctx), "no profiler on a plain ctx")
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx.WithGasMeter(store.NewInfiniteGasMeter()))
	// The meter is not a profiling wrapper.
	_, wrapped := ctx.GasMeter().(interface{ Unwrap() store.GasMeter })
	require.False(t, wrapped, "no wrapper installed without WithGasProfile")
}

// Profiling must be observation-only: the same tx consumes identical gas with
// and without the profiler. Gas is deterministic across fresh envs.
func TestVMKeeperGasProfile_observationOnly(t *testing.T) {
	run := func(profile bool) int64 {
		env := setupTestEnv()
		base := env.ctx.WithGasMeter(store.NewInfiniteGasMeter())
		if profile {
			base, _ = WithGasProfile(base)
		}
		ctx := env.vmk.MakeGnoTransactionStore(base)
		addr := crypto.AddressFromPreimage([]byte("addr1"))
		acc := env.acck.NewAccountWithAddress(ctx, addr)
		env.acck.SetAccount(ctx, acc)
		env.bankk.SetCoins(ctx, addr, initialBalance)
		files := []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest("gno.land/r/test")},
			{Name: "test.gno", Body: `package test

var data []int

func Append(cur realm, n int) int {
	for i := 0; i < n; i++ {
		data = append(data, i)
	}
	return len(data)
}`},
		}
		require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(addr, "gno.land/r/test", files)))
		_, err := env.vmk.Call(ctx, NewMsgCall(addr, nil, "gno.land/r/test", "Append", []string{"40"}))
		require.NoError(t, err)
		return ctx.GasMeter().GasConsumed()
	}
	require.Equal(t, run(false), run(true), "profiling must not change gas consumed")
}

// The Run handler uses TWO profiler-driven machines (RunMemPackage +
// RunMainMaybeCrossing) sharing one cursor across a Release()-boundary reset.
// Profile a MsgRun and confirm both machines contribute to one reconciling tree.
func TestVMKeeperGasProfile_Run(t *testing.T) {
	env := setupTestEnv()
	base := env.ctx.WithGasMeter(store.NewInfiniteGasMeter())
	base.GasMeter().ConsumeGas(9_000, "txSize")
	pctx, prof := WithGasProfile(base)
	ctx := env.vmk.MakeGnoTransactionStore(pctx)

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)

	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest("gno.land/r/test")},
		{Name: "script.gno", Body: `package main

func main() {
	s := 0
	for i := 0; i < 100; i++ {
		s += i
	}
	println(s)
}
`},
	}
	res, err := env.vmk.Run(ctx, NewMsgRun(addr, std.MustParseCoins(""), files))
	require.NoError(t, err)
	require.Contains(t, res, "4950")

	tot := prof.Totals()
	require.Positive(t, tot.CPU, "cpu captured across both Run machines")
	require.Equal(t, int64(9_000), tot.Other, "(ante) snapshot booked")
	net := tot.CPU + tot.Alloc + tot.Store + tot.Other - tot.Refund
	require.Equal(t, ctx.GasMeter().GasConsumed(), net, "Run profile reconciles with the tx meter")

	var fb strings.Builder
	require.NoError(t, prof.WriteFolded(&fb))
	require.Contains(t, fb.String(), ".main ", "main() attributed with gas")
}
