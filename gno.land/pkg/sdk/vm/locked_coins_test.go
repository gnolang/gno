package vm

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// TestGetLockedCoinsTracksTheSchedule is the realm-visible form of
// std.VestingSchedule.LockedCoins. The numbers are the ones vesting_test.go
// already pins for a 1000ugnot grant from t=100 to t=200, read back through
// SDKBanker and through a realm query at each block time. GetCoins stays the
// whole balance, locked coins included.
func TestGetLockedCoinsTracksTheSchedule(t *testing.T) {
	env := setupTestEnv()

	const pkgPath = "gno.land/r/lockedcoins"
	deployLockedRealm(t, env, "locked-deployer", pkgPath, `package lockedcoins

import "chain/banker"

func Report(addr address) string {
	return banker.NewReadonlyBanker().GetLockedCoins(addr).String()
}
`)

	balance := std.MustParseCoins("5000ugnot")
	linearAddr := fundVesting(t, env, "linear", balance, std.VestingSchedule{
		OriginalVesting: std.MustParseCoins("1000ugnot"),
		StartTime:       100,
		EndTime:         200,
	})
	cliffAddr := fundVesting(t, env, "cliff", balance, std.VestingSchedule{
		OriginalVesting: std.MustParseCoins("1000ugnot"),
		EndTime:         200,
		Type:            std.VestingDelayed,
	})
	// Two denoms, so a partial vest cannot drop one and keep the other by accident.
	multiAddr := fundVesting(t, env, "multi", std.MustParseCoins("5000atom,5000ugnot"), std.VestingSchedule{
		OriginalVesting: std.MustParseCoins("1000atom,2000ugnot"),
		StartTime:       100,
		EndTime:         200,
	})
	openAddr := fundVesting(t, env, "open", balance, std.VestingSchedule{})
	missing := crypto.AddressFromPreimage([]byte("no-such-account"))

	// Same samples as TestVestingSchedule_LinearVesting / CliffVesting.
	linearLocked := []struct {
		at     int64
		locked int64
	}{
		{50, 1000},
		{100, 1000},
		{101, 990},
		{150, 500},
		{199, 10},
		{200, 0},
		{250, 0},
	}
	cliffLocked := []struct {
		at     int64
		locked int64
	}{
		{50, 1000},
		{150, 1000},
		{199, 1000},
		{200, 0},
		{250, 0},
	}

	for _, tc := range linearLocked {
		qctx := atBlockTime(env, tc.at)
		t.Run(fmt.Sprintf("linear/%d", tc.at), func(t *testing.T) {
			assertLocked(t, env, qctx, pkgPath, linearAddr, "ugnot", tc.locked, balance)
		})
	}
	for _, tc := range cliffLocked {
		qctx := atBlockTime(env, tc.at)
		t.Run(fmt.Sprintf("cliff/%d", tc.at), func(t *testing.T) {
			assertLocked(t, env, qctx, pkgPath, cliffAddr, "ugnot", tc.locked, balance)
		})
	}

	// Halfway: 1000*50/100 atom still locked, 2000*50/100 ugnot still locked.
	qctx := atBlockTime(env, 150)
	multi := NewSDKBanker(env.vmk, qctx).GetLockedCoins(multiAddr.Bech32())
	require.Equal(t, int64(500), multi.AmountOf("atom"))
	require.Equal(t, int64(1000), multi.AmountOf("ugnot"))
	require.Equal(t, "500atom,1000ugnot", reportLocked(t, env, qctx, pkgPath, multiAddr))

	require.True(t, NewSDKBanker(env.vmk, qctx).GetLockedCoins(openAddr.Bech32()).IsZero())
	require.Empty(t, reportLocked(t, env, qctx, pkgPath, openAddr))
	require.Nil(t, NewSDKBanker(env.vmk, qctx).GetLockedCoins(missing.Bech32()))
	require.Empty(t, reportLocked(t, env, qctx, pkgPath, missing))

	require.Panics(t, func() {
		NewSDKBanker(env.vmk, qctx).GetLockedCoins("not-an-address")
	})
}

func fundVesting(t *testing.T, env testEnv, name string, balance std.Coins, vs std.VestingSchedule) crypto.Address {
	t.Helper()
	addr := crypto.AddressFromPreimage([]byte("locked-" + name))
	acc := env.acck.NewAccountWithAddress(env.ctx, addr)
	acc.SetVesting(vs)
	env.acck.SetAccount(env.ctx, acc)
	require.NoError(t, env.bankk.SetCoins(env.ctx, addr, balance))
	return addr
}

func atBlockTime(env testEnv, unix int64) sdk.Context {
	return env.ctx.WithBlockHeader(&bft.Header{
		ChainID: env.ctx.ChainID(),
		Height:  env.ctx.BlockHeight(),
		Time:    time.Unix(unix, 0).UTC(),
	})
}

func assertLocked(t *testing.T, env testEnv, ctx sdk.Context, pkgPath string, addr crypto.Address, denom string, locked int64, balance std.Coins) {
	t.Helper()
	bnk := NewSDKBanker(env.vmk, ctx)
	got := bnk.GetLockedCoins(addr.Bech32())
	require.Equal(t, locked, got.AmountOf(denom))
	require.True(t, got.IsEqual(env.acck.GetAccount(ctx, addr).LockedCoins(ctx.BlockTime())),
		"banker and account disagree: banker=%s account=%s", got, env.acck.GetAccount(ctx, addr).LockedCoins(ctx.BlockTime()))
	require.True(t, bnk.GetCoins(addr.Bech32()).IsEqual(balance),
		"GetCoins must keep returning the whole balance")
	require.Equal(t, got.String(), reportLocked(t, env, ctx, pkgPath, addr))
}

func deployLockedRealm(t *testing.T, env testEnv, deployerSeed, pkgPath, body string) crypto.Address {
	t.Helper()
	deployer := crypto.AddressFromPreimage([]byte(deployerSeed))
	env.acck.SetAccount(env.ctx, env.acck.NewAccountWithAddress(env.ctx, deployer))
	require.NoError(t, env.bankk.SetCoins(env.ctx, deployer, initialBalance))

	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(deployer, pkgPath, []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "locked.gno", Body: body},
	})))
	env.vmk.CommitGnoTransactionStore(ctx)
	return deployer
}

func reportLocked(t *testing.T, env testEnv, ctx sdk.Context, pkgPath string, addr crypto.Address) string {
	t.Helper()
	got, err := env.vmk.QueryEvalString(ctx, pkgPath, fmt.Sprintf("Report(%q)", addr.Bech32()))
	require.NoError(t, err)
	return got
}

// TestGetLockedCoinsUnclampedAndSpendable pins two properties together:
//   - GetLockedCoins returns the schedule's unclamped remainder even when
//     unrestricted debits (such as gas fees) drop the balance below the locked
//     amount, all the way to zero.
//   - Spendable = max(GetCoin - GetLockedCoins.AmountOf, 0) matches what
//     BankKeeper.SendCoins / SubtractCoins enforces.
func TestGetLockedCoinsUnclampedAndSpendable(t *testing.T) {
	env := setupTestEnv()

	const pkgPath = "gno.land/r/lockedunclamped"
	deployLockedRealm(t, env, "unclamped-deployer", pkgPath, `package lockedunclamped

import "chain/banker"

func Report(addr address) string {
	return banker.NewReadonlyBanker().GetLockedCoins(addr).String()
}
`)

	addr := fundVesting(t, env, "unclamped", std.MustParseCoins("1500ugnot"), std.VestingSchedule{
		OriginalVesting: std.MustParseCoins("1000ugnot"),
		StartTime:       100,
		EndTime:         200,
	})
	payee := crypto.AddressFromPreimage([]byte("unclamped-payee"))
	feeCollector := crypto.AddressFromPreimage([]byte("unclamped-fees"))

	tctx := atBlockTime(env, 100)
	bnk := NewSDKBanker(env.vmk, tctx)

	// 1500 total - 1000 locked = 500 spendable.
	require.Equal(t, int64(1500), bnk.GetCoin(addr.Bech32(), "ugnot"))
	require.Equal(t, int64(1000), bnk.GetLockedCoins(addr.Bech32()).AmountOf("ugnot"))
	require.Equal(t, "1000ugnot", reportLocked(t, env, tctx, pkgPath, addr))

	// Sending 1 over spendable fails with VestingLockedCoinsError; exact spendable succeeds.
	err := env.bankk.SendCoins(tctx, addr, payee, std.MustParseCoins("501ugnot"))
	require.ErrorIs(t, err, std.VestingLockedCoinsError{})
	require.NoError(t, env.bankk.SendCoins(tctx, addr, payee, std.MustParseCoins("500ugnot")))

	// Unrestricted fee debit spends into the locked portion: balance drops to 400,
	// while the schedule still locks 1000. GetLockedCoins must stay unclamped.
	require.NoError(t, env.bankk.SendCoinsUnrestricted(tctx, addr, feeCollector, std.MustParseCoins("600ugnot")))
	require.Equal(t, int64(400), bnk.GetCoin(addr.Bech32(), "ugnot"))
	require.Equal(t, int64(1000), bnk.GetLockedCoins(addr.Bech32()).AmountOf("ugnot"))
	require.Equal(t, "1000ugnot", reportLocked(t, env, tctx, pkgPath, addr))

	err = env.bankk.SendCoins(tctx, addr, payee, std.MustParseCoins("1ugnot"))
	require.ErrorIs(t, err, std.VestingLockedCoinsError{})

	// Even at zero balance, the schedule's locked remainder is still reported unclamped.
	require.NoError(t, env.bankk.SendCoinsUnrestricted(tctx, addr, feeCollector, std.MustParseCoins("400ugnot")))
	require.Zero(t, bnk.GetCoin(addr.Bech32(), "ugnot"))
	require.Equal(t, int64(1000), bnk.GetLockedCoins(addr.Bech32()).AmountOf("ugnot"))
	require.Equal(t, "1000ugnot", reportLocked(t, env, tctx, pkgPath, addr))
}

// TestGetLockedCoinsAllBankerTypesAndRun checks that every BankerType (including
// sub-realm RealmSend) in a crossing MsgCall, and RunSDKBanker in MsgRun, can
// call GetLockedCoins and agree with SDKBanker, and that an invalid address
// fails from realm code.
func TestGetLockedCoinsAllBankerTypesAndRun(t *testing.T) {
	env := setupTestEnv()

	const pkgPath = "gno.land/r/lockedalltypes"
	caller := deployLockedRealm(t, env, "alltypes-caller", pkgPath, `package lockedalltypes

import "chain/banker"

func ReadAll(cur realm, addr address) string {
	ro := banker.NewReadonlyBanker().GetLockedCoins(addr).String()
	os := banker.NewBanker(banker.BankerTypeOriginSend, cur).GetLockedCoins(addr).String()
	rs := banker.NewBanker(banker.BankerTypeRealmSend, cur).GetLockedCoins(addr).String()
	ri := banker.NewBanker(banker.BankerTypeRealmIssue, cur).GetLockedCoins(addr).String()
	sub := banker.NewBanker(banker.BankerTypeRealmSend, cur.Sub("sub")).GetLockedCoins(addr).String()
	if ro != os || ro != rs || ro != ri || ro != sub {
		panic("banker types disagree: " + ro + "/" + os + "/" + rs + "/" + ri + "/" + sub)
	}
	return ro
}
`)

	vestedAddr := fundVesting(t, env, "alltypes-vested", std.MustParseCoins("5000ugnot"), std.VestingSchedule{
		OriginalVesting: std.MustParseCoins("1000ugnot"),
		StartTime:       100,
		EndTime:         200,
	})

	callCtx := env.vmk.MakeGnoTransactionStore(atBlockTime(env, 150).WithGasMeter(store.NewGasMeter(10_000_000)))
	res, err := env.vmk.Call(callCtx, NewMsgCall(caller, nil, pkgPath, "ReadAll", []string{vestedAddr.String()}))
	require.NoError(t, err)
	require.Equal(t, `("500ugnot" string)`+"\n\n", res)
	require.Greater(t, callCtx.GasMeter().GasConsumed(), int64(0))

	// Invalid address must error from realm code too (SDKBanker panics on malformed bech32).
	_, err = env.vmk.QueryEvalString(atBlockTime(env, 150), pkgPath, `ReadAll(cross, "not-an-address")`)
	require.Error(t, err)

	// MsgRun uses RunSDKBanker (embedding *SDKBanker); both Readonly and RealmSend work.
	runCtx := env.vmk.MakeGnoTransactionStore(atBlockTime(env, 150).WithGasMeter(store.NewGasMeter(10_000_000)))
	runOut, err := env.vmk.Run(runCtx, NewMsgRun(caller, std.Coins{}, []*std.MemFile{{
		Name: "script.gno",
		Body: fmt.Sprintf(`package main

import "chain/banker"

func main(cur realm) {
	a := address(%q)
	println(banker.NewReadonlyBanker().GetLockedCoins(a).String())
	println(banker.NewBanker(banker.BankerTypeRealmSend, cur).GetLockedCoins(a).String())
}
`, vestedAddr.String()),
	}}))
	require.NoError(t, err)
	require.Equal(t, "500ugnot\n500ugnot\n", runOut)
}

// TestGetLockedCoinsChargesGas checks that bankerGetLockedCoins runs cleanly
// under a real GasMeter and that its post-call cost grows with the number of
// locked denoms returned (confirming PostSlopeIdx reads the returned slice).
func TestGetLockedCoinsChargesGas(t *testing.T) {
	env := setupTestEnv()

	const pkgPath = "gno.land/r/lockedgas"
	deployer := deployLockedRealm(t, env, "gas-deployer", pkgPath, `package lockedgas

import "chain/banker"

func Count(cur realm, addr address) int {
	return len(banker.NewReadonlyBanker().GetLockedCoins(addr))
}
`)

	one := fundVesting(t, env, "gas-one", std.MustParseCoins("1000d0000"), std.VestingSchedule{
		OriginalVesting: std.MustParseCoins("1000d0000"),
		EndTime:         200,
		Type:            std.VestingDelayed,
	})
	var manyCoins std.Coins
	for i := range 20 {
		manyCoins = append(manyCoins, std.Coin{Denom: fmt.Sprintf("d%04d", i), Amount: 1000})
	}
	many := fundVesting(t, env, "gas-many", manyCoins, std.VestingSchedule{
		OriginalVesting: manyCoins,
		EndTime:         200,
		Type:            std.VestingDelayed,
	})

	measure := func(addr crypto.Address, wantCount string) int64 {
		gm := store.NewGasMeter(10_000_000)
		cctx := env.vmk.MakeGnoTransactionStore(atBlockTime(env, 100).WithGasMeter(gm))
		got, err := env.vmk.Call(cctx, NewMsgCall(deployer, nil, pkgPath, "Count", []string{addr.String()}))
		require.NoError(t, err)
		require.Equal(t, "("+wantCount+" int)\n\n", got)
		return gm.GasConsumed()
	}

	gasOne := measure(one, "1")
	gasMany := measure(many, "20")
	require.Greater(t, gasOne, int64(0))
	require.Greater(t, gasMany, gasOne, "20 locked denoms must cost more gas than 1")
}
