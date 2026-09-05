package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
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
// list by the bit array of the caller-supplied Multisignature, which is only
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
		{
			// Rejected for the shape rather than the contents: the trailing
			// signature is a valid one that the loop simply never indexes.
			// VerifyBytes rejects it too, so that the encoding stays canonical.
			"more signatures than set bits",
			multisig.Multisignature{BitArray: bitArray(2, 0), Sigs: sigSet},
		},
		{
			// The bit array's own shape, which every clause above takes on
			// trust. Its size is 2 and one bit is set, so the counts line up;
			// what does not is the final byte's low six bits, which no part of
			// verification reads. Setting them is the same malleability
			// appending a signature is.
			"bit array with bits set past its end",
			multisig.Multisignature{
				BitArray: &bitarray.CompactBitArray{ExtraBitsStored: 2, Elems: []byte{0b1000_0001}},
				Sigs:     sigSet[:1],
			},
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

// The signature carried by a transaction is caller-supplied bytes, so
// decoding it as a Multisignature has to be able to fail. It used to be a
// MustUnmarshal, which panicked past the ante handler's own recover — that
// only handles OutOfGasError — and landed in runTx's blanket recover as an
// ErrInternal carrying a stack trace.
func TestDefaultSigVerificationGasConsumerMalformedMultisignature(t *testing.T) {
	t.Parallel()

	pkSet, _ := generatePubKeysAndSignatures(2, []byte{1, 2, 3, 4}, false)
	pubkey := multisig.PubKeyMultisigThreshold{K: 1, PubKeys: pkSet}

	for _, sig := range [][]byte{
		{0xff, 0xff, 0xff, 0xff},
		{},
		[]byte("not amino"),
	} {
		require.NotPanics(t, func() {
			res := DefaultSigVerificationGasConsumer(
				store.NewInfiniteGasMeter(), sig, pubkey, DefaultParams())
			require.False(t, res.IsOK())
		}, "signature %X", sig)
	}
}

// nestedMultisigKey returns a chain of depth threshold keys over a single
// ed25519 leaf, along with that leaf's private key.
//
// Nothing upstream of the signature check bounds the depth: std.CountSubKeys,
// which is what ValidateSigCount limits, counts leaves, and this key has
// exactly one however deep the chain runs.
func nestedMultisigKey(t *testing.T, depth int) (crypto.PrivKey, crypto.PubKey) {
	t.Helper()

	priv := ed25519.GenPrivKey()
	var pub crypto.PubKey = priv.PubKey()
	for range depth {
		pub = multisig.PubKeyMultisigThreshold{K: 1, PubKeys: []crypto.PubKey{pub}}
	}
	require.Equal(t, 1, std.CountSubKeys(pub))
	return priv, pub
}

// newFundedTx returns an unsigned transaction from a newly created account at
// pub's address, funded with just the gas fee.
func newFundedTx(t *testing.T, env testEnv, pub crypto.PubKey) (std.Tx, std.Account) {
	t.Helper()

	addr := pub.Address()
	fee := tu.NewTestFee()
	acc := env.acck.NewAccountWithAddress(env.ctx, addr)
	require.NoError(t, acc.SetCoins(std.Coins{fee.GasFee}))
	env.acck.SetAccount(env.ctx, acc)

	return std.Tx{Msgs: []std.Msg{tu.NewTestMsg(addr)}, Fee: fee}, acc
}

// nestedMultisigTx returns a transaction from an account keyed by a chain of
// depth threshold keys, signed by the leaf and wrapped once per level.
func nestedMultisigTx(t *testing.T, env testEnv, depth int) std.Tx {
	t.Helper()

	priv, pub := nestedMultisigKey(t, depth)
	tx, acc := newFundedTx(t, env, pub)

	signBytes, err := tx.GetSignBytes(env.ctx.ChainID(), acc.GetAccountNumber(), acc.GetSequence())
	require.NoError(t, err)

	sig, err := priv.Sign(signBytes)
	require.NoError(t, err)
	for range depth {
		outer := multisig.NewMultisig(1)
		outer.AddSignature(sig, 0)
		sig = outer.Marshal()
	}

	tx.Signatures = []std.Signature{{PubKey: pub, Signature: sig}}
	return tx
}

// Verifying a nested threshold key decodes the whole of the remaining
// signature once per level and charges nothing for the level itself, so an
// unbounded chain costs a validator time quadratic in a transaction it pays
// for linearly. multisig.MaxNestingDepth is what bounds it.
func TestAnteHandlerMultisigNestingDepth(t *testing.T) {
	t.Parallel()

	t.Run("within the bound", func(t *testing.T) {
		t.Parallel()

		env := setupTestEnv()
		anteHandler := NewAnteHandler(env.acck, env.bankk, DefaultSigVerificationGasConsumer, defaultAnteOptions())
		checkValidTx(t, anteHandler, env.ctx, nestedMultisigTx(t, env, multisig.MaxNestingDepth), false)
	})

	t.Run("beyond the bound", func(t *testing.T) {
		t.Parallel()

		env := setupTestEnv()
		anteHandler := NewAnteHandler(env.acck, env.bankk, DefaultSigVerificationGasConsumer, defaultAnteOptions())
		tx := nestedMultisigTx(t, env, multisig.MaxNestingDepth+1)
		checkInvalidTx(t, anteHandler, env.ctx, tx, false, std.InvalidPubKeyError{})
	})

	t.Run("a long chain is rejected on the key alone", func(t *testing.T) {
		t.Parallel()

		// The attack payload is a chain long enough that decoding its
		// signature once per level is quadratic in the size of the
		// transaction. Rejection has to come from the key, before any of that
		// signature is decoded, so this passes a signature that is not one:
		// reaching the decode at all would report ErrUnauthorized instead.
		env := setupTestEnv()
		anteHandler := NewAnteHandler(env.acck, env.bankk, DefaultSigVerificationGasConsumer, defaultAnteOptions())

		_, pub := nestedMultisigKey(t, 20_000)
		tx, _ := newFundedTx(t, env, pub)
		tx.Signatures = []std.Signature{{PubKey: pub, Signature: []byte{0xff, 0xff}}}

		checkInvalidTx(t, anteHandler, env.ctx, tx, false, std.InvalidPubKeyError{})
	})
}

// A constituent key that holds no keys of its own contributes no leaves, so
// std.CountSubKeys counts zero for a key of them however wide it is, and
// ValidateSigCount never fires: the blind spot that leaves nesting depth
// unbounded leaves the width unbounded too. The width is the more expensive of
// the two, because verification scans each level's constituents pairwise to
// reject duplicates — 36,000 of them fit inside MaxTxBytes and cost seconds of
// ante handler time, which CheckTx does not charge for.
// multisig.MaxTotalKeys is what bounds it.
func TestAnteHandlerMultisigTotalKeys(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	anteHandler := NewAnteHandler(env.acck, env.bankk, DefaultSigVerificationGasConsumer, defaultAnteOptions())

	subs := make([]crypto.PubKey, 36_000)
	for i := range subs {
		// Pairwise distinct, so the duplicate scan compares every pair, and
		// leafless, so none of them counts against TxSigLimit.
		subs[i] = multisig.PubKeyMultisigThreshold{K: uint(i + 1)}
	}
	pub := multisig.PubKeyMultisigThreshold{K: 1, PubKeys: subs}
	require.Equal(t, 0, std.CountSubKeys(pub), "counts no subkeys, so TxSigLimit never fires")

	tx, _ := newFundedTx(t, env, pub)
	// Not one bit of the bit array is set: the caller needs no signature at
	// all, valid or otherwise, for the key itself to be walked.
	tx.Signatures = []std.Signature{{PubKey: pub, Signature: multisig.NewMultisig(len(subs)).Marshal()}}

	// Unlike a deep chain, this shape survives the amino decode a node performs
	// on an incoming transaction, so round-trip it the way one arrives.
	var decoded std.Tx
	require.NoError(t, amino.Unmarshal(amino.MustMarshal(tx), &decoded))

	checkInvalidTx(t, anteHandler, env.ctx, decoded, false, std.InvalidPubKeyError{})
}

// The other half of "the signature field is untrusted bytes", and the half a
// failed amino decode does not cover: a bit array whose ExtraBitsStored runs
// past its Elems decodes perfectly well, and only then makes Size() claim more
// bits than are stored — which the gas consumer walks, indexing off the end of
// Elems. It panicked past the ante handler's own recover, which handles only
// OutOfGasError, and landed in runTx's blanket recover as an ErrInternal
// carrying a stack trace: the same failure mode amino.MustUnmarshal on the
// signature had, reached by a payload that decodes cleanly.
func TestAnteHandlerMultisigMalformedBitArray(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	anteHandler := NewAnteHandler(env.acck, env.bankk, DefaultSigVerificationGasConsumer, defaultAnteOptions())

	// Nine constituents holding no keys of their own, so that CountSubKeys
	// counts zero and ValidateSigCount never fires, and so that ten keys at
	// depth two sit inside the structural bounds. What is left is the bit array.
	subs := make([]crypto.PubKey, 9)
	for i := range subs {
		subs[i] = multisig.PubKeyMultisigThreshold{K: uint(i + 1)}
	}
	pub := multisig.PubKeyMultisigThreshold{K: 1, PubKeys: subs}
	require.Equal(t, 0, std.CountSubKeys(pub), "counts no subkeys, so TxSigLimit never fires")
	require.NoError(t, pub.ValidateStructure(), "and the structural bounds admit it")

	// Size() is (1-1)*8+9 == 9, matching the nine constituents, over one byte.
	sig := multisig.Multisignature{
		BitArray: &bitarray.CompactBitArray{ExtraBitsStored: 9, Elems: []byte{0x00}},
	}
	tx, _ := newFundedTx(t, env, pub)
	tx.Signatures = []std.Signature{{PubKey: pub, Signature: amino.MustMarshal(sig)}}

	// Round-trip it the way a node receives it: what makes this reachable is
	// that the whole payload decodes without complaint.
	var decoded std.Tx
	require.NoError(t, amino.Unmarshal(amino.MustMarshal(tx), &decoded))

	checkInvalidTx(t, anteHandler, env.ctx, decoded, false, std.UnauthorizedError{})
}

// The third half of "the key field is untrusted bytes", and the one neither the
// structural bounds nor the bit array checks cover: validate nil-checked its own
// constituents, but Equals recurses into the constituents of the keys it
// compares, so a nil nested one level below the top was reached by the duplicate
// scan and dereferenced. The panic escaped the ante handler's own recover, which
// handles only OutOfGasError, into runTx's blanket recover as an ErrInternal
// carrying a stack trace — the same failure mode MustUnmarshal on the signature
// and the malformed bit array had.
func TestAnteHandlerMultisigNilNestedConstituent(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	anteHandler := NewAnteHandler(env.acck, env.bankk, DefaultSigVerificationGasConsumer, defaultAnteOptions())

	// Two nested threshold keys with equal K and equal len(PubKeys), so Equals
	// gets past its cheap comparisons, and the nil in the later one, which is the
	// receiver of Equals in the scan.
	leaf := ed25519.GenPrivKey().PubKey()
	pub := multisig.PubKeyMultisigThreshold{K: 1, PubKeys: []crypto.PubKey{
		multisig.PubKeyMultisigThreshold{K: 1, PubKeys: []crypto.PubKey{leaf}},
		multisig.PubKeyMultisigThreshold{K: 1, PubKeys: []crypto.PubKey{nil}},
	}}
	// Provably past every guard upstream of the signature check, so that what
	// rejects it is the nil and not an earlier bound.
	require.LessOrEqual(t, int64(std.CountSubKeys(pub)), DefaultParams().TxSigLimit,
		"counts inside TxSigLimit, so ValidateSigCount never fires")

	// No bits set and no signatures: the counts line up, so the gas consumer
	// passes and the key alone is what is left to reject it.
	sig := multisig.Multisignature{
		BitArray: &bitarray.CompactBitArray{ExtraBitsStored: 2, Elems: []byte{0x00}},
	}
	tx, _ := newFundedTx(t, env, pub)
	tx.Signatures = []std.Signature{{PubKey: pub, Signature: amino.MustMarshal(sig)}}

	// Round-trip it the way a node receives it: amino writes a nil interface
	// element and reads it back, so this key arrives intact from the wire.
	var decoded std.Tx
	require.NoError(t, amino.Unmarshal(amino.MustMarshal(tx), &decoded))
	require.Equal(t,
		[]crypto.PubKey{nil},
		decoded.Signatures[0].PubKey.(multisig.PubKeyMultisigThreshold).
			PubKeys[1].(multisig.PubKeyMultisigThreshold).PubKeys,
		"the nil constituent survives the wire round-trip")

	checkInvalidTx(t, anteHandler, env.ctx, decoded, false, std.InvalidPubKeyError{})
}
