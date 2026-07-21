package auth

import (
	"fmt"
	"math"
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
	const (
		maxGas    = int64(3_000_000_000)
		targetGas = int64(2_100_000_000)
	)
	params := Params{
		TargetGasRatio:            70,
		GasPricesChangeCompressor: 10,
		InitialGasPrice: std.GasPrice{
			Gas:   1000,
			Price: std.Coin{Amount: 1, Denom: "ugnot"},
		},
	}
	price := func(amount int64) std.GasPrice {
		return std.GasPrice{
			Gas:   1000,
			Price: std.Coin{Amount: amount, Denom: "ugnot"},
		}
	}

	tests := []struct {
		name     string
		start    int64
		gasUsed  int64
		expected int64
	}{
		{"at target", 10, targetGas, 10},
		{"one below target", 10, targetGas - 1, 9},
		{"one above target", 10, targetGas + 1, 11},
		{"low usage", 10, 420_000_000, 9},
		{"empty block", 10, 0, 9},
		{"floor below target", 1, targetGas - 1, 1},
		{"floor empty block", 1, 0, 1},
		{"floor one above target", 1, targetGas + 1, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gk.calcBlockGasPrice(price(tt.start), tt.gasUsed, maxGas, params)
			require.Equal(t, price(tt.expected), got)
		})
	}

	t.Run("calculated deltas", func(t *testing.T) {
		params := Params{TargetGasRatio: 50, GasPricesChangeCompressor: 2}
		start := std.GasPrice{Gas: 7, Price: std.Coin{Amount: 100, Denom: "atom"}}
		require.Equal(t, int64(125), gk.calcBlockGasPrice(start, 7500, 10000, params).Price.Amount)
		require.Equal(t, int64(75), gk.calcBlockGasPrice(start, 2500, 10000, params).Price.Amount)
	})

	t.Run("disabled", func(t *testing.T) {
		zeroPrice := price(0)
		require.Equal(t, zeroPrice, gk.calcBlockGasPrice(zeroPrice, targetGas+1, maxGas, params))

		disabledParams := params
		disabledParams.TargetGasRatio = 0
		require.Equal(t, price(10), gk.calcBlockGasPrice(price(10), targetGas+1, maxGas, disabledParams))
	})

	t.Run("int64 overflow", func(t *testing.T) {
		require.PanicsWithValue(t, "The min gas price is out of int64 range", func() {
			gk.calcBlockGasPrice(price(math.MaxInt64), targetGas+1, maxGas, params)
		})
	})
}

func TestCalcBlockGasPriceRatchet(t *testing.T) {
	gk := GasPriceKeeper{}
	const (
		maxGas    = int64(3_000_000_000)
		targetGas = int64(2_100_000_000)
	)
	params := Params{
		TargetGasRatio:            70,
		GasPricesChangeCompressor: 10,
		InitialGasPrice: std.GasPrice{
			Gas:   1000,
			Price: std.Coin{Amount: 1, Denom: "ugnot"},
		},
	}
	price := std.GasPrice{
		Gas:   1000,
		Price: std.Coin{Amount: 10, Denom: "ugnot"},
	}

	for _, tt := range []struct {
		name     string
		gasUsed  []int64
		expected []int64
	}{
		{"below then above target", []int64{targetGas - 1, targetGas + 1}, []int64{9, 10, 9, 10}},
		{"above then below target", []int64{targetGas + 1, targetGas - 1}, []int64{11, 10, 11, 10}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			next := price
			for i, expected := range tt.expected {
				next = gk.calcBlockGasPrice(next, tt.gasUsed[i%len(tt.gasUsed)], maxGas, params)
				require.Equal(t, expected, next.Price.Amount)
			}
		})
	}

	for _, tt := range []struct {
		name    string
		gasUsed int64
	}{
		{"low usage", 420_000_000},
		{"empty blocks", 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			next := price
			for block := int64(1); block <= 9; block++ {
				next = gk.calcBlockGasPrice(next, tt.gasUsed, maxGas, params)
				require.Equal(t, int64(10)-block, next.Price.Amount)
			}
			next = gk.calcBlockGasPrice(next, tt.gasUsed, maxGas, params)
			require.Equal(t, int64(1), next.Price.Amount)
		})
	}
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
		addr := crypto.AddressFromPreimage(fmt.Appendf(nil, "addr-%d", i))
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
	for i := range n {
		addr := crypto.AddressFromPreimage(fmt.Appendf(nil, "addr-%d", i))
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

// TestCalcBlockGasPriceUnboundedMaxGas covers the two consensus-param
// spellings of "no block gas bound". Neither carries a congestion signal, so
// neither may move the price.
func TestCalcBlockGasPriceUnboundedMaxGas(t *testing.T) {
	gk := GasPriceKeeper{}
	params := Params{
		TargetGasRatio:            70,
		GasPricesChangeCompressor: 10,
		InitialGasPrice: std.GasPrice{
			Gas:   1000,
			Price: std.Coin{Amount: 1, Denom: "ugnot"},
		},
	}
	price := func(amount int64) std.GasPrice {
		return std.GasPrice{Gas: 1000, Price: std.Coin{Amount: amount, Denom: "ugnot"}}
	}

	// MaxGas == -1 is the sentinel ValidateConsensusParams accepts for "no
	// gas bounds". targetGas floors to -1, so every gasUsed >= 0 used to take
	// the increase branch and the min-1 clamp ratcheted the price up by 1 per
	// block, idle blocks included.
	t.Run("MaxGas -1 does not ratchet", func(t *testing.T) {
		next := price(1000)
		for range 5 {
			next = gk.calcBlockGasPrice(next, 0, -1, params)
			require.Equal(t, int64(1000), next.Price.Amount)
		}
		// Not even a full block moves it: there is no maximum to be full of.
		require.Equal(t, int64(1000), gk.calcBlockGasPrice(price(1000), 1_000_000, -1, params).Price.Amount)
	})

	// maxGas*ratio < 100 makes the target 0 and every branch divides by it.
	// With the default ratio of 70 that is MaxGas 0 and 1. MaxGas == 0 is
	// reachable: getMaximumBlockGas maps it to an infinite gas meter, so
	// blocks consume gas and EndBlock is reached with gasUsed > 0.
	t.Run("zero target does not panic", func(t *testing.T) {
		for _, maxGas := range []int64{0, 1} {
			for _, gasUsed := range []int64{0, 1, 1_000_000} {
				require.NotPanics(t, func() {
					got := gk.calcBlockGasPrice(price(10), gasUsed, maxGas, params)
					require.Equal(t, int64(10), got.Price.Amount)
				}, "maxGas=%d gasUsed=%d", maxGas, gasUsed)
			}
		}
	})

	// The smallest MaxGas that still yields a target keeps working.
	t.Run("smallest usable MaxGas still adjusts", func(t *testing.T) {
		require.Equal(t, int64(9), gk.calcBlockGasPrice(price(10), 0, 2, params).Price.Amount)
	})
}

// TestCalcBlockGasPriceZeroInitialPrice pins the floor at 1 even when
// InitialGasPrice is 0. Params.Validate accepts that value, and a stored price
// of 0 is absorbing: calcBlockGasPrice reads it as "dynamic pricing disabled"
// and returns early, so the price could never recover.
func TestCalcBlockGasPriceZeroInitialPrice(t *testing.T) {
	gk := GasPriceKeeper{}
	const (
		maxGas    = int64(3_000_000_000)
		targetGas = int64(2_100_000_000)
	)
	params := Params{
		TargetGasRatio:            70,
		GasPricesChangeCompressor: 10,
		InitialGasPrice: std.GasPrice{
			Gas:   1000,
			Price: std.Coin{Amount: 0, Denom: "ugnot"},
		},
	}
	require.NoError(t, func() error {
		p := DefaultParams()
		p.InitialGasPrice = params.InitialGasPrice
		return p.Validate()
	}(), "Params.Validate accepts a zero InitialGasPrice, so calc must cope with it")

	next := std.GasPrice{Gas: 1000, Price: std.Coin{Amount: 3, Denom: "ugnot"}}
	for range 10 {
		next = gk.calcBlockGasPrice(next, 0, maxGas, params)
		require.GreaterOrEqual(t, next.Price.Amount, int64(1))
	}
	require.Equal(t, int64(1), next.Price.Amount)

	// Still responsive: from the floor, a full block raises the price again.
	up := gk.calcBlockGasPrice(next, targetGas+1, maxGas, params)
	require.Equal(t, int64(2), up.Price.Amount)
}
