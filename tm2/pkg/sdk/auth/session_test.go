package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// setupSessionEnv creates a test environment with a funded master account
// and a block time set to avoid genesis special-casing.
func setupSessionEnv(t *testing.T) (testEnv, sdk.AnteHandler, crypto.PrivKey, crypto.PubKey, crypto.Address) {
	t.Helper()

	env := setupTestEnv()
	anteHandler := NewAnteHandler(env.acck, env.bankk, DefaultSigVerificationGasConsumer, AnteOptions{VerifyGenesisSignatures: false})

	// Create and fund master account.
	masterPriv, masterPub, masterAddr := tu.KeyTestPubAddr()
	masterAcc := env.acck.NewAccountWithAddress(env.ctx, masterAddr)
	masterAcc.SetCoins(tu.NewTestCoins())
	masterAcc.SetPubKey(masterPub)
	env.acck.SetAccount(env.ctx, masterAcc)

	// Set block time > 0 to avoid genesis special casing.
	now := time.Now()
	env.ctx = env.ctx.WithBlockHeader(&bft.Header{
		ChainID: env.ctx.ChainID(),
		Height:  1,
		Time:    now,
	})

	return env, anteHandler, masterPriv, masterPub, masterAddr
}

// sessionSpendLimit returns a spend limit large enough to cover test fees.
func sessionSpendLimit() std.Coins {
	return std.Coins{std.NewCoin("atom", 10000000)}
}

// createSessionDirect creates a session account directly via the keeper (not via handler).
func createSessionDirect(t *testing.T, env testEnv, masterAddr crypto.Address, sessionPub crypto.PubKey, expiresAt int64) std.Account {
	t.Helper()

	sa := env.acck.NewSessionAccount(env.ctx, masterAddr, sessionPub)
	da := sa.(std.DelegatedAccount)
	da.SetExpiresAt(expiresAt)
	da.SetSpendLimit(sessionSpendLimit())
	da.SetSpendReset(env.ctx.BlockTime().Unix())
	env.acck.SetSessionAccount(env.ctx, masterAddr, sa)
	return sa
}

func TestSessionBasicAuth(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	// Create session key pair.
	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()

	// Create session account with 1-hour expiry.
	sa := createSessionDirect(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600)
	sessionAccNum := sa.GetAccountNumber()
	sessionSeq := sa.GetSequence()

	// Build a message signed by the session key.
	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()

	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, sessionSeq, fee)

	// Should pass.
	newCtx, result, abort := anteHandler(ctx, tx, false)
	require.False(t, abort, "expected session tx to pass, got: %s", result.Log)
	require.True(t, result.IsOK(), result.Log)

	// Check session sequence was incremented.
	updatedSA := env.acck.GetSessionAccount(newCtx, masterAddr, sessionAddr)
	require.NotNil(t, updatedSA)
	assert.Equal(t, sessionSeq+1, updatedSA.GetSequence(), "session sequence should be incremented")

	// Check session accounts are set in context.
	saMap, ok := newCtx.Value(std.SessionAccountsContextKey{}).(map[crypto.Address]std.DelegatedAccount)
	require.True(t, ok, "session accounts should be in context")
	_, found := saMap[masterAddr]
	assert.True(t, found, "session account for master should be in context map")
}

func TestSessionExpiry(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()

	// Create an already-expired session (expires 1 second ago).
	sa := createSessionDirect(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()-1)
	sessionAccNum := sa.GetAccountNumber()

	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	checkInvalidTx(t, anteHandler, ctx, tx, false, std.SessionExpiredError{})
}

func TestSessionUnknown(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	// Use a session address that was never created.
	sessionPriv, _, sessionAddr := tu.KeyTestPubAddr()

	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, 999, 0, fee)

	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})
}

func TestSessionMasterStillWorks(t *testing.T) {
	t.Parallel()

	env, anteHandler, masterPriv, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	// Create a session (should not interfere with master signing).
	_, sessionPub, _ := tu.KeyTestPubAddr()
	createSessionDirect(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600)

	// Master account signs a normal tx (no SessionAddr).
	masterAcc := env.acck.GetAccount(ctx, masterAddr)
	masterAccNum := masterAcc.GetAccountNumber()
	masterSeq := masterAcc.GetSequence()

	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewTestTx(t, ctx.ChainID(), msgs, []crypto.PrivKey{masterPriv}, []uint64{masterAccNum}, []uint64{masterSeq}, fee)

	checkValidTx(t, anteHandler, ctx, tx, false)
}

func TestSessionReplayProtection(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()

	sa := createSessionDirect(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600)
	sessionAccNum := sa.GetAccountNumber()

	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	// First tx should pass.
	newCtx, result, abort := anteHandler(ctx, tx, false)
	require.False(t, abort, result.Log)
	require.True(t, result.IsOK(), result.Log)

	// Replay the same tx (same sequence=0) against updated context — should fail.
	// The sequence was incremented to 1, so signing with seq=0 is invalid.
	checkInvalidTx(t, anteHandler, newCtx, tx, false, std.UnauthorizedError{})
}

func TestSessionRevoke(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()

	// Create session via handler.
	h := NewHandler(env.acck, env.gk)
	createMsg := MsgCreateSession{
		Creator:    masterAddr,
		SessionKey: sessionPub,
		ExpiresAt:  ctx.BlockTime().Unix() + 3600,
		SpendLimit: sessionSpendLimit(),
	}
	res := h.Process(ctx, createMsg)
	require.True(t, res.IsOK(), res.Log)

	// Verify session works.
	sa := env.acck.GetSessionAccount(ctx, masterAddr, sessionAddr)
	require.NotNil(t, sa, "session should exist after creation")
	sessionAccNum := sa.GetAccountNumber()

	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// Revoke session via handler.
	revokeMsg := MsgRevokeSession{
		Creator:    masterAddr,
		SessionKey: sessionPub,
	}
	res = h.Process(ctx, revokeMsg)
	require.True(t, res.IsOK(), res.Log)

	// Verify session no longer exists.
	sa = env.acck.GetSessionAccount(ctx, masterAddr, sessionAddr)
	require.Nil(t, sa, "session should be removed after revocation")

	// Verify using revoked session fails.
	tx2 := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)
	checkInvalidTx(t, anteHandler, ctx, tx2, false, std.UnauthorizedError{})
}

func TestSessionRevokeAll(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	h := NewHandler(env.acck, env.gk)

	// Create multiple sessions.
	var sessionPrivs []crypto.PrivKey
	var sessionAddrs []crypto.Address
	var sessionAccNums []uint64

	for i := 0; i < 3; i++ {
		spriv, spub, saddr := tu.KeyTestPubAddr()
		createMsg := MsgCreateSession{
			Creator:    masterAddr,
			SessionKey: spub,
			ExpiresAt:  ctx.BlockTime().Unix() + 3600,
			SpendLimit: sessionSpendLimit(),
		}
		res := h.Process(ctx, createMsg)
		require.True(t, res.IsOK(), res.Log)

		sa := env.acck.GetSessionAccount(ctx, masterAddr, saddr)
		require.NotNil(t, sa)

		sessionPrivs = append(sessionPrivs, spriv)
		sessionAddrs = append(sessionAddrs, saddr)
		sessionAccNums = append(sessionAccNums, sa.GetAccountNumber())
	}

	// Verify all sessions work.
	fee := tu.NewTestFee()
	for i := 0; i < 3; i++ {
		msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
		tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPrivs[i], sessionAddrs[i], sessionAccNums[i], 0, fee)
		checkValidTx(t, anteHandler, ctx, tx, false)
	}

	// Revoke all sessions.
	revokeAllMsg := MsgRevokeAllSessions{Creator: masterAddr}
	res := h.Process(ctx, revokeAllMsg)
	require.True(t, res.IsOK(), res.Log)

	// Verify none of the sessions work anymore.
	for i := 0; i < 3; i++ {
		msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
		tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPrivs[i], sessionAddrs[i], sessionAccNums[i], 0, fee)
		checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})
	}
}

func TestSessionCreateValidation(t *testing.T) {
	t.Parallel()

	env, _, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	h := NewHandler(env.acck, env.gk)

	t.Run("expired time rejected", func(t *testing.T) {
		_, spub, _ := tu.KeyTestPubAddr()
		msg := MsgCreateSession{
			Creator:    masterAddr,
			SessionKey: spub,
			ExpiresAt:  ctx.BlockTime().Unix() - 10, // already expired
		}
		res := h.Process(ctx, msg)
		assert.False(t, res.IsOK(), "should reject already-expired session")
	})

	t.Run("too many sessions", func(t *testing.T) {
		// Create MaxSessionsPerAccount sessions.
		for i := 0; i < std.MaxSessionsPerAccount; i++ {
			_, spub, _ := tu.KeyTestPubAddr()
			msg := MsgCreateSession{
				Creator:    masterAddr,
				SessionKey: spub,
				ExpiresAt:  ctx.BlockTime().Unix() + 3600,
			}
			res := h.Process(ctx, msg)
			require.True(t, res.IsOK(), "session %d should succeed: %s", i, res.Log)
		}

		// Next one should fail.
		_, spub, _ := tu.KeyTestPubAddr()
		msg := MsgCreateSession{
			Creator:    masterAddr,
			SessionKey: spub,
			ExpiresAt:  ctx.BlockTime().Unix() + 3600,
		}
		res := h.Process(ctx, msg)
		assert.False(t, res.IsOK(), "should reject session when limit is reached")
	})

	t.Run("duplicate key rejected", func(t *testing.T) {
		// Use a fresh master for this subtest.
		_, masterPub2, masterAddr2 := tu.KeyTestPubAddr()
		acc2 := env.acck.NewAccountWithAddress(ctx, masterAddr2)
		acc2.SetCoins(tu.NewTestCoins())
		acc2.SetPubKey(masterPub2)
		env.acck.SetAccount(ctx, acc2)

		_, spub, _ := tu.KeyTestPubAddr()
		msg := MsgCreateSession{
			Creator:    masterAddr2,
			SessionKey: spub,
			ExpiresAt:  ctx.BlockTime().Unix() + 3600,
		}
		res := h.Process(ctx, msg)
		require.True(t, res.IsOK(), res.Log)

		// Try to create same session key again.
		res = h.Process(ctx, msg)
		assert.False(t, res.IsOK(), "should reject duplicate session key")
	})
}

func TestSessionReplayAfterRevoke(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	h := NewHandler(env.acck, env.gk)

	// Create session.
	createMsg := MsgCreateSession{
		Creator:    masterAddr,
		SessionKey: sessionPub,
		ExpiresAt:  ctx.BlockTime().Unix() + 3600,
		SpendLimit: sessionSpendLimit(),
	}
	res := h.Process(ctx, createMsg)
	require.True(t, res.IsOK(), res.Log)

	sa := env.acck.GetSessionAccount(ctx, masterAddr, sessionAddr)
	require.NotNil(t, sa)
	oldAccNum := sa.GetAccountNumber()

	// Use the session (seq=0).
	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, oldAccNum, 0, fee)
	newCtx, result, abort := anteHandler(ctx, tx, false)
	require.False(t, abort, result.Log)
	require.True(t, result.IsOK(), result.Log)
	_ = newCtx

	// Revoke session.
	revokeMsg := MsgRevokeSession{
		Creator:    masterAddr,
		SessionKey: sessionPub,
	}
	res = h.Process(ctx, revokeMsg)
	require.True(t, res.IsOK(), res.Log)

	// Recreate the same session key. It gets a new AccountNumber.
	res = h.Process(ctx, createMsg)
	require.True(t, res.IsOK(), res.Log)

	sa2 := env.acck.GetSessionAccount(ctx, masterAddr, sessionAddr)
	require.NotNil(t, sa2)
	newAccNum := sa2.GetAccountNumber()
	assert.NotEqual(t, oldAccNum, newAccNum, "recreated session should have a different account number")

	// Old signature (signed with oldAccNum, seq=0) should fail because
	// the account number has changed.
	oldTx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, oldAccNum, 0, fee)
	checkInvalidTx(t, anteHandler, ctx, oldTx, false, std.UnauthorizedError{})

	// New signature with newAccNum, seq=0 should work.
	newTx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, newAccNum, 0, fee)
	checkValidTx(t, anteHandler, ctx, newTx, false)
}

func TestSessionGasFeeDeductsFromMaster(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	// Record master's coins before.
	masterCoinsBefore := env.acck.GetAccount(ctx, masterAddr).GetCoins()

	// Create session.
	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createSessionDirect(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600)
	sessionAccNum := sa.GetAccountNumber()

	// Send session tx with a fee.
	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee() // 150 atom gas fee
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)
	newCtx, res, abort := anteHandler(ctx, tx, false)
	require.False(t, abort, res.Log)

	// Master's coins should be reduced by the gas fee.
	masterCoinsAfter := env.acck.GetAccount(newCtx, masterAddr).GetCoins()
	expected := masterCoinsBefore.SubUnsafe(std.Coins{fee.GasFee})
	assert.True(t, expected.IsEqual(masterCoinsAfter),
		"master coins: before=%s after=%s expected=%s", masterCoinsBefore, masterCoinsAfter, expected)

	// Session account should have NO coins (fees come from master).
	updatedSA := env.acck.GetSessionAccount(newCtx, masterAddr, sessionAddr)
	assert.True(t, updatedSA.GetCoins().IsZero(), "session account should have no coins")
}

func TestSessionGasExceedsSpendLimit(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	// Create session with a tiny spend limit (1 atom).
	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := env.acck.NewSessionAccount(env.ctx, masterAddr, sessionPub)
	da := sa.(std.DelegatedAccount)
	da.SetExpiresAt(ctx.BlockTime().Unix() + 3600)
	da.SetSpendLimit(std.Coins{std.NewCoin("atom", 1)}) // only 1 atom allowed
	da.SetSpendReset(ctx.BlockTime().Unix())
	env.acck.SetSessionAccount(env.ctx, masterAddr, sa)

	sessionAccNum := sa.GetAccountNumber()

	// Fee is 150 atom — exceeds the 1 atom spend limit.
	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	// Should be rejected because gas fee exceeds spend limit.
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.SessionNotAllowedError{})
}

func TestSessionNoSpendLimitRejectsGas(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	// Create session with NO spend limit.
	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := env.acck.NewSessionAccount(env.ctx, masterAddr, sessionPub)
	da := sa.(std.DelegatedAccount)
	da.SetExpiresAt(ctx.BlockTime().Unix() + 3600)
	// No SetSpendLimit — defaults to nil/empty.
	da.SetSpendReset(ctx.BlockTime().Unix())
	env.acck.SetSessionAccount(env.ctx, masterAddr, sa)

	sessionAccNum := sa.GetAccountNumber()

	// Any non-zero fee should be rejected since there's no spend limit.
	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	checkInvalidTx(t, anteHandler, ctx, tx, false, std.SessionNotAllowedError{})
}

func TestNonSessionTxStillChargesFees(t *testing.T) {
	t.Parallel()

	env, anteHandler, masterPriv, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	// Record master's coins before.
	masterCoinsBefore := env.acck.GetAccount(ctx, masterAddr).GetCoins()

	// Regular (non-session) tx.
	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	accNum := env.acck.GetAccount(ctx, masterAddr).GetAccountNumber()
	seq := env.acck.GetAccount(ctx, masterAddr).GetSequence()

	tx := tu.NewTestTx(t, ctx.ChainID(), msgs, []crypto.PrivKey{masterPriv}, []uint64{accNum}, []uint64{seq}, fee)
	newCtx, res, abort := anteHandler(ctx, tx, false)
	require.False(t, abort, res.Log)

	// Master's coins should be reduced by the gas fee.
	masterCoinsAfter := env.acck.GetAccount(newCtx, masterAddr).GetCoins()
	expected := masterCoinsBefore.SubUnsafe(std.Coins{fee.GasFee})
	assert.True(t, expected.IsEqual(masterCoinsAfter),
		"non-session tx: master coins before=%s after=%s expected=%s", masterCoinsBefore, masterCoinsAfter, expected)
}

func TestSessionMasterInsufficientFunds(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	// Drain master's coins so it can't pay fees.
	masterAcc := env.acck.GetAccount(ctx, masterAddr)
	masterAcc.SetCoins(std.Coins{std.NewCoin("atom", 1)}) // only 1 atom
	env.acck.SetAccount(ctx, masterAcc)

	// Create session with a generous spend limit.
	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createSessionDirect(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600)
	sessionAccNum := sa.GetAccountNumber()

	// Fee is 150 atom — session spend limit allows it, but master can't pay.
	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	// Should fail with InsufficientFundsError, not SessionNotAllowedError.
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.InsufficientFundsError{})
}

// --- Pentester-style attack tests ---

func TestSessionCrossAccountAttack(t *testing.T) {
	// Attack: masterA's session tries to sign a tx where msg.Caller = masterB.
	// The session is keyed under masterA, but the msg claims to be from masterB.
	t.Parallel()

	env, anteHandler, _, _, masterAddrA := setupSessionEnv(t)
	ctx := env.ctx

	// Create a second master account (masterB).
	_, masterPubB, masterAddrB := tu.KeyTestPubAddr()
	masterAccB := env.acck.NewAccountWithAddress(ctx, masterAddrB)
	masterAccB.SetCoins(tu.NewTestCoins())
	masterAccB.SetPubKey(masterPubB)
	env.acck.SetAccount(ctx, masterAccB)

	// Create session under masterA.
	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createSessionDirect(t, env, masterAddrA, sessionPub, ctx.BlockTime().Unix()+3600)
	sessionAccNum := sa.GetAccountNumber()

	// Msg claims to be from masterB, but SessionAddr is masterA's session.
	// tx.GetSigners() returns [masterB], so AnteHandler loads masterB's account.
	// Then looks for session at /a/<masterB>/s/<sessionAddr> — doesn't exist.
	msgs := []std.Msg{tu.NewTestMsg(masterAddrB)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})
}

func TestSessionSelfReferentialMaster(t *testing.T) {
	// Attack: create a session where MasterAddress = session's own address.
	// This shouldn't work because the session key's address won't have
	// a regular account, so GetSignerAcc would fail.
	t.Parallel()

	env, anteHandler, _, _, _ := setupSessionEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()

	// Manually create a self-referential session (master = session addr).
	sa := env.acck.NewSessionAccount(ctx, sessionAddr, sessionPub)
	da := sa.(std.DelegatedAccount)
	da.SetExpiresAt(ctx.BlockTime().Unix() + 3600)
	da.SetSpendLimit(sessionSpendLimit())
	da.SetSpendReset(ctx.BlockTime().Unix())
	env.acck.SetSessionAccount(ctx, sessionAddr, sa)

	// Try to sign a tx. GetSigners() returns [sessionAddr].
	// AnteHandler tries GetAccount(sessionAddr) — no regular account exists.
	msgs := []std.Msg{tu.NewTestMsg(sessionAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sa.GetAccountNumber(), 0, fee)

	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnknownAddressError{})
}

func TestSessionExpiryBoundary(t *testing.T) {
	// ExpiresAt == blockTime. The check is `>=`, so this should be expired.
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx
	blockTime := ctx.BlockTime().Unix()

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createSessionDirect(t, env, masterAddr, sessionPub, blockTime) // expires exactly at block time
	// Manually override the expiry since createSessionDirect won't create an expired one.
	da := sa.(std.DelegatedAccount)
	da.SetExpiresAt(blockTime)
	env.acck.SetSessionAccount(ctx, masterAddr, sa)

	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sa.GetAccountNumber(), 0, fee)

	checkInvalidTx(t, anteHandler, ctx, tx, false, std.SessionExpiredError{})
}

func TestSessionZeroExpiry(t *testing.T) {
	// ExpiresAt = 0 means "no expiry" — the session is valid until revoked.
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := env.acck.NewSessionAccount(ctx, masterAddr, sessionPub)
	da := sa.(std.DelegatedAccount)
	da.SetExpiresAt(0) // zero means no expiry
	da.SetSpendLimit(sessionSpendLimit())
	da.SetSpendReset(ctx.BlockTime().Unix())
	env.acck.SetSessionAccount(ctx, masterAddr, sa)

	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sa.GetAccountNumber(), 0, fee)

	// ExpiresAt=0 means no expiry, so this should succeed.
	checkValidTx(t, anteHandler, ctx, tx, false)
}

func TestSessionNoExpiry(t *testing.T) {
	// ExpiresAt = 0 means "no expiry" — verify it works even at a much later block time.
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := env.acck.NewSessionAccount(ctx, masterAddr, sessionPub)
	da := sa.(std.DelegatedAccount)
	da.SetExpiresAt(0) // no expiry
	da.SetSpendLimit(sessionSpendLimit())
	da.SetSpendReset(ctx.BlockTime().Unix())
	env.acck.SetSessionAccount(ctx, masterAddr, sa)

	// Advance block time far into the future (1 year).
	futureCtx := ctx.WithBlockHeader(&bft.Header{
		ChainID: ctx.ChainID(),
		Height:  2,
		Time:    ctx.BlockTime().Add(365 * 24 * time.Hour),
	})
	futureCtx = futureCtx.WithValue(AuthParamsContextKey{}, env.acck.GetParams(futureCtx))

	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, futureCtx.ChainID(), msgs, sessionPriv, sessionAddr, sa.GetAccountNumber(), 0, fee)

	// Should still work — no expiry means valid forever (until revoked).
	checkValidTx(t, anteHandler, futureCtx, tx, false)
}

func TestSessionWrongSignature(t *testing.T) {
	// Attack: valid SessionAddr, but signed by a completely different key.
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	_, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createSessionDirect(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600)

	// Sign with a different key entirely.
	attackerPriv, _, _ := tu.KeyTestPubAddr()

	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, attackerPriv, sessionAddr, sa.GetAccountNumber(), 0, fee)

	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})
}

func TestSessionFutureSequence(t *testing.T) {
	// Attack: use sequence=999 to skip ahead, potentially replaying after
	// the real sequence catches up.
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createSessionDirect(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600)

	// Sign with sequence=999 (actual is 0).
	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sa.GetAccountNumber(), 999, fee)

	// Signature is over (accNum, seq=999), but the stored sequence is 0.
	// Sign bytes won't match — should fail.
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})
}

func TestSessionCannotCreateSubSession(t *testing.T) {
	// Attack: session key signs MsgCreateSession to create a sub-session.
	// Sessions should not be able to create other sessions.
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createSessionDirect(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600)

	// Session tries to create another session.
	_, subPub, _ := tu.KeyTestPubAddr()
	createMsg := MsgCreateSession{
		Creator:    masterAddr,
		SessionKey: subPub,
		ExpiresAt:  ctx.BlockTime().Unix() + 3600,
		SpendLimit: sessionSpendLimit(),
	}

	// This tx is signed by the session key, so it should pass AnteHandler
	// (auth-level check). But MsgCreateSession.GetSigners() returns [masterAddr],
	// and the session is for masterAddr, so auth passes.
	// However, the handler should process the MsgCreateSession and it would work...
	// The protection is at the gno.land layer: MsgCreateSession.Type() = "create_session" != "exec",
	// so the gno.land ante wrapper would reject it.
	// At the tm2 level, there's no msg type restriction.
	// This test verifies the AnteHandler itself doesn't block it (that's gno.land's job).
	msgs := []std.Msg{createMsg}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sa.GetAccountNumber(), 0, fee)

	// tm2 AnteHandler passes — it doesn't check msg types.
	_, res, abort := anteHandler(ctx, tx, false)
	assert.False(t, abort, "tm2 AnteHandler should not reject based on msg type: %s", res.Log)
}

func TestSessionSpendPeriodReset(t *testing.T) {
	// Verify that spend tracking resets after the period expires.
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := env.acck.NewSessionAccount(ctx, masterAddr, sessionPub)
	da := sa.(std.DelegatedAccount)
	da.SetExpiresAt(ctx.BlockTime().Unix() + 7200) // 2 hours
	da.SetSpendLimit(std.Coins{std.NewCoin("atom", 200)})
	da.SetSpendPeriod(3600) // 1 hour period
	da.SetSpendReset(ctx.BlockTime().Unix())
	env.acck.SetSessionAccount(ctx, masterAddr, sa)

	// First tx: fee=150 atom, within 200 limit.
	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee() // 150 atom
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sa.GetAccountNumber(), 0, fee)

	newCtx, res, abort := anteHandler(ctx, tx, false)
	require.False(t, abort, res.Log)

	// Second tx at same time: another 150 would exceed 200 limit.
	tx2 := tu.NewSessionTestTx(t, newCtx.ChainID(), msgs, sessionPriv, sessionAddr, sa.GetAccountNumber(), 1, fee)
	_, res2, abort2 := anteHandler(newCtx, tx2, false)
	require.True(t, abort2, "should exceed spend limit")
	_ = res2

	// Advance time past the spend period (1 hour + 1 second).
	futureCtx := newCtx.WithBlockHeader(&bft.Header{
		ChainID: ctx.ChainID(),
		Height:  2,
		Time:    ctx.BlockTime().Add(3601 * time.Second),
	})
	futureCtx = futureCtx.WithValue(AuthParamsContextKey{}, env.acck.GetParams(futureCtx))

	// Now 150 should work again — period has reset.
	tx3 := tu.NewSessionTestTx(t, futureCtx.ChainID(), msgs, sessionPriv, sessionAddr, sa.GetAccountNumber(), 1, fee)
	_, res3, abort3 := anteHandler(futureCtx, tx3, false)
	require.False(t, abort3, "should pass after spend period reset: %s", res3.Log)
}

func TestSessionMultipleMsgsSameSession(t *testing.T) {
	// A tx with multiple msgs all signed by the same session key.
	// The session should only count once in the signers list.
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createSessionDirect(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600)

	// Two msgs, both from masterAddr (same signer).
	msgs := []std.Msg{
		tu.NewTestMsg(masterAddr),
		tu.NewTestMsg(masterAddr),
	}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sa.GetAccountNumber(), 0, fee)

	// Should pass — one signer, one signature, one session.
	_, res, abort := anteHandler(ctx, tx, false)
	require.False(t, abort, res.Log)
	assert.True(t, res.IsOK(), res.Log)
}

func TestSessionPubKeyMismatchAttack(t *testing.T) {
	// Attack: provide a sig.PubKey that differs from the stored session pubkey.
	// This tests the pubkey mismatch check.
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	_, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createSessionDirect(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600)

	// Sign with a different key but manually set SessionAddr.
	attackerPriv, _, _ := tu.KeyTestPubAddr()

	// Build tx manually with attacker's pubkey AND sessionAddr set.
	signBytes, err := std.GetSignaturePayload(std.SignDoc{
		ChainID:       ctx.ChainID(),
		AccountNumber: sa.GetAccountNumber(),
		Sequence:      0,
		Fee:           tu.NewTestFee(),
		Msgs:          []std.Msg{tu.NewTestMsg(masterAddr)},
	})
	require.NoError(t, err)

	sig, err := attackerPriv.Sign(signBytes)
	require.NoError(t, err)

	tx := std.NewTx(
		[]std.Msg{tu.NewTestMsg(masterAddr)},
		tu.NewTestFee(),
		[]std.Signature{{
			PubKey:      attackerPriv.PubKey(), // wrong pubkey
			SessionAddr: sessionAddr,
			Signature:   sig,
		}},
		"",
	)

	// Should fail: sig.PubKey doesn't match stored session pubkey.
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})
}

func TestSessionChainIDReplay(t *testing.T) {
	// Attack: replay a session tx from one chain on another chain.
	// Sign bytes include chainID, so this should fail.
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createSessionDirect(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600)

	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()

	// Sign for a different chain ID.
	tx := tu.NewSessionTestTx(t, "other-chain-id", msgs, sessionPriv, sessionAddr, sa.GetAccountNumber(), 0, fee)

	// Should fail — chainID mismatch in sign bytes.
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})
}
