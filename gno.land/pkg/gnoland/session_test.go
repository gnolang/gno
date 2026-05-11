package gnoland

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func setupSessionGnoEnv(t *testing.T) (testEnv, sdk.AnteHandler, crypto.PrivKey, crypto.PubKey, crypto.Address) {
	t.Helper()

	env := setupTestEnv()

	// Set auth params including fee_collector.
	params := auth.DefaultParams()
	env.acck.SetParams(env.ctx, params)

	// Build the full ante handler chain: tm2 auth + gno.land wrapper.
	authAnteHandler := auth.NewAnteHandler(
		env.acck, env.bankk, auth.DefaultSigVerificationGasConsumer,
		auth.AnteOptions{VerifyGenesisSignatures: false})

	// Wrap with gno.land session restrictions using the same function as app.go.
	anteHandler := func(ctx sdk.Context, tx std.Tx, simulate bool) (
		newCtx sdk.Context, res sdk.Result, abort bool,
	) {
		ctx = ctx.WithValue(auth.AuthParamsContextKey{}, env.acck.GetParams(ctx))
		newCtx, res, abort = authAnteHandler(ctx, tx, simulate)
		if abort {
			return
		}
		if res, abort = checkSessionRestrictions(newCtx, tx); abort {
			return newCtx, res, true
		}
		return
	}

	// Create and fund master account.
	masterPriv, masterPub, masterAddr := tu.KeyTestPubAddr()
	masterAcc := env.acck.NewAccountWithAddress(env.ctx, masterAddr)
	masterAcc.SetCoins(tu.NewTestCoins())
	masterAcc.SetPubKey(masterPub)
	env.acck.SetAccount(env.ctx, masterAcc)

	// Set block time.
	now := time.Now()
	env.ctx = env.ctx.WithBlockHeader(&bft.Header{
		ChainID: env.ctx.ChainID(),
		Height:  1,
		Time:    now,
	})

	return env, anteHandler, masterPriv, masterPub, masterAddr
}

func createGnoSession(t *testing.T, env testEnv, masterAddr crypto.Address, sessionPub crypto.PubKey, expiresAt int64, allowPaths []string) std.Account {
	t.Helper()

	sa := env.acck.NewSessionAccount(env.ctx, masterAddr, sessionPub)
	da := sa.(std.DelegatedAccount)
	da.SetExpiresAt(expiresAt)
	da.SetSpendLimit(std.Coins{std.NewCoin("atom", 10000000)})
	da.SetSpendReset(env.ctx.BlockTime().Unix())

	if len(allowPaths) > 0 {
		sa.(*GnoSessionAccount).SetAllowPaths(allowPaths)
	}

	env.acck.SetSessionAccount(env.ctx, masterAddr, sa)
	return sa
}

func TestSessionAllowPathsExactMatch(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionGnoEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createGnoSession(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600,
		[]string{"gno.land/r/demo/boards"})
	sessionAccNum := sa.GetAccountNumber()

	// MsgCall to the exact allowed path — should pass.
	msgs := []std.Msg{tu.MockMsgCall{Caller: masterAddr, PkgPath: "gno.land/r/demo/boards"}}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	_, res, abort := anteHandler(ctx, tx, false)
	require.False(t, abort, res.Log)
	assert.True(t, res.IsOK(), res.Log)
}

func TestSessionAllowPathsSubPath(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionGnoEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createGnoSession(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600,
		[]string{"gno.land/r/demo"})
	sessionAccNum := sa.GetAccountNumber()

	// MsgCall to a sub-path — should pass.
	msgs := []std.Msg{tu.MockMsgCall{Caller: masterAddr, PkgPath: "gno.land/r/demo/boards"}}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	_, res, abort := anteHandler(ctx, tx, false)
	require.False(t, abort, res.Log)
	assert.True(t, res.IsOK(), res.Log)
}

func TestSessionAllowPathsDenied(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionGnoEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createGnoSession(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600,
		[]string{"gno.land/r/demo/boards"})
	sessionAccNum := sa.GetAccountNumber()

	// MsgCall to a different path — should be denied.
	msgs := []std.Msg{tu.MockMsgCall{Caller: masterAddr, PkgPath: "gno.land/r/demo/chat"}}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	_, res, abort := anteHandler(ctx, tx, false)
	require.True(t, abort, "should reject disallowed path")
	assert.Contains(t, res.Log, "AllowPaths")
}

func TestSessionAllowPathsPrefixAttack(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionGnoEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createGnoSession(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600,
		[]string{"gno.land/r/demo"})
	sessionAccNum := sa.GetAccountNumber()

	// "gno.land/r/demo_evil" shares the prefix "gno.land/r/demo" but is NOT
	// a sub-path — should be denied.
	msgs := []std.Msg{tu.MockMsgCall{Caller: masterAddr, PkgPath: "gno.land/r/demo_evil"}}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	_, res, abort := anteHandler(ctx, tx, false)
	require.True(t, abort, "should reject prefix attack path")
	assert.Contains(t, res.Log, "AllowPaths")
}

func TestSessionAllowPathsUnrestricted(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionGnoEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	// Empty AllowPaths = unrestricted for MsgCall.
	sa := createGnoSession(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600, nil)
	sessionAccNum := sa.GetAccountNumber()

	// MsgCall to any path should pass.
	msgs := []std.Msg{tu.MockMsgCall{Caller: masterAddr, PkgPath: "gno.land/r/anything/here"}}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	_, res, abort := anteHandler(ctx, tx, false)
	require.False(t, abort, res.Log)
	assert.True(t, res.IsOK(), res.Log)
}

func TestSessionAllowPathsMultipleEntries(t *testing.T) {
	// Subtests share state and must run in sequence order (seq 0, 1, 2),
	// so the outer test is intentionally non-parallel.
	env, anteHandler, _, _, masterAddr := setupSessionGnoEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createGnoSession(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600,
		[]string{"gno.land/r/demo/boards", "gno.land/r/demo/chat"})
	sessionAccNum := sa.GetAccountNumber()
	fee := tu.NewTestFee()

	t.Run("first allowed path succeeds", func(t *testing.T) {
		msgs := []std.Msg{tu.MockMsgCall{Caller: masterAddr, PkgPath: "gno.land/r/demo/boards"}}
		tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)
		_, res, abort := anteHandler(ctx, tx, false)
		require.False(t, abort, res.Log)
		assert.True(t, res.IsOK(), res.Log)
	})

	t.Run("second allowed path succeeds", func(t *testing.T) {
		msgs := []std.Msg{tu.MockMsgCall{Caller: masterAddr, PkgPath: "gno.land/r/demo/chat"}}
		tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 1, fee)
		_, res, abort := anteHandler(ctx, tx, false)
		require.False(t, abort, res.Log)
		assert.True(t, res.IsOK(), res.Log)
	})

	t.Run("disallowed path fails", func(t *testing.T) {
		msgs := []std.Msg{tu.MockMsgCall{Caller: masterAddr, PkgPath: "gno.land/r/demo/users"}}
		tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 2, fee)
		_, res, abort := anteHandler(ctx, tx, false)
		require.True(t, abort, "should reject disallowed path")
		assert.Contains(t, res.Log, "AllowPaths")
	})
}

func TestSessionAllowsMsgRun(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionGnoEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	// Empty AllowPaths — unrestricted session. MsgRun is in the allowlist.
	sa := createGnoSession(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600, nil)
	sessionAccNum := sa.GetAccountNumber()

	msgs := []std.Msg{tu.MockMsgRun{Caller: masterAddr}}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	_, res, abort := anteHandler(ctx, tx, false)
	require.False(t, abort, res.Log)
	assert.True(t, res.IsOK(), res.Log)
}

func TestSessionDeniesMsgRunWithAllowPaths(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionGnoEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	// Non-empty AllowPaths makes the session realm-scoped. MsgRun has no
	// pkgPather, so the AllowPaths check rejects it — intentional, since
	// MsgRun can execute arbitrary code and would escape path scope.
	sa := createGnoSession(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600,
		[]string{"gno.land/r/demo/boards"})
	sessionAccNum := sa.GetAccountNumber()

	msgs := []std.Msg{tu.MockMsgRun{Caller: masterAddr}}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	_, res, abort := anteHandler(ctx, tx, false)
	require.True(t, abort, "session with AllowPaths set should reject MsgRun")
	assert.Contains(t, res.Log, "AllowPaths")
}

// TestSessionAllowsMsgSend confirms bank.MsgSend passes the session
// allowlist. Spend-limit enforcement happens inside bank.Keeper.SendCoins
// (tm2/pkg/sdk/bank/keeper.go) when the msg is actually handled; the
// gno.land-layer ante wrapper only decides msg-type admissibility. See
// tm2/pkg/sdk/auth/session_test.go for end-to-end spend-limit tests.
func TestSessionAllowsMsgSend(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionGnoEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createGnoSession(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600, nil)
	sessionAccNum := sa.GetAccountNumber()

	_, _, recipient := tu.KeyTestPubAddr()
	msgs := []std.Msg{tu.MockMsgSend{From: masterAddr, To: recipient, Amount: std.Coins{std.NewCoin("atom", 100)}}}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	_, res, abort := anteHandler(ctx, tx, false)
	require.False(t, abort, res.Log)
	assert.True(t, res.IsOK(), res.Log)
}

// TestSessionAllowsMsgMultiSend mirrors TestSessionAllowsMsgSend for the
// multisend path. Per-input spend enforcement is in
// bank.Keeper.InputOutputCoins.
func TestSessionAllowsMsgMultiSend(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionGnoEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createGnoSession(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600, nil)
	sessionAccNum := sa.GetAccountNumber()

	// MockMsgSend is in the "send" type which is allowed; a MockMsgMultiSend
	// would exercise the "multisend" allowlist entry. We reuse MockMsgSend
	// here to confirm the gate passes; MsgMultiSend end-to-end with a real
	// keeper is tested in tm2/pkg/sdk/auth/session_test.go.
	_, _, recipient := tu.KeyTestPubAddr()
	msgs := []std.Msg{tu.MockMsgSend{From: masterAddr, To: recipient, Amount: std.Coins{std.NewCoin("atom", 10)}}}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	_, res, abort := anteHandler(ctx, tx, false)
	require.False(t, abort, res.Log)
	assert.True(t, res.IsOK(), res.Log)
}

// TestSessionDeniesMsgSendWithAllowPaths confirms that a session with
// non-empty AllowPaths rejects bank.MsgSend: a realm-scoped session
// should not escape its scope via path-less coin transfers.
func TestSessionDeniesMsgSendWithAllowPaths(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionGnoEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createGnoSession(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600,
		[]string{"gno.land/r/demo/boards"})
	sessionAccNum := sa.GetAccountNumber()

	_, _, recipient := tu.KeyTestPubAddr()
	msgs := []std.Msg{tu.MockMsgSend{From: masterAddr, To: recipient, Amount: std.Coins{std.NewCoin("atom", 100)}}}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	_, res, abort := anteHandler(ctx, tx, false)
	require.True(t, abort, "session with AllowPaths set should reject MsgSend")
	assert.Contains(t, res.Log, "AllowPaths")
}

// TestSessionCreateRejectsFromSession confirms that a session-signed
// tx carrying a MsgCreateSession is rejected at the gno.land allowlist.
// Sessions must not be able to create other sessions — that would be
// privilege escalation equivalent to the master key.
func TestSessionCreateRejectsFromSession(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionGnoEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createGnoSession(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600, nil)
	sessionAccNum := sa.GetAccountNumber()

	// Build MsgCreateSession signed by the session key (via the session's
	// SessionTestTx helper). Creator is the master address (same as any
	// MsgCreateSession) but the tx is signed by the session, not the
	// master. The allowlist must reject it.
	_, subPub, _ := tu.KeyTestPubAddr()
	createMsg := auth.MsgCreateSession{
		Creator:    masterAddr,
		SessionKey: subPub,
		ExpiresAt:  ctx.BlockTime().Unix() + 600,
		SpendLimit: std.Coins{std.NewCoin("atom", 10)},
	}

	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), []std.Msg{createMsg}, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	_, res, abort := anteHandler(ctx, tx, false)
	require.True(t, abort, "session-signed MsgCreateSession must be rejected")
	assert.Contains(t, res.Log, "not allowed for session")
}

func TestSessionDeniesDisallowedMsg(t *testing.T) {
	t.Parallel()

	env, anteHandler, _, _, masterAddr := setupSessionGnoEnv(t)
	ctx := env.ctx

	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	// Even with empty AllowPaths (unrestricted), msg types outside the
	// session allowlist (exec, run, send, multisend) are denied.
	sa := createGnoSession(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600, nil)
	sessionAccNum := sa.GetAccountNumber()

	// TestMsg has Type() = "Test message", not in the allowlist — should be denied.
	msgs := []std.Msg{tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, 0, fee)

	_, res, abort := anteHandler(ctx, tx, false)
	require.True(t, abort, "should reject disallowed msg type for session")
	assert.Contains(t, res.Log, "not allowed for session")
}
