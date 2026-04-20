package bank

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestKeeper(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	addr2 := crypto.AddressFromPreimage([]byte("addr2"))
	addr3 := crypto.AddressFromPreimage([]byte("addr3"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	env.acck.SetAccount(ctx, acc)
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins()))

	env.bankk.SetCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))

	// Test HasCoins
	require.True(t, env.bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, env.bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 5))))
	require.False(t, env.bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 15))))
	require.False(t, env.bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 5))))

	// Test AddCoins
	env.bankk.AddCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 15)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 25))))

	env.bankk.AddCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 15)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 15), std.NewCoin("foocoin", 25))))

	// Test SubtractCoins
	env.bankk.SubtractCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10)))
	env.bankk.SubtractCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 5)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 15))))

	env.bankk.SubtractCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 11)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 15))))

	env.bankk.SubtractCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 10)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 15))))
	require.False(t, env.bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 1))))

	// Test SendCoins
	env.bankk.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("foocoin", 5)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, env.bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("foocoin", 5))))

	_ = env.bankk.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("foocoin", 50)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, env.bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("foocoin", 5))))

	env.bankk.AddCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 30)))
	env.bankk.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 5)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 20), std.NewCoin("foocoin", 5))))
	require.True(t, env.bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 10))))

	// Test InputOutputCoins
	input1 := NewInput(addr2, std.NewCoins(std.NewCoin("foocoin", 2)))
	output1 := NewOutput(addr, std.NewCoins(std.NewCoin("foocoin", 2)))
	env.bankk.InputOutputCoins(ctx, []Input{input1}, []Output{output1})
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 20), std.NewCoin("foocoin", 7))))
	require.True(t, env.bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 8))))

	inputs := []Input{
		NewInput(addr, std.NewCoins(std.NewCoin("foocoin", 3))),
		NewInput(addr2, std.NewCoins(std.NewCoin("barcoin", 3), std.NewCoin("foocoin", 2))),
	}

	outputs := []Output{
		NewOutput(addr, std.NewCoins(std.NewCoin("barcoin", 1))),
		NewOutput(addr3, std.NewCoins(std.NewCoin("barcoin", 2), std.NewCoin("foocoin", 5))),
	}
	env.bankk.InputOutputCoins(ctx, inputs, outputs)
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 21), std.NewCoin("foocoin", 4))))
	require.True(t, env.bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 7), std.NewCoin("foocoin", 6))))
	require.True(t, env.bankk.GetCoins(ctx, addr3).IsEqual(std.NewCoins(std.NewCoin("barcoin", 2), std.NewCoin("foocoin", 5))))
}

func TestBankKeeper(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx

	bankk := env.bankk

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	addr2 := crypto.AddressFromPreimage([]byte("addr2"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	env.acck.SetAccount(ctx, acc)
	require.True(t, bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins()))

	env.bankk.SetCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10)))
	require.True(t, bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))

	// Test HasCoins
	require.True(t, bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 5))))
	require.False(t, bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 15))))
	require.False(t, bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 5))))

	env.bankk.SetCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 15)))

	// Test SendCoins
	bankk.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("foocoin", 5)))
	require.True(t, bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("foocoin", 5))))

	err := bankk.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("foocoin", 50)))
	require.Error(t, err)
	// Balances of addr and addr2 should stay the same.
	require.True(t, bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("foocoin", 5))))

	env.bankk.AddCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 30)))
	bankk.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 5)))
	require.True(t, bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 20), std.NewCoin("foocoin", 5))))
	require.True(t, bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 10))))

	// validate coins with invalid denoms or negative values cannot be sent
	// NOTE: We must use the Coin literal as the constructor does not allow
	// negative values.
	err = bankk.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.Coin{Denom: "FOOCOIN", Amount: -5}})
	require.Error(t, err)
}

func TestViewKeeper(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	view := NewViewKeeper(env.acck)

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	env.acck.SetAccount(ctx, acc)
	require.True(t, view.GetCoins(ctx, addr).IsEqual(std.NewCoins()))

	env.bankk.SetCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10)))
	require.True(t, view.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))

	// Test HasCoins
	require.True(t, view.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, view.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 5))))
	require.False(t, view.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 15))))
	require.False(t, view.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 5))))
}

// setupSessionCtx creates a master account funded with the given coins,
// installs a session account with the given SpendLimit under it, and
// returns the ctx populated with the session map — the same shape the
// auth ante produces on a session-signed tx.
func setupSessionCtx(t *testing.T, env testEnv, masterCoins, spendLimit std.Coins) (sdk.Context, crypto.Address, std.DelegatedAccount) {
	t.Helper()
	masterAddr := crypto.AddressFromPreimage([]byte("master"))
	masterAcc := env.acck.NewAccountWithAddress(env.ctx, masterAddr)
	masterAcc.SetCoins(masterCoins)
	env.acck.SetAccount(env.ctx, masterAcc)

	sessionPub := crypto.AddressFromPreimage([]byte("session"))
	// ProtoBaseSessionAccount creates a *BaseSessionAccount directly.
	base := std.ProtoBaseSessionAccount().(*std.BaseSessionAccount)
	base.SetMasterAddress(masterAddr)
	base.SetAddress(sessionPub)
	base.SetExpiresAt(0)
	base.SetSpendLimit(spendLimit)
	base.SetSpendReset(env.ctx.BlockTime().Unix())
	env.acck.SetSessionAccount(env.ctx, masterAddr, base)

	sessions := map[crypto.Address]std.DelegatedAccount{masterAddr: base}
	ctx := env.ctx.WithValue(std.SessionAccountsContextKey{}, sessions)
	return ctx, masterAddr, base
}

// TestSessionSendCoinsWithinSpendLimit verifies a session-scoped SendCoins
// succeeds and debits SpendUsed.
func TestSessionSendCoinsWithinSpendLimit(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx, masterAddr, da := setupSessionCtx(t, env,
		std.NewCoins(std.NewCoin("foo", 1000)),
		std.NewCoins(std.NewCoin("foo", 500)))

	recipient := crypto.AddressFromPreimage([]byte("recipient"))
	require.NoError(t, env.bankk.SendCoins(ctx, masterAddr, recipient, std.NewCoins(std.NewCoin("foo", 100))))

	assert.Equal(t, int64(900), env.bankk.GetCoins(ctx, masterAddr).AmountOf("foo"))
	assert.Equal(t, int64(100), env.bankk.GetCoins(ctx, recipient).AmountOf("foo"))
	assert.Equal(t, int64(100), da.GetSpendUsed().AmountOf("foo"))
}

// TestSessionSendCoinsExceedingSpendLimit verifies SendCoins rejects when
// the cumulative session spend would exceed SpendLimit, and no coins move.
func TestSessionSendCoinsExceedingSpendLimit(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx, masterAddr, da := setupSessionCtx(t, env,
		std.NewCoins(std.NewCoin("foo", 1000)),
		std.NewCoins(std.NewCoin("foo", 50)))

	recipient := crypto.AddressFromPreimage([]byte("recipient"))
	err := env.bankk.SendCoins(ctx, masterAddr, recipient, std.NewCoins(std.NewCoin("foo", 100)))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session not allowed")

	// Master balance unchanged; SpendUsed unchanged (DeductSessionSpend
	// does not persist on failure).
	assert.Equal(t, int64(1000), env.bankk.GetCoins(ctx, masterAddr).AmountOf("foo"))
	assert.Equal(t, int64(0), env.bankk.GetCoins(ctx, recipient).AmountOf("foo"))
	assert.Equal(t, int64(0), da.GetSpendUsed().AmountOf("foo"))
}

// TestSessionSendCoinsAccumulates verifies multiple SendCoins calls within
// the same tx ctx accumulate SpendUsed across calls (in-memory pointer is
// shared via the session map).
func TestSessionSendCoinsAccumulates(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx, masterAddr, da := setupSessionCtx(t, env,
		std.NewCoins(std.NewCoin("foo", 1000)),
		std.NewCoins(std.NewCoin("foo", 300)))

	recipient := crypto.AddressFromPreimage([]byte("recipient"))
	require.NoError(t, env.bankk.SendCoins(ctx, masterAddr, recipient, std.NewCoins(std.NewCoin("foo", 100))))
	require.NoError(t, env.bankk.SendCoins(ctx, masterAddr, recipient, std.NewCoins(std.NewCoin("foo", 150))))

	// Third send would put SpendUsed at 350 > 300 → rejected.
	err := env.bankk.SendCoins(ctx, masterAddr, recipient, std.NewCoins(std.NewCoin("foo", 100)))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session not allowed")

	assert.Equal(t, int64(250), da.GetSpendUsed().AmountOf("foo"))
	assert.Equal(t, int64(750), env.bankk.GetCoins(ctx, masterAddr).AmountOf("foo"))
	assert.Equal(t, int64(250), env.bankk.GetCoins(ctx, recipient).AmountOf("foo"))
}

// TestSessionInputOutputCoinsPerInput verifies each input in MsgMultiSend
// is individually session-checked against SpendLimit.
func TestSessionInputOutputCoinsPerInput(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx, masterAddr, da := setupSessionCtx(t, env,
		std.NewCoins(std.NewCoin("foo", 1000)),
		std.NewCoins(std.NewCoin("foo", 200)))

	// A second non-session addr with its own balance — its input is not
	// session-scoped and should pass freely.
	otherAddr := crypto.AddressFromPreimage([]byte("other"))
	otherAcc := env.acck.NewAccountWithAddress(ctx, otherAddr)
	otherAcc.SetCoins(std.NewCoins(std.NewCoin("foo", 500)))
	env.acck.SetAccount(ctx, otherAcc)

	recipient := crypto.AddressFromPreimage([]byte("recipient"))
	inputs := []Input{
		NewInput(masterAddr, std.NewCoins(std.NewCoin("foo", 150))), // within session limit
		NewInput(otherAddr, std.NewCoins(std.NewCoin("foo", 300))),  // non-session
	}
	outputs := []Output{NewOutput(recipient, std.NewCoins(std.NewCoin("foo", 450)))}

	require.NoError(t, env.bankk.InputOutputCoins(ctx, inputs, outputs))
	assert.Equal(t, int64(150), da.GetSpendUsed().AmountOf("foo"))

	// A multi-send whose session input exceeds SpendLimit should fail
	// entirely — no coins move because the error propagates from the
	// session check in the input loop.
	inputs = []Input{
		NewInput(masterAddr, std.NewCoins(std.NewCoin("foo", 100))), // 150+100 = 250 > 200
	}
	outputs = []Output{NewOutput(recipient, std.NewCoins(std.NewCoin("foo", 100)))}
	err := env.bankk.InputOutputCoins(ctx, inputs, outputs)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session not allowed")
	// SpendUsed unchanged from the previous run.
	assert.Equal(t, int64(150), da.GetSpendUsed().AmountOf("foo"))
}

// TestSessionSendCoinsUnrestrictedBypasses verifies that
// SendCoinsUnrestricted skips the session spend hook — this is the escape
// valve for gas collection and storage deposit refunds.
func TestSessionSendCoinsUnrestrictedBypasses(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx, masterAddr, da := setupSessionCtx(t, env,
		std.NewCoins(std.NewCoin("foo", 1000)),
		std.NewCoins(std.NewCoin("foo", 50))) // tight limit

	recipient := crypto.AddressFromPreimage([]byte("recipient"))
	// 100 > 50 limit, but Unrestricted bypasses the session check.
	require.NoError(t, env.bankk.SendCoinsUnrestricted(ctx, masterAddr, recipient, std.NewCoins(std.NewCoin("foo", 100))))
	assert.Equal(t, int64(0), da.GetSpendUsed().AmountOf("foo"))
	assert.Equal(t, int64(900), env.bankk.GetCoins(ctx, masterAddr).AmountOf("foo"))
}

// TestNonSessionSendCoinsNoOp verifies the hook no-ops for txs that don't
// carry a session map in ctx (master-signed, all other non-session txs).
func TestNonSessionSendCoinsNoOp(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx // no SessionAccountsContextKey

	addr := crypto.AddressFromPreimage([]byte("addr"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	acc.SetCoins(std.NewCoins(std.NewCoin("foo", 1000)))
	env.acck.SetAccount(ctx, acc)

	recipient := crypto.AddressFromPreimage([]byte("recipient"))
	require.NoError(t, env.bankk.SendCoins(ctx, addr, recipient, std.NewCoins(std.NewCoin("foo", 500))))
	assert.Equal(t, int64(500), env.bankk.GetCoins(ctx, addr).AmountOf("foo"))
}

// Test SetRestrictedDenoms
func TestSetRestrictedDenoms(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx
	bankk := env.bankk
	prmk := env.prmk
	// Add a single denom
	prmk.SetStrings(ctx, "bank:p:restricted_denoms", []string{"foo"})
	params := bankk.GetParams(ctx)
	require.Contains(t, params.RestrictedDenoms, "foo")

	// Add multiple denoms
	prmk.SetStrings(ctx, "bank:p:restricted_denoms", []string{"goo", "bar"})
	params = bankk.GetParams(ctx)
	require.NotContains(t, params.RestrictedDenoms, "foo")
	require.Contains(t, params.RestrictedDenoms, "goo")
	require.Contains(t, params.RestrictedDenoms, "bar")

	// Add empty list
	prmk.SetStrings(ctx, "bank:p:restricted_denoms", []string{})
	params = bankk.GetParams(ctx)
	require.Empty(t, params.RestrictedDenoms)
}
