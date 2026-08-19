package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/mock"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig/bitarray"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// The point of the test below is that a forgeable key type is decodable from the
// wire, so it has to be registered here for the tx to round-trip.
func init() {
	mock.UnsafeRegisterAminoPackage()
}

// A multisig whose constituent key is a key type the gas consumer does not
// recognize must be rejected before VerifyBytes is reached.
//
// mock.PubKeyMock stands in for such a type here: it satisfies crypto.PubKey,
// and its VerifyBytes accepts a signature anyone can compute from the message
// and the public key. Since a PubKeyMultisigThreshold is only ever as strong as
// its constituent keys, a K=1 multisig over such a key yields an address whose
// transactions can be signed without holding any private key. The multisig case
// of DefaultSigVerificationGasConsumer used to discard the result of the
// recursive per-subkey call, so the rejection never reached the ante handler.
func TestAnteHandlerMultisigUnrecognizedSubkey(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	anteHandler := NewAnteHandler(env.acck, env.bankk, DefaultSigVerificationGasConsumer, defaultAnteOptions())

	// The forger needs no private key: the address is derived from the multisig
	// public key, which is entirely of their choosing.
	mockPub := mock.PrivKeyMock([]byte("forger")).PubKey()
	msPub := multisig.PubKeyMultisigThreshold{K: 1, PubKeys: []crypto.PubKey{mockPub}}
	addr := msPub.Address()

	fee := tu.NewTestFee()
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	require.NoError(t, acc.SetCoins(std.Coins{fee.GasFee}))
	env.acck.SetAccount(ctx, acc)

	tx := std.Tx{Msgs: []std.Msg{tu.NewTestMsg(addr)}, Fee: fee}
	signBytes, err := tx.GetSignBytes(ctx.ChainID(), acc.GetAccountNumber(), acc.GetSequence())
	require.NoError(t, err)

	forged := fmt.Appendf(nil, "signature-for-%X-by-%X", signBytes, []byte(mockPub.(mock.PubKeyMock)))
	msig := multisig.NewMultisig(1)
	msig.AddSignature(forged, 0)
	tx.Signatures = []std.Signature{{PubKey: msPub, Signature: msig.Marshal()}}

	// Round-trip through amino the way a node decoding an incoming tx does; the
	// constituent key is resolved from its type URL against the global codec,
	// bypassing NewPubKeyMultisigThreshold entirely.
	var decoded std.Tx
	amino.MustUnmarshal(amino.MustMarshal(tx), &decoded)
	require.True(t, decoded.Signatures[0].PubKey.Equals(msPub))

	// The forged signature would satisfy VerifyBytes, so the tx must not get
	// that far.
	checkInvalidTx(t, anteHandler, ctx, decoded, false, std.InvalidPubKeyError{})
	require.True(t, msPub.VerifyBytes(signBytes, decoded.Signatures[0].Signature),
		"precondition: the signature this test forges is one VerifyBytes accepts")
}

// The multisig gas consumer indexes both the constituent keys and the signature
// list by the bit array of the attacker-supplied Multisignature, which is only
// checked for consistency later, in VerifyBytes. Mismatched shapes must be
// rejected rather than run off the end of either slice.
func TestConsumeMultisignatureVerificationGasShapeMismatch(t *testing.T) {
	t.Parallel()

	params := DefaultParams()
	msg := []byte{1, 2, 3, 4}

	bitArray := func(size int, set ...int) *bitarray.CompactBitArray {
		ba := bitarray.NewCompactBitArray(size)
		for _, i := range set {
			ba.SetIndex(i, true)
		}
		return ba
	}

	pkSet, sigSet := generatePubKeysAndSignatures(2, msg, false)

	tests := []struct {
		name string
		sig  multisig.Multisignature
	}{
		{
			"bit array larger than the key set",
			multisig.Multisignature{BitArray: bitArray(4, 3), Sigs: sigSet},
		},
		{
			"bit array smaller than the key set",
			multisig.Multisignature{BitArray: bitArray(1, 0), Sigs: sigSet[:1]},
		},
		{
			"more set bits than signatures",
			multisig.Multisignature{BitArray: bitArray(2, 0, 1), Sigs: sigSet[:1]},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pubkey := multisig.PubKeyMultisigThreshold{K: 1, PubKeys: pkSet}
			res := DefaultSigVerificationGasConsumer(
				store.NewInfiniteGasMeter(), amino.MustMarshal(&tt.sig), pubkey, params)
			require.False(t, res.IsOK())
		})
	}
}
