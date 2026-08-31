package bank

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func ugnotCoins(n int64) std.Coins {
	return std.NewCoins(std.NewCoin("ugnot", n))
}

// setupVestingSchedule is the schedule setupVestingAccount installs: `total`
// ugnot vesting linearly from t=100 to t=200.
func setupVestingSchedule(total int64) std.VestingSchedule {
	return std.VestingSchedule{
		OriginalVesting: ugnotCoins(total),
		StartTime:       100,
		EndTime:         200,
	}
}

// setupVestingAccount stores a continuous vesting account at addr: `total`
// ugnot, all of it vesting linearly from t=100 to t=200.
func setupVestingAccount(t *testing.T, env testEnv, addr crypto.Address, total int64) {
	t.Helper()

	baseAcc := std.NewBaseAccount(addr, ugnotCoins(total), nil, 0, 0)
	baseAcc.Vesting = setupVestingSchedule(total)
	env.acck.SetAccount(env.ctx, baseAcc)
}

// atTime returns a copy of env.ctx whose block time is the given unix second.
func atTime(env testEnv, unix int64) sdk.Context {
	return env.ctx.WithBlockHeader(&bft.Header{
		ChainID: env.ctx.ChainID(),
		Time:    time.Unix(unix, 0),
	})
}

// TestBankKeeper_VestingSpendEnforcement exercises the runtime vesting
// enforcement in SubtractCoins through the public SendCoins path: locked
// coins cannot be sent, the vested portion can, and once the schedule
// completes the account is upgraded to a plain BaseAccount.
func TestBankKeeper_VestingSpendEnforcement(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()

	fromAddr := crypto.AddressFromPreimage([]byte("vesting-from"))
	toAddr := crypto.AddressFromPreimage([]byte("vesting-to"))
	setupVestingAccount(t, env, fromAddr, 1000)

	// --- Before start (t=50): all 1000 locked, 0 spendable. ---
	ctx := atTime(env, 50)
	err := env.bankk.SendCoins(ctx, fromAddr, toAddr, ugnotCoins(1))
	require.Error(t, err, "spending locked coins before vesting start must be rejected")
	require.Equal(t, int64(1000), env.bankk.GetCoins(ctx, fromAddr).AmountOf("ugnot"))
	require.Equal(t, int64(0), env.bankk.GetCoins(ctx, toAddr).AmountOf("ugnot"))

	// --- Halfway (t=150): 500 vested/spendable, 500 locked. ---
	ctx = atTime(env, 150)

	// Above the spendable amount → rejected, balances untouched. A 1000-coin
	// balance rules out ErrInsufficientCoins, so only the vesting check can
	// reject a 600-coin send here.
	err = env.bankk.SendCoins(ctx, fromAddr, toAddr, ugnotCoins(600))
	require.Error(t, err, "spending more than the vested amount must be rejected")
	require.Equal(t, int64(1000), env.bankk.GetCoins(ctx, fromAddr).AmountOf("ugnot"))
	require.Equal(t, int64(0), env.bankk.GetCoins(ctx, toAddr).AmountOf("ugnot"))

	// Within the spendable amount → allowed.
	err = env.bankk.SendCoins(ctx, fromAddr, toAddr, ugnotCoins(500))
	require.NoError(t, err, "spending the vested amount must be allowed")
	require.Equal(t, int64(500), env.bankk.GetCoins(ctx, fromAddr).AmountOf("ugnot"))
	require.Equal(t, int64(500), env.bankk.GetCoins(ctx, toAddr).AmountOf("ugnot"))

	// Spendable is now exhausted: remaining balance (500) equals the still-locked
	// amount (500), so nothing more can be sent until more vests.
	err = env.bankk.SendCoins(ctx, fromAddr, toAddr, ugnotCoins(1))
	require.Error(t, err, "no spendable coins remain until more vests")

	// Mid-schedule the lock is still live.
	require.False(t, env.acck.GetAccount(ctx, fromAddr).LockedCoins(ctx.BlockTime()).IsZero(),
		"coins must still be locked mid-schedule")

	// --- After end (t=250): fully vested, nothing locked. ---
	ctx = atTime(env, 250)
	err = env.bankk.SendCoins(ctx, fromAddr, toAddr, ugnotCoins(500))
	require.NoError(t, err, "all coins must be spendable once fully vested")
	require.Equal(t, int64(0), env.bankk.GetCoins(ctx, fromAddr).AmountOf("ugnot"))
	require.Equal(t, int64(1000), env.bankk.GetCoins(ctx, toAddr).AmountOf("ugnot"))

	// The schedule stays on the account once it has elapsed. It locks nothing,
	// so there is nothing to collapse and nothing to rewrite.
	acc := env.acck.GetAccount(ctx, fromAddr)
	require.True(t, acc.LockedCoins(ctx.BlockTime()).IsZero(),
		"an elapsed schedule must lock nothing")
	require.False(t, acc.GetVesting().IsZero(),
		"the schedule is kept, not erased")
}

// TestBankKeeper_VestingUnrestrictedBypass documents the deliberate policy that
// unrestricted transfers (gas payments, storage refunds) bypass the vesting
// lock even while a regular SendCoins of the same amount is rejected.
func TestBankKeeper_VestingUnrestrictedBypass(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()

	fromAddr := crypto.AddressFromPreimage([]byte("vesting-gas-from"))
	toAddr := crypto.AddressFromPreimage([]byte("vesting-gas-to"))
	setupVestingAccount(t, env, fromAddr, 1000)

	// Before start: everything is locked for regular transfers.
	ctx := atTime(env, 50)
	require.Error(t, env.bankk.SendCoins(ctx, fromAddr, toAddr, ugnotCoins(100)),
		"regular transfer of locked coins must be rejected")

	// But unrestricted transfers bypass the lock.
	err := env.bankk.SendCoinsUnrestricted(ctx, fromAddr, toAddr, ugnotCoins(100))
	require.NoError(t, err, "unrestricted transfers must bypass the vesting lock")
	require.Equal(t, int64(900), env.bankk.GetCoins(ctx, fromAddr).AmountOf("ugnot"))

	// Bypassing the lock must not clear it. This is the path every transaction
	// takes to pay gas, and it rewrites the account object, so a schedule lost
	// here would be lost on the first fee any vesting account ever paid.
	require.Equal(t, setupVestingSchedule(1000), env.acck.GetAccount(ctx, fromAddr).GetVesting(),
		"paying through the unrestricted path must leave the schedule intact")
	require.Equal(t, int64(100), env.bankk.GetCoins(ctx, toAddr).AmountOf("ugnot"))
}

// An elapsed schedule is left on the account rather than collapsed into a
// different account type, so a debit past EndTime rewrites nothing. The old
// design swapped the account object out at that point, which meant a debit that
// then failed its affordability check had still rewritten it.
func TestAnElapsedScheduleIsLeftAloneByADebit(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("vested"))
	vesting := std.Coins{{Denom: "ugnot", Amount: 100}}

	cva := std.NewBaseAccount(addr, vesting, nil, 0, 0)
	cva.Vesting = std.VestingSchedule{OriginalVesting: vesting, StartTime: 100, EndTime: 200}
	env.acck.SetAccount(env.ctx, cva)
	require.NoError(t, env.bankk.SetCoins(env.ctx, addr, vesting))

	ctx := atTime(env, 500) // past EndTime: the schedule has elapsed
	before := env.acck.GetAccount(ctx, addr)
	require.True(t, before.LockedCoins(ctx.BlockTime()).IsZero(), "nothing may remain locked")

	// More than is held: the debit must fail and touch nothing.
	require.Error(t, env.bankk.SubtractCoins(ctx, addr, std.Coins{{Denom: "ugnot", Amount: 500}}))
	require.Equal(t, before.GetVesting(), env.acck.GetAccount(ctx, addr).GetVesting(),
		"a failed debit must not have altered the schedule")
	require.Equal(t, int64(100), env.bankk.GetCoin(ctx, addr, "ugnot"))

	// A debit that succeeds keeps the schedule too. This one touches only a
	// SPLIT-tier denom, so setAccountTierCoins never writes the account object --
	// which is what makes an unexpected rewrite visible here at all.
	env.bankk.setSplitBalance(ctx, addr, testRealmDenom, 50)
	require.NoError(t, env.bankk.SubtractCoins(ctx, addr,
		std.Coins{{Denom: testRealmDenom, Amount: 10}}))
	after := env.acck.GetAccount(ctx, addr)
	require.Equal(t, before.GetVesting(), after.GetVesting(),
		"a successful debit must not have altered the schedule either")
	require.Equal(t, int64(40), env.bankk.GetCoin(ctx, addr, testRealmDenom))
	require.Equal(t, int64(100), env.bankk.GetCoin(ctx, addr, "ugnot"), "untouched")
}

// TestBurnIsBlockedByAVestingLock records that a realm cannot always remove its own
// coin. BurnCoins debits through SubtractCoins, whose vesting check is deliberately
// tier-agnostic, so a genesis schedule naming a realm denom locks it against the
// issuer too. That is reachable: applyBalance checks the schedule against the whole
// genesis amount, which spans both tiers, then SetCoins moves the non-gas denoms to
// their own keys, leaving OriginalVesting naming a split-tier denom.
//
// Whether an issuer should be able to burn past a lock is a policy question and is
// left alone; this pins what the code does, and that a refused burn moves nothing.
func TestBurnIsBlockedByAVestingLock(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("realm-vester-burn"))
	locked := std.Coins{{Denom: testRealmDenom, Amount: 1000}}

	cva := std.NewBaseAccount(addr, locked, nil, 0, 0)
	cva.Vesting = std.VestingSchedule{OriginalVesting: locked, StartTime: 100, EndTime: 200}
	env.acck.SetAccount(env.ctx, cva)
	require.NoError(t, env.bankk.SetCoins(env.ctx, addr, locked))
	// Seeded, so the refusal below cannot come from nextSupply finding no counter.
	env.bankk.RecomputeSupply(env.ctx)
	require.Equal(t, int64(1000), env.bankk.TotalSupply(env.ctx, testRealmDenom),
		"precondition: the supply must be recorded, or the burn is refused for the wrong reason")

	ctx := atTime(env, 100) // nothing vested yet
	err := env.bankk.BurnCoins(ctx, addr, locked)
	// Which refusal, not merely that it failed. std.Err* renders its detail only
	// under %+v, so Error() is the bare type name — which is exactly what
	// discriminates here: an unaffordable debit would say "insufficient coins error".
	require.EqualError(t, err, "vesting locked coins error",
		"a burn of a fully locked denom must be refused as locked, not as unaffordable")
	require.Equal(t, int64(1000), env.bankk.TotalSupply(ctx, testRealmDenom),
		"a refused burn must not move the counter")
	require.Equal(t, int64(1000), env.bankk.GetCoin(ctx, addr, testRealmDenom),
		"a refused burn must not move the balance")

	// Once vested the same burn succeeds, so this cannot pass against a BurnCoins
	// that refuses everything.
	vested := atTime(env, 300)
	require.NoError(t, env.bankk.BurnCoins(vested, addr, locked))
	require.Zero(t, env.bankk.TotalSupply(vested, testRealmDenom))
}

// A schedule locks the denoms it names and nothing else. An account that has one
// but cannot afford some other denom is short of funds, not locked, and the two
// are different errors: only the first tells the holder to wait.
//
// This is what the lockedAmt==0 shortcut decides. Without it every denom would go
// through the spendable comparison, which reports a shortfall as a vesting lock.
func TestAnUnlockedDenomReportsAShortfallNotALock(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("partly-locked"))
	to := crypto.AddressFromPreimage([]byte("partly-locked-to"))

	// ugnot is fully locked; the realm denom is not named by the schedule at all.
	acc := std.NewBaseAccount(addr, ugnotCoins(1000), nil, 0, 0)
	acc.Vesting = std.VestingSchedule{
		OriginalVesting: ugnotCoins(1000),
		StartTime:       100,
		EndTime:         200,
	}
	env.acck.SetAccount(env.ctx, acc)
	require.NoError(t, env.bankk.SetCoins(env.ctx, addr, std.Coins{
		{Denom: testRealmDenom, Amount: 10},
		{Denom: testAccountDenom, Amount: 1000},
	}))

	ctx := atTime(env, 50) // nothing vested

	// More of the unlocked denom than is held: a shortfall.
	err := env.bankk.SendCoins(ctx, addr, to, std.Coins{{Denom: testRealmDenom, Amount: 50}})
	require.EqualError(t, err, "insufficient coins error",
		"an unlocked denom the account cannot afford is a shortfall, not a lock")

	// And the locked denom still reports as locked, so this cannot pass by
	// reporting a shortfall for everything.
	err = env.bankk.SendCoins(ctx, addr, to, std.Coins{{Denom: testAccountDenom, Amount: 1}})
	require.EqualError(t, err, "vesting locked coins error")
}

// A schedule can lock more than the account still holds: fees and storage
// deposits debit through the unrestricted path, which never consults it. The
// spendable figure is clamped so the refusal reports nothing available rather
// than a negative amount.
func TestSpendingIntoTheLockedPortionReportsZeroNotANegative(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("overspent"))
	to := crypto.AddressFromPreimage([]byte("overspent-to"))
	setupVestingAccount(t, env, addr, 1000)

	ctx := atTime(env, 50) // nothing vested: all 1000 locked

	// The unrestricted path spends into the locked portion, exactly as a gas
	// payment does. Now 900 is held while 1000 is locked.
	require.NoError(t, env.bankk.SendCoinsUnrestricted(ctx, addr, to, ugnotCoins(100)))
	require.Equal(t, int64(900), env.bankk.GetCoin(ctx, addr, "ugnot"),
		"precondition: the account must hold less than its schedule locks")

	err := env.bankk.SendCoins(ctx, addr, to, ugnotCoins(1))
	require.Error(t, err)
	// The detail lives under %+v; Error() is only the type name.
	assert.Contains(t, fmt.Sprintf("%+v", err), "0ugnot <",
		"the refusal must report nothing spendable, not a negative amount")
	assert.NotContains(t, fmt.Sprintf("%+v", err), "-100",
		"a negative spendable must never reach the message")
}
