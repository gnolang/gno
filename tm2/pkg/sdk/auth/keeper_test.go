package auth

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
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

/*
TestSessionManagement tests the basic CRUD operations and lifecycle of a session.

Flow:

 1. Session Creation and Initial State
    [Create Session] -> [Verify Not In Store]

 2. Session Storage and Retrieval
    [Store Session] -> [Retrieve Session] -> [Verify Fields]

 3. Session Update
    [Update Sequence] -> [Store Updated] -> [Verify Update]

 4. Session Querying
    [Get All Sessions] -> [Verify Count & Content]

 5. Session Deletion
    [Remove Session] -> [Verify Removal]

 6. Error Handling
    [Query Non-existent] -> [Verify Error]

This test ensures:
- Proper session initialization
- Accurate storage and retrieval of session data
- Correct sequence number management
- Successful session removal
- Appropriate error handling for missing sessions
*/
func TestSessionManagement(t *testing.T) {
	env := setupTestEnv()
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()

	// Create a new session
	session := std.NewBaseSession(addr, pubKey, 0)
	require.NotNil(t, session)

	// Test initial state
	storedSession := env.acck.GetSession(env.ctx, pubKey)
	require.Nil(t, storedSession, "Session should not exist before being set")

	// Test setting session
	env.acck.SetSession(env.ctx, session)
	storedSession = env.acck.GetSession(env.ctx, pubKey)
	require.NotNil(t, storedSession)
	require.Equal(t, pubKey, storedSession.GetPubKey())
	require.Equal(t, addr, storedSession.GetAddress())
	require.Equal(t, uint64(0), storedSession.GetSequence())

	// Test updating sequence
	session.SetSequence(1)
	env.acck.SetSession(env.ctx, session)
	storedSession = env.acck.GetSession(env.ctx, pubKey)
	require.Equal(t, uint64(1), storedSession.GetSequence())

	// Test GetSequence
	seq, err := env.acck.GetSequence(env.ctx, pubKey)
	require.NoError(t, err)
	require.Equal(t, uint64(1), seq)

	// Test GetAllSessions
	sessions := env.acck.GetAllSessions(env.ctx)
	require.Len(t, sessions, 1)
	require.Equal(t, pubKey, sessions[0].GetPubKey())

	// Test removing session
	env.acck.RemoveSession(env.ctx, session)
	storedSession = env.acck.GetSession(env.ctx, pubKey)
	require.Nil(t, storedSession)

	// Test GetSequence for non-existent session
	_, err = env.acck.GetSequence(env.ctx, pubKey)
	require.Error(t, err)
}

/*
TestSessionIteration tests the iteration functionality over multiple sessions.

Flow:

 1. Setup Multiple Sessions
    [Create Session 1] -> [Create Session 2] -> ... -> [Create Session 5]
    |
    v
    [Store Each Session]

 2. Complete Iteration
    [Start Iterator] -> [Process All Sessions] -> [Verify Total Count]

 3. Early Termination
    [Start Iterator] -> [Process 3 Sessions] -> [Stop] -> [Verify Partial Count]

This test ensures:
- Correct handling of multiple sessions
- Proper iteration over all stored sessions
- Working early termination mechanism
- Accurate session counting
- Proper cleanup after iteration

Key aspects tested:
- Iterator initialization
- Session enumeration
- Early exit conditions
- Resource cleanup
*/
func TestSessionIteration(t *testing.T) {
	env := setupTestEnv()

	// Create multiple sessions
	numSessions := 5
	sessions := make([]std.Session, numSessions)
	for i := range numSessions {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey()
		addr := pubKey.Address()
		session := std.NewBaseSession(addr, pubKey, uint64(i))
		sessions[i] = session
		env.acck.SetSession(env.ctx, session)
	}

	var count int
	env.acck.IterateSessions(env.ctx, func(sess std.Session) bool {
		require.NotNil(t, sess)
		count++
		return false
	})
	require.Equal(t, numSessions, count)

	// early termination
	count = 0
	env.acck.IterateSessions(env.ctx, func(sess std.Session) bool {
		count++
		return count >= 3 // Stop after processing 3 sessions
	})
	require.Equal(t, 3, count)
}

// XXX: test sessions
// XXX: test account creation flows (especially multistep)
// XXX: test session validity
