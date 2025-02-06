package auth

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
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
