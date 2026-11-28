package auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestAnteHandlerMultisigSameSignerAcrossTypes(t *testing.T) {
	t.Parallel()

	priv := ed25519.GenPrivKey()
	leaf := priv.PubKey()
	wrapped := multisig.PubKeyMultisigThreshold{K: 1, PubKeys: []crypto.PubKey{leaf}}

	// Each parallel subtest gets its own env: fund writes accounts into the
	// keeper's iavl tree, so a shared env races and drops one subtest's
	// SetAccount against the other's.
	setup := func(t *testing.T) (sdk.Context, sdk.AnteHandler, func(crypto.PubKey) (std.Tx, []byte)) {
		t.Helper()
		env := setupTestEnv()
		ctx := env.ctx
		anteHandler := NewAnteHandler(env.acck, env.bankk, DefaultSigVerificationGasConsumer, defaultAnteOptions())
		fund := func(pub crypto.PubKey) (std.Tx, []byte) {
			addr := pub.Address()
			fee := tu.NewTestFee()
			acc := env.acck.NewAccountWithAddress(ctx, addr)
			require.NoError(t, acc.SetCoins(std.Coins{fee.GasFee}))
			env.acck.SetAccount(ctx, acc)

			tx := std.Tx{Msgs: []std.Msg{tu.NewTestMsg(addr)}, Fee: fee}
			signBytes, err := tx.GetSignBytes(ctx.ChainID(), acc.GetAccountNumber(), acc.GetSequence())
			require.NoError(t, err)
			return tx, signBytes
		}
		return ctx, anteHandler, fund
	}

	t.Run("two distinct leaves pass", func(t *testing.T) {
		t.Parallel()

		ctx, anteHandler, fund := setup(t)
		otherPriv := ed25519.GenPrivKey()
		pub := multisig.PubKeyMultisigThreshold{K: 2, PubKeys: []crypto.PubKey{leaf, otherPriv.PubKey()}}
		tx, signBytes := fund(pub)

		sig1, err := priv.Sign(signBytes)
		require.NoError(t, err)
		sig2, err := otherPriv.Sign(signBytes)
		require.NoError(t, err)
		msig := multisig.NewMultisig(2)
		require.NoError(t, msig.AddSignature(sig1, 0))
		require.NoError(t, msig.AddSignature(sig2, 1))
		tx.Signatures = []std.Signature{{PubKey: pub, Signature: msig.Marshal()}}

		var decoded std.Tx
		amino.MustUnmarshal(amino.MustMarshal(tx), &decoded)
		checkValidTx(t, anteHandler, ctx, decoded, false)
	})

	t.Run("one signer behind two positions is rejected", func(t *testing.T) {
		t.Parallel()

		ctx, anteHandler, fund := setup(t)
		pub := multisig.PubKeyMultisigThreshold{K: 2, PubKeys: []crypto.PubKey{leaf, wrapped}}
		tx, signBytes := fund(pub)

		leafSig, err := priv.Sign(signBytes)
		require.NoError(t, err)

		// Wrap the second signature for the nested multisig.
		inner := multisig.NewMultisig(1)
		require.NoError(t, inner.AddSignature(leafSig, 0))
		outer := multisig.NewMultisig(2)
		require.NoError(t, outer.AddSignature(leafSig, 0))
		require.NoError(t, outer.AddSignature(inner.Marshal(), 1))
		tx.Signatures = []std.Signature{{PubKey: pub, Signature: outer.Marshal()}}

		require.NoError(t, pub.ValidateStructure())
		require.False(t, leaf.Equals(wrapped),
			"precondition: a leaf and its wrapper are never Equals")
		require.True(t, wrapped.VerifyBytes(signBytes, inner.Marshal()),
			"precondition: the wrapper answers with the same signature")

		// Decode the transaction through its wire format before running the ante handler.
		var decoded std.Tx
		amino.MustUnmarshal(amino.MustMarshal(tx), &decoded)
		checkInvalidTx(t, anteHandler, ctx, decoded, false, std.UnauthorizedError{})
	})
}
