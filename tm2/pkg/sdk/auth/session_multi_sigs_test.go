package auth

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TestTwoSessionSignaturesTwoMasters creates a trasaction with two session account
// signatures and two messages with different master accounts. Validation should pass.
func TestTwoSessionSignaturesTwoMasters(t *testing.T) {
	t.Parallel()

	// Create master accounts.
	env, anteHandler, _, masterAddr1 := setupSessionEnv(t)
	ctx := env.ctx
	_, masterAddr2 := setupSessionFromEnv(t, env)

	// Create session key pairs.
	sessionPriv1, sessionPub1, sessionAddr1 := tu.KeyTestPubAddr()
	sessionPriv2, sessionPub2, sessionAddr2 := tu.KeyTestPubAddr()

	// Create session accounts (with separate master accounts) with 1-hour expiry.
	sa1 := createSessionDirect(t, env, masterAddr1, sessionPub1, ctx.BlockTime().Unix()+3600)
	sessionAccNum1 := sa1.GetAccountNumber()
	sessionSeq1 := sa1.GetSequence()
	sa2 := createSessionDirect(t, env, masterAddr2, sessionPub2, ctx.BlockTime().Unix()+3600)
	sessionAccNum2 := sa2.GetAccountNumber()
	sessionSeq2 := sa2.GetSequence()

	// Build a tx with two messages, initially signed by session key 1.
	msgs := []std.Msg{tu.NewTestMsg(masterAddr1), tu.NewTestMsg(masterAddr2)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv1, sessionAddr1, sessionAccNum1, sessionSeq1, fee)

	// Add a second signature by session key 2.
	sig2 := tu.NewSessionTestSignature(t, ctx.ChainID(), msgs, sessionPriv2, sessionAddr2, sessionAccNum2, sessionSeq2, fee)
	tx.Signatures = append(tx.Signatures, sig2)

	checkValidTx(t, anteHandler, ctx, tx, false)
}

// TestTwoSessionSignaturesOneMaster creates a transaction with two session account
// signatures and two messages with the same master. Validation should fail because GetSigners()
// deduplicates and returns only one signer address, a mismatch with two signatures.
// (However, this would succeed if we deduplicate the combined signer-addr/signature-pubkey.)
func TestTwoSessionSignaturesOneMaster(t *testing.T) {
	t.Parallel()

	// Create the master account.
	env, anteHandler, _, masterAddr := setupSessionEnv(t)
	ctx := env.ctx

	// Create session key pairs.
	sessionPriv1, sessionPub1, sessionAddr1 := tu.KeyTestPubAddr()
	sessionPriv2, sessionPub2, sessionAddr2 := tu.KeyTestPubAddr()

	// Create session accounts (same master) with 1-hour expiry.
	sa1 := createSessionDirect(t, env, masterAddr, sessionPub1, ctx.BlockTime().Unix()+3600)
	sessionAccNum1 := sa1.GetAccountNumber()
	sessionSeq1 := sa1.GetSequence()
	sa2 := createSessionDirect(t, env, masterAddr, sessionPub2, ctx.BlockTime().Unix()+3600)
	sessionAccNum2 := sa2.GetAccountNumber()
	sessionSeq2 := sa2.GetSequence()

	// Build a tx with two messages (same master "signer"), initially signed by session key 1.
	msgs := []std.Msg{tu.NewTestMsg(masterAddr), tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv1, sessionAddr1, sessionAccNum1, sessionSeq1, fee)

	// Add a second signature by session key 2.
	sig2 := tu.NewSessionTestSignature(t, ctx.ChainID(), msgs, sessionPriv2, sessionAddr2, sessionAccNum2, sessionSeq2, fee)
	tx.Signatures = append(tx.Signatures, sig2)

	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})
}

// TestOneMasterOneSessionSignature creates a transaction with a signature by the master
// account plus a signature by a session account under that master. It has
// two messages with the same signer address of the master. Validation should fail because GetSigners()
// deduplicates and returns only one signer address, a mismatch with two signatures.
// (However, this would succeed if we deduplicate the combined signer-addr/signature-pubkey.)
func TestOneMasterOneSessionSignature(t *testing.T) {
	t.Parallel()

	// Create master account.
	env, anteHandler, masterPriv, masterAddr := setupSessionEnv(t)
	ctx := env.ctx
	masterAcct := env.acck.GetAccount(ctx, masterAddr)
	masterAccNum := masterAcct.GetAccountNumber()
	masterSeq := masterAcct.GetSequence()

	// Create session accounts (same master) with 1-hour expiry.
	sessionPriv, sessionPub, sessionAddr := tu.KeyTestPubAddr()
	sa := createSessionDirect(t, env, masterAddr, sessionPub, ctx.BlockTime().Unix()+3600)
	sessionAccNum := sa.GetAccountNumber()
	sessionSeq := sa.GetSequence()

	// Build a tx with two messages (same master "signer"), initially signed by the master.
	msgs := []std.Msg{tu.NewTestMsg(masterAddr), tu.NewTestMsg(masterAddr)}
	fee := tu.NewTestFee()
	tx := tu.NewTestTx(t, ctx.ChainID(), msgs, []crypto.PrivKey{masterPriv}, []uint64{masterAccNum}, []uint64{masterSeq}, fee)

	// Add a second signature by the session account.
	sig2 := tu.NewSessionTestSignature(t, ctx.ChainID(), msgs, sessionPriv, sessionAddr, sessionAccNum, sessionSeq, fee)
	tx.Signatures = append(tx.Signatures, sig2)

	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})
}

// TestThreeSessionSignaturesTwoMasters creates a transaction with three session account
// signatures and three messages with two different masters. Validation should fail because GetSigners()
// deduplicates and returns only two signer addresses, a mismatch with three signatures.
// (We could try to deduplicate the combined signer-addr/signature-pubkey but it is not clear which
// master belongs to which account session signature.)
func TestThreeSessionSignaturesTwoMasters(t *testing.T) {
	t.Parallel()

	// Create master accounts.
	env, anteHandler, _, masterAddr1 := setupSessionEnv(t)
	ctx := env.ctx
	_, masterAddr2 := setupSessionFromEnv(t, env)

	// Create session key pairs.
	sessionPriv1, sessionPub1, sessionAddr1 := tu.KeyTestPubAddr()
	sessionPriv2, sessionPub2, sessionAddr2 := tu.KeyTestPubAddr()
	sessionPriv3, sessionPub3, sessionAddr3 := tu.KeyTestPubAddr()

	// Create session accounts (one with masterAddr1, two with masterAddr2) with 1-hour expiry.
	sa1 := createSessionDirect(t, env, masterAddr1, sessionPub1, ctx.BlockTime().Unix()+3600)
	sessionAccNum1 := sa1.GetAccountNumber()
	sessionSeq1 := sa1.GetSequence()
	sa2 := createSessionDirect(t, env, masterAddr2, sessionPub2, ctx.BlockTime().Unix()+3600)
	sessionAccNum2 := sa2.GetAccountNumber()
	sessionSeq2 := sa2.GetSequence()
	sa3 := createSessionDirect(t, env, masterAddr2, sessionPub3, ctx.BlockTime().Unix()+3600)
	sessionAccNum3 := sa3.GetAccountNumber()
	sessionSeq3 := sa3.GetSequence()

	// Build a tx with three messages, initially signed by session key 1.
	msgs := []std.Msg{tu.NewTestMsg(masterAddr1), tu.NewTestMsg(masterAddr2), tu.NewTestMsg(masterAddr2)}
	fee := tu.NewTestFee()
	tx := tu.NewSessionTestTx(t, ctx.ChainID(), msgs, sessionPriv1, sessionAddr1, sessionAccNum1, sessionSeq1, fee)

	// Add a second signature by session key 2.
	sig2 := tu.NewSessionTestSignature(t, ctx.ChainID(), msgs, sessionPriv2, sessionAddr2, sessionAccNum2, sessionSeq2, fee)
	tx.Signatures = append(tx.Signatures, sig2)
	// Add a third signature by session key 3.
	sig3 := tu.NewSessionTestSignature(t, ctx.ChainID(), msgs, sessionPriv3, sessionAddr3, sessionAccNum3, sessionSeq3, fee)
	tx.Signatures = append(tx.Signatures, sig3)

	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})
}
