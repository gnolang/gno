package auth

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

func TestAccountMapperGetSet(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("some-address"))

	// no account before its created
	acc := env.acck.GetAccount(env.ctx, addr)
	require.Nil(t, acc)

	// create account and check default values
	acc = env.acck.NewAccountWithAddress(env.ctx, addr)
	require.NotNil(t, acc)
	require.Equal(t, addr, acc.GetAddress())
	require.EqualValues(t, nil, acc.GetPubKey())
	require.EqualValues(t, 0, acc.GetSequence())

	// NewAccount doesn't call Set, so it's still nil
	require.Nil(t, env.acck.GetAccount(env.ctx, addr))

	// set some values on the account and save it
	newSequence := uint64(20)
	acc.SetSequence(newSequence)
	env.acck.SetAccount(env.ctx, acc)

	// check the new values
	acc = env.acck.GetAccount(env.ctx, addr)
	require.NotNil(t, acc)
	require.Equal(t, newSequence, acc.GetSequence())
}

func TestAccountMapperRemoveAccount(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	addr1 := crypto.AddressFromPreimage([]byte("addr1"))
	addr2 := crypto.AddressFromPreimage([]byte("addr2"))

	// create accounts
	acc1 := env.acck.NewAccountWithAddress(env.ctx, addr1)
	acc2 := env.acck.NewAccountWithAddress(env.ctx, addr2)

	accSeq1 := uint64(20)
	accSeq2 := uint64(40)

	acc1.SetSequence(accSeq1)
	acc2.SetSequence(accSeq2)
	env.acck.SetAccount(env.ctx, acc1)
	env.acck.SetAccount(env.ctx, acc2)

	acc1 = env.acck.GetAccount(env.ctx, addr1)
	require.NotNil(t, acc1)
	require.Equal(t, accSeq1, acc1.GetSequence())

	// remove one account
	env.acck.RemoveAccount(env.ctx, acc1)
	acc1 = env.acck.GetAccount(env.ctx, addr1)
	require.Nil(t, acc1)

	acc2 = env.acck.GetAccount(env.ctx, addr2)
	require.NotNil(t, acc2)
	require.Equal(t, accSeq2, acc2.GetSequence())
}

func TestAccountKeeperParams(t *testing.T) {
	env := setupTestEnv()

	dp := DefaultParams()
	err := env.acck.SetParams(env.ctx, dp)
	require.NoError(t, err)

	dp2 := env.acck.GetParams(env.ctx)
	require.True(t, dp.Equals(dp2))
}

func TestGasPrice(t *testing.T) {
	env := setupTestEnv()
	gp := std.GasPrice{
		Gas: 100,
		Price: std.Coin{
			Denom:  "token",
			Amount: 10,
		},
	}
	env.gk.SetGasPrice(env.ctx, gp)
	gp2 := env.gk.LastGasPrice(env.ctx)
	require.True(t, gp == gp2)
}

func TestMax(t *testing.T) {
	tests := []struct {
		name     string
		x, y     *big.Int
		expected *big.Int
	}{
		{
			name:     "X is less than Y",
			x:        big.NewInt(5),
			y:        big.NewInt(10),
			expected: big.NewInt(10),
		},
		{
			name:     "X is greater than Y",
			x:        big.NewInt(15),
			y:        big.NewInt(10),
			expected: big.NewInt(15),
		},
		{
			name:     "X is equal to Y",
			x:        big.NewInt(10),
			y:        big.NewInt(10),
			expected: big.NewInt(10),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := maxBig(tc.x, tc.y)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestCalcBlockGasPrice(t *testing.T) {
	gk := GasPriceKeeper{}

	lastGasPrice := std.GasPrice{
		Price: std.Coin{
			Amount: 100,
			Denom:  "atom",
		},
	}
	gasUsed := int64(5000)
	maxGas := int64(10000)
	params := Params{
		TargetGasRatio:            50,
		GasPricesChangeCompressor: 2,
	}

	// Test with normal parameters
	newGasPrice := gk.calcBlockGasPrice(lastGasPrice, gasUsed, maxGas, params)
	expectedAmount := big.NewInt(100)
	num := big.NewInt(gasUsed - maxGas*params.TargetGasRatio/100)
	num.Mul(num, expectedAmount)
	num.Div(num, big.NewInt(maxGas*params.TargetGasRatio/100))
	num.Div(num, big.NewInt(params.GasPricesChangeCompressor))
	expectedAmount.Add(expectedAmount, num)
	require.Equal(t, expectedAmount.Int64(), newGasPrice.Price.Amount)

	// Test with lastGasPrice amount as 0
	lastGasPrice.Price.Amount = 0
	newGasPrice = gk.calcBlockGasPrice(lastGasPrice, gasUsed, maxGas, params)
	require.Equal(t, int64(0), newGasPrice.Price.Amount)

	// Test with TargetGasRatio as 0 (should not change the last price)
	params.TargetGasRatio = 0
	newGasPrice = gk.calcBlockGasPrice(lastGasPrice, gasUsed, maxGas, params)
	require.Equal(t, int64(0), newGasPrice.Price.Amount)

	// Test with gasUsed as 0 (should not change the last price)
	params.TargetGasRatio = 50
	lastGasPrice.Price.Amount = 100
	gasUsed = 0
	newGasPrice = gk.calcBlockGasPrice(lastGasPrice, gasUsed, maxGas, params)
	require.Equal(t, int64(100), newGasPrice.Price.Amount)
}

func TestNewAccountWithUncheckedNumber(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("test-addr-1"))

	// Create account with specific number
	acc := env.acck.NewAccountWithUncheckedNumber(env.ctx, addr, 42)
	require.NotNil(t, acc)
	require.Equal(t, addr, acc.GetAddress())
	require.EqualValues(t, 42, acc.GetAccountNumber())
	require.EqualValues(t, 0, acc.GetSequence())

	// Global counter should be updated to 43
	nextNum := env.acck.GetNextAccountNumber(env.ctx)
	require.EqualValues(t, 43, nextNum)
	// GetNextAccountNumber increments, so next call returns 44
	nextNum2 := env.acck.GetNextAccountNumber(env.ctx)
	require.EqualValues(t, 44, nextNum2)
}

func TestNewAccountWithUncheckedNumber_Zero(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("test-addr-zero"))

	// Account number 0 is valid (first account)
	acc := env.acck.NewAccountWithUncheckedNumber(env.ctx, addr, 0)
	require.NotNil(t, acc)
	require.EqualValues(t, 0, acc.GetAccountNumber())

	// Global counter should be 1
	nextNum := env.acck.GetNextAccountNumber(env.ctx)
	require.EqualValues(t, 1, nextNum)
}

func TestNewAccountWithUncheckedNumber_DoesNotLowerCounter(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()

	// Create some accounts to advance the counter
	for i := range 5 {
		addr := crypto.AddressFromPreimage([]byte(fmt.Sprintf("addr-%d", i)))
		acc := env.acck.NewAccountWithAddress(env.ctx, addr)
		env.acck.SetAccount(env.ctx, acc)
	}
	// Counter is now 5

	// Create account with number lower than counter
	addr := crypto.AddressFromPreimage([]byte("low-number-addr"))
	acc := env.acck.NewAccountWithUncheckedNumber(env.ctx, addr, 2)
	require.NotNil(t, acc)
	require.EqualValues(t, 2, acc.GetAccountNumber())

	// Counter should still be 5, not lowered to 3
	nextNum := env.acck.GetNextAccountNumber(env.ctx)
	require.EqualValues(t, 5, nextNum)
}

func TestNewAccountWithUncheckedNumber_HighNumber(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()

	// Create account with high number (simulating hardfork replay)
	addr := crypto.AddressFromPreimage([]byte("high-number-addr"))
	acc := env.acck.NewAccountWithUncheckedNumber(env.ctx, addr, 1000000)
	require.NotNil(t, acc)
	require.EqualValues(t, 1000000, acc.GetAccountNumber())

	// Counter should jump to 1000001
	nextNum := env.acck.GetNextAccountNumber(env.ctx)
	require.EqualValues(t, 1000001, nextNum)

	// Normal account creation should get 1000002
	addr2 := crypto.AddressFromPreimage([]byte("normal-addr"))
	acc2 := env.acck.NewAccountWithAddress(env.ctx, addr2)
	require.EqualValues(t, 1000002, acc2.GetAccountNumber())
}

// TestNewAccountWithUncheckedNumber_DocumentedUnchecked exercises the
// documented precondition: the keeper does NOT check uniqueness, so calling
// twice with the same accNum but different addresses produces two accounts
// with the same number. Callers must enforce uniqueness upstream (see
// validateSignerInfo in gno.land/pkg/gnoland).
func TestNewAccountWithUncheckedNumber_DocumentedUnchecked(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()

	addrA := crypto.AddressFromPreimage([]byte("a"))
	addrB := crypto.AddressFromPreimage([]byte("b"))

	accA := env.acck.NewAccountWithUncheckedNumber(env.ctx, addrA, 99)
	env.acck.SetAccount(env.ctx, accA)
	accB := env.acck.NewAccountWithUncheckedNumber(env.ctx, addrB, 99)
	env.acck.SetAccount(env.ctx, accB)

	// Both accounts exist, both claim accNum 99. No keeper-level rejection.
	gotA := env.acck.GetAccount(env.ctx, addrA)
	gotB := env.acck.GetAccount(env.ctx, addrB)
	require.NotNil(t, gotA)
	require.NotNil(t, gotB)
	require.EqualValues(t, 99, gotA.GetAccountNumber())
	require.EqualValues(t, 99, gotB.GetAccountNumber())
}

// TestIterateAccountsChargesGas asserts that IterateAccounts propagates
// gas through the gctx it threads to PrefixIterator. Today all
// production query contexts carry an infinite meter, so this mostly
// confirms the wiring works and the charge fires; if a future caller
// sets a bounded meter the enforcement is already in place.
func TestIterateAccountsChargesGas(t *testing.T) {
	t.Parallel()
	env := setupTestEnv()

	// Populate a handful of accounts.
	const n = 5
	for i := 0; i < n; i++ {
		addr := crypto.AddressFromPreimage([]byte(fmt.Sprintf("addr-%d", i)))
		acc := env.acck.NewAccountWithAddress(env.ctx, addr)
		env.acck.SetAccount(env.ctx, acc)
	}

	// Swap in a bounded meter AND a cache-wrapped multistore. Gas is
	// only charged at the cache.Store iterator layer, so we must
	// cache-wrap the multistore for this test. Production tx paths
	// cache-wrap inside runTx; query paths do not (see ADR).
	meter := store.NewGasMeter(1 << 62)
	ctx := env.ctx.
		WithGasMeter(meter).
		WithMultiStore(env.ctx.MultiStore().MultiCacheWrap())

	before := meter.GasConsumed()
	count := 0
	env.acck.IterateAccounts(ctx, func(acc std.Account) bool {
		count++
		return false
	})
	used := meter.GasConsumed() - before

	require.Equal(t, n, count)
	require.Greater(t, used, store.Gas(0),
		"IterateAccounts should consume gas through the threaded gctx")
}
