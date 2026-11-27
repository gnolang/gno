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
)

// TestGetLockedCoinsTracksTheSchedule is the realm-visible form of
// std.VestingSchedule.LockedCoins. The numbers are the ones vesting_test.go
// already pins for a 1000ugnot grant from t=100 to t=200, read back through
// SDKBanker and through a realm query at each block time. GetCoins stays the
// whole balance, locked coins included.
func TestGetLockedCoinsTracksTheSchedule(t *testing.T) {
	env := setupTestEnv()

	deployer := crypto.AddressFromPreimage([]byte("locked-deployer"))
	env.acck.SetAccount(env.ctx, env.acck.NewAccountWithAddress(env.ctx, deployer))
	env.bankk.SetCoins(env.ctx, deployer, initialBalance)

	const pkgPath = "gno.land/r/lockedcoins"
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(deployer, pkgPath, []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "locked.gno", Body: `package lockedcoins

import "chain/banker"

func Report(addr address) string {
	return banker.NewReadonlyBanker().GetLockedCoins(addr).String()
}
`},
	})))
	env.vmk.CommitGnoTransactionStore(ctx)

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

func reportLocked(t *testing.T, env testEnv, ctx sdk.Context, pkgPath string, addr crypto.Address) string {
	t.Helper()
	got, err := env.vmk.QueryEvalString(ctx, pkgPath, fmt.Sprintf("Report(%q)", addr.Bech32()))
	require.NoError(t, err)
	return got
}
