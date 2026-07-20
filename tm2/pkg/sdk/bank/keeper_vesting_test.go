package bank

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func ugnotCoins(n int64) std.Coins {
	return std.NewCoins(std.NewCoin("ugnot", n))
}

// setupVestingAccount stores a continuous vesting account at addr: `total`
// ugnot, all of it vesting linearly from t=100 to t=200.
func setupVestingAccount(t *testing.T, env testEnv, addr crypto.Address, total int64) {
	t.Helper()

	baseAcc := std.NewBaseAccount(addr, ugnotCoins(total), nil, 0, 0)
	cva, err := std.NewContinuousVestingAccount(baseAcc, std.VestingSchedule{
		OriginalVesting: ugnotCoins(total),
		StartTime:       100,
		EndTime:         200,
	})
	require.NoError(t, err)
	env.acck.SetAccount(env.ctx, cva)
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

	// Schedule not complete → still a vesting account.
	_, isVesting := env.acck.GetAccount(ctx, fromAddr).(std.VestingAccount)
	require.True(t, isVesting, "account must remain a vesting account mid-schedule")

	// --- After end (t=250): fully vested, nothing locked. ---
	ctx = atTime(env, 250)
	err = env.bankk.SendCoins(ctx, fromAddr, toAddr, ugnotCoins(500))
	require.NoError(t, err, "all coins must be spendable once fully vested")
	require.Equal(t, int64(0), env.bankk.GetCoins(ctx, fromAddr).AmountOf("ugnot"))
	require.Equal(t, int64(1000), env.bankk.GetCoins(ctx, toAddr).AmountOf("ugnot"))

	// The fully-vested account is upgraded to a plain BaseAccount.
	_, isBase := env.acck.GetAccount(ctx, fromAddr).(*std.BaseAccount)
	require.True(t, isBase, "fully-vested account must be upgraded to BaseAccount")
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
	require.Equal(t, int64(100), env.bankk.GetCoins(ctx, toAddr).AmountOf("ugnot"))
}
