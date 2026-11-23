package auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TestAnteHandlerRejectsSmallOrderPubKey covers the account path. An address
// is the hash of its pubkey, and the first transaction fixes the pubkey, so a
// caller can present a small-order key for an address that already holds
// value. Before the VerifyBytes gate, the forged (R = identity, S = 0)
// signature verified and the transaction was admitted, making the address
// spendable by anyone.
//
// The fee is deducted in phase 2, before signature verification in phase 3;
// in a real run the whole transaction sits in a cache store that is discarded
// on abort, so only the account's own state is asserted here.
func TestAnteHandlerRejectsSmallOrderPubKey(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx
	anteHandler := NewAnteHandler(env.acck, env.bankk, DefaultSigVerificationGasConsumer, defaultAnteOptions())

	var pk ed25519.PubKeyEd25519
	pk[0] = 1 // identity point
	addr := pk.Address()

	acc := env.acck.NewAccountWithAddress(ctx, addr)
	require.NoError(t, acc.SetCoins(tu.NewTestCoins()))
	env.acck.SetAccount(ctx, acc)

	fee := tu.NewTestFee()
	msg := tu.NewTestMsg(addr)
	signBytes, err := std.GetSignaturePayload(std.SignDoc{
		ChainID:       ctx.ChainID(),
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      acc.GetSequence(),
		Fee:           fee,
		Msgs:          []std.Msg{msg},
	})
	require.NoError(t, err)

	forged := make([]byte, ed25519.SignatureSize)
	forged[0] = 1 // R = identity, S = 0
	require.False(t, pk.VerifyBytes(signBytes, forged))

	tx := std.NewTx([]std.Msg{msg}, fee,
		[]std.Signature{{PubKey: pk, Signature: forged}}, "")

	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})

	// The account did not advance: no pubkey stored, sequence unchanged.
	got := env.acck.GetAccount(ctx, addr)
	require.NotNil(t, got)
	require.Nil(t, got.GetPubKey())
	require.Equal(t, acc.GetSequence(), got.GetSequence())
}
