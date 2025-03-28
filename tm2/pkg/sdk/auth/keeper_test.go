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

func TestSessionMapperGetSet(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("some-address"))
	pubKey := crypto.GenPrivKeyEd25519().PubKey()

	// no session before it's created
	sess := env.acck.GetSession(env.ctx, pubKey)
	require.Nil(t, sess)

	// create session and set it
	session := std.NewBaseSession(addr, pubKey, 0, true)
	env.acck.SetSession(env.ctx, session)

	// check the values
	sess = env.acck.GetSession(env.ctx, pubKey)
	require.NotNil(t, sess)
	require.True(t, sess.GetPubKey().Equals(pubKey))
	require.Equal(t, addr, sess.GetAccountAddress())
	require.True(t, sess.IsMaster())
	require.Equal(t, uint64(0), sess.GetSequence())

	// modify sequence and update
	sess.SetSequence(5)
	env.acck.SetSession(env.ctx, sess)

	// verify update
	sess = env.acck.GetSession(env.ctx, pubKey)
	require.Equal(t, uint64(5), sess.GetSequence())
}

func TestSessionMapperRemoveSession(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("some-address"))
	pubKey1 := crypto.GenPrivKeyEd25519().PubKey()
	pubKey2 := crypto.GenPrivKeyEd25519().PubKey()

	// create two sessions
	session1 := std.NewBaseSession(addr, pubKey1, 1, true)
	session2 := std.NewBaseSession(addr, pubKey2, 2, false)

	env.acck.SetSession(env.ctx, session1)
	env.acck.SetSession(env.ctx, session2)

	// verify both exist
	sess1 := env.acck.GetSession(env.ctx, pubKey1)
	sess2 := env.acck.GetSession(env.ctx, pubKey2)
	require.NotNil(t, sess1)
	require.NotNil(t, sess2)

	// remove one session
	env.acck.RemoveSession(env.ctx, pubKey1)

	// verify first session is gone but second remains
	sess1 = env.acck.GetSession(env.ctx, pubKey1)
	sess2 = env.acck.GetSession(env.ctx, pubKey2)
	require.Nil(t, sess1)
	require.NotNil(t, sess2)
}

func TestGetAllSessions(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("some-address"))
	pubKey1 := crypto.GenPrivKeyEd25519().PubKey()
	pubKey2 := crypto.GenPrivKeyEd25519().PubKey()
	pubKey3 := crypto.GenPrivKeyEd25519().PubKey()

	// create three sessions
	session1 := std.NewBaseSession(addr, pubKey1, 1, true)
	session2 := std.NewBaseSession(addr, pubKey2, 2, false)
	session3 := std.NewBaseSession(addr, pubKey3, 3, false)

	env.acck.SetSession(env.ctx, session1)
	env.acck.SetSession(env.ctx, session2)
	env.acck.SetSession(env.ctx, session3)

	// get all sessions
	sessions := env.acck.GetAllSessions(env.ctx)
	require.Len(t, sessions, 3)

	// verify each session exists in the result
	found := make(map[string]bool)
	for _, sess := range sessions {
		found[string(sess.GetPubKey().Bytes())] = true

		// verify session properties
		require.Equal(t, addr, sess.GetAccountAddress())
		switch {
		case sess.GetPubKey().Equals(pubKey1):
			require.True(t, sess.IsMaster())
			require.Equal(t, uint64(1), sess.GetSequence())
		case sess.GetPubKey().Equals(pubKey2):
			require.False(t, sess.IsMaster())
			require.Equal(t, uint64(2), sess.GetSequence())
		case sess.GetPubKey().Equals(pubKey3):
			require.False(t, sess.IsMaster())
			require.Equal(t, uint64(3), sess.GetSequence())
		}
	}
	require.Len(t, found, 3)
}

func TestIterateSessions(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("some-address"))
	pubKey1 := crypto.GenPrivKeyEd25519().PubKey()
	pubKey2 := crypto.GenPrivKeyEd25519().PubKey()

	// create two sessions
	session1 := std.NewBaseSession(addr, pubKey1, 1, true)
	session2 := std.NewBaseSession(addr, pubKey2, 2, false)

	env.acck.SetSession(env.ctx, session1)
	env.acck.SetSession(env.ctx, session2)

	// test iteration
	count := 0
	env.acck.IterateSessions(env.ctx, func(sess std.Session) bool {
		count++
		require.Equal(t, addr, sess.GetAccountAddress())
		if sess.GetPubKey().Equals(pubKey1) {
			require.True(t, sess.IsMaster())
			require.Equal(t, uint64(1), sess.GetSequence())
		} else {
			require.False(t, sess.IsMaster())
			require.Equal(t, uint64(2), sess.GetSequence())
		}
		return false // continue iteration
	})
	require.Equal(t, 2, count)

	// test early termination
	count = 0
	env.acck.IterateSessions(env.ctx, func(sess std.Session) bool {
		count++
		return true // stop after first session
	})
	require.Equal(t, 1, count)
}
