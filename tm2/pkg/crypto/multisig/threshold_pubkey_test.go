package multisig

import (
	stded25519 "crypto/ed25519"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig/bitarray"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

// This tests multisig functionality, but it expects the first k signatures to be valid
// TODO: Adapt it to give more flexibility about first k signatures being valid
func TestThresholdMultisigValidCases(t *testing.T) {
	t.Parallel()

	pkSet1, sigSet1 := generatePubKeysAndSignatures(5, []byte{1, 2, 3, 4})
	cases := []struct {
		msg            []byte
		k              int
		pubkeys        []crypto.PubKey
		signingIndices []int
		// signatures should be the same size as signingIndices.
		signatures           [][]byte
		passAfterKSignatures []bool
	}{
		{
			msg:                  []byte{1, 2, 3, 4},
			k:                    2,
			pubkeys:              pkSet1,
			signingIndices:       []int{0, 3, 1},
			signatures:           sigSet1,
			passAfterKSignatures: []bool{false},
		},
	}
	for tcIndex, tc := range cases {
		multisigKey := NewPubKeyMultisigThreshold(tc.k, tc.pubkeys)
		multisignature := NewMultisig(len(tc.pubkeys))

		for i := range tc.k - 1 {
			signingIndex := tc.signingIndices[i]
			require.NoError(
				t,
				multisignature.AddSignatureFromPubKey(tc.signatures[signingIndex], tc.pubkeys[signingIndex], tc.pubkeys),
			)
			require.False(
				t,
				multisigKey.VerifyBytes(tc.msg, amino.MustMarshal(multisignature)),
				"multisig passed when i < k, tc %d, i %d", tcIndex, i,
			)
			require.NoError(
				t,
				multisignature.AddSignatureFromPubKey(tc.signatures[signingIndex], tc.pubkeys[signingIndex], tc.pubkeys),
			)
			require.Equal(
				t,
				i+1,
				len(multisignature.Sigs),
				"adding a signature for the same pubkey twice increased signature count by 2, tc %d", tcIndex,
			)
		}

		require.False(
			t,
			multisigKey.VerifyBytes(tc.msg, amino.MustMarshal(multisignature)),
			"multisig passed with k - 1 sigs, tc %d", tcIndex,
		)
		require.NoError(
			t,
			multisignature.AddSignatureFromPubKey(tc.signatures[tc.signingIndices[tc.k]], tc.pubkeys[tc.signingIndices[tc.k]], tc.pubkeys),
		)
		require.True(
			t,
			multisigKey.VerifyBytes(tc.msg, amino.MustMarshal(multisignature)),
			"multisig failed after k good signatures, tc %d", tcIndex,
		)

		for i := tc.k + 1; i < len(tc.signingIndices); i++ {
			signingIndex := tc.signingIndices[i]

			require.NoError(
				t,
				multisignature.AddSignatureFromPubKey(tc.signatures[signingIndex], tc.pubkeys[signingIndex], tc.pubkeys),
			)
			require.Equal(
				t,
				tc.passAfterKSignatures[i-tc.k-1],
				multisigKey.VerifyBytes(tc.msg, amino.MustMarshal(multisignature)),
				"multisig didn't verify as expected after k sigs, tc %d, i %d", tcIndex, i,
			)
			require.NoError(
				t,
				multisignature.AddSignatureFromPubKey(tc.signatures[signingIndex], tc.pubkeys[signingIndex], tc.pubkeys),
			)
			require.Equal(
				t,
				i+1,
				len(multisignature.Sigs),
				"adding a signature for the same pubkey twice increased signature count by 2, tc %d", tcIndex,
			)
		}
	}
}

// TODO: Fully replace this test with table driven tests
func TestThresholdMultisigDuplicateSignatures(t *testing.T) {
	t.Parallel()

	msg := []byte{1, 2, 3, 4, 5}
	pubkeys, sigs := generatePubKeysAndSignatures(5, msg)
	multisigKey := NewPubKeyMultisigThreshold(2, pubkeys)
	multisignature := NewMultisig(5)
	require.False(t, multisigKey.VerifyBytes(msg, amino.MustMarshal(multisignature)))
	multisignature.AddSignatureFromPubKey(sigs[0], pubkeys[0], pubkeys)
	// Add second signature manually
	multisignature.Sigs = append(multisignature.Sigs, sigs[0])
	require.False(t, multisigKey.VerifyBytes(msg, amino.MustMarshal(multisignature)))
}

// Every signature here is valid, and any K of them verify. Accepting more than
// K would let whoever relays the signature drop the spares and still verify.
func TestThresholdMultisigTakesExactlyK(t *testing.T) {
	t.Parallel()

	msg := []byte{1, 2, 3, 4, 5}
	pubkeys, sigs := generatePubKeysAndSignatures(4, msg)
	multisigKey := NewPubKeyMultisigThreshold(2, pubkeys)
	signedBy := func(idx ...int) []byte {
		multisignature := NewMultisig(len(pubkeys))
		for _, i := range idx {
			require.NoError(t, multisignature.AddSignatureFromPubKey(sigs[i], pubkeys[i], pubkeys))
		}
		return multisignature.Marshal()
	}

	require.True(t, multisigKey.VerifyBytes(msg, signedBy(0, 2)))
	require.True(t, multisigKey.VerifyBytes(msg, signedBy(1, 3)))
	require.False(t, multisigKey.VerifyBytes(msg, signedBy(0, 1, 2)))
	require.False(t, multisigKey.VerifyBytes(msg, signedBy(0, 1, 2, 3)))
}

// A multisignature is amino nested in the transaction's signature bytes, which
// the transaction's canonical check sees as opaque. Padding a varint in it
// decodes to the same multisignature, at any nesting level, and must not
// verify.
func TestVerifyBytesTakesOnlyCanonicalMultisignature(t *testing.T) {
	t.Parallel()

	msg := []byte{1, 2, 3, 4, 5}
	pubkeys, sigs := generatePubKeysAndSignatures(2, msg)
	multisigKey := NewPubKeyMultisigThreshold(1, pubkeys)
	multisignature := NewMultisig(len(pubkeys))
	require.NoError(t, multisignature.AddSignatureFromPubKey(sigs[0], pubkeys[0], pubkeys))
	canonical := multisignature.Marshal()
	padded := padBitArrayVarint(t, canonical)

	var decoded Multisignature
	require.NoError(t, amino.Unmarshal(padded, &decoded))
	require.Equal(t, *multisignature, decoded)

	require.True(t, multisigKey.VerifyBytes(msg, canonical))
	require.False(t, multisigKey.VerifyBytes(msg, padded))

	outerPubs := []crypto.PubKey{multisigKey}
	outerKey := NewPubKeyMultisigThreshold(1, outerPubs)
	outer := func(inner []byte) []byte {
		ms := NewMultisig(1)
		require.NoError(t, ms.AddSignatureFromPubKey(inner, multisigKey, outerPubs))
		return ms.Marshal()
	}
	require.True(t, outerKey.VerifyBytes(msg, outer(canonical)))
	require.False(t, outerKey.VerifyBytes(msg, outer(padded)))
}

// padBitArrayVarint writes a multisignature's ExtraBitsStored as a two-byte
// varint, so its BitArray field grows by one byte: 0a L 08 E becomes
// 0a L+1 08 E|80 00.
func padBitArrayVarint(t *testing.T, bz []byte) []byte {
	t.Helper()
	require.True(t, bz[0] == 0x0a && bz[1] < 0x7f && bz[2] == 0x08 && bz[3] < 0x80, "% x", bz[:4])
	return append([]byte{0x0a, bz[1] + 1, 0x08, bz[3] | 0x80, 0x00}, bz[4:]...)
}

// TODO: Fully replace this test with table driven tests
func TestMultiSigPubKeyEquality(t *testing.T) {
	t.Parallel()

	msg := []byte{1, 2, 3, 4}
	pubkeys, _ := generatePubKeysAndSignatures(5, msg)
	multisigKey := NewPubKeyMultisigThreshold(2, pubkeys)
	var unmarshalledMultisig crypto.PubKey
	amino.MustUnmarshal(multisigKey.Bytes(), &unmarshalledMultisig)
	require.True(t, multisigKey.Equals(unmarshalledMultisig))

	// Ensure that reordering pubkeys is treated as a different pubkey
	pubkeysCpy := make([]crypto.PubKey, 5)
	copy(pubkeysCpy, pubkeys)
	pubkeysCpy[4] = pubkeys[3]
	pubkeysCpy[3] = pubkeys[4]
	multisigKey2 := NewPubKeyMultisigThreshold(2, pubkeysCpy)
	require.False(t, multisigKey.Equals(multisigKey2))
}

func TestAddress(t *testing.T) {
	t.Parallel()

	msg := []byte{1, 2, 3, 4}
	pubkeys, _ := generatePubKeysAndSignatures(5, msg)
	multisigKey := NewPubKeyMultisigThreshold(2, pubkeys)
	require.Len(t, multisigKey.Address(), 20)
}

func TestPubKeyMultisigThresholdAminoToIface(t *testing.T) {
	t.Parallel()

	msg := []byte{1, 2, 3, 4}
	pubkeys, _ := generatePubKeysAndSignatures(5, msg)
	multisigKey := NewPubKeyMultisigThreshold(2, pubkeys)

	ab, err := amino.MarshalAnySized(multisigKey)
	require.NoError(t, err)
	// like other crypto.Pubkey implementations (e.g. ed25519.PubKeyEd25519),
	// PubKeyMultisigThreshold should be deserializable into a crypto.PubKey:
	var pubKey crypto.PubKey
	err = amino.UnmarshalSized(ab, &pubKey)
	require.NoError(t, err)

	require.Equal(t, multisigKey, pubKey)
}

// A PubKeyMultisigThreshold decoded from a transaction signature never went
// through NewPubKeyMultisigThreshold, so VerifyBytes cannot assume K and
// PubKeys hold sane values.
func TestVerifyBytesRejectsMalformedPubKey(t *testing.T) {
	t.Parallel()

	msg := []byte{1, 2, 3, 4}
	pubkeys, sigs := generatePubKeysAndSignatures(2, msg)

	t.Run("zero threshold", func(t *testing.T) {
		t.Parallel()

		// K is omitted from the encoding when it is zero, so this key costs an
		// caller nothing to produce and its address is fixed and public.
		var pk crypto.PubKey
		amino.MustUnmarshalAny(amino.MustMarshalAny(PubKeyMultisigThreshold{}), &pk)
		require.Equal(t, uint(0), pk.(PubKeyMultisigThreshold).K)

		require.False(t, pk.VerifyBytes(msg, nil))
		require.False(t, pk.VerifyBytes(msg, amino.MustMarshal(NewMultisig(0))))
	})

	t.Run("zero threshold with pubkeys", func(t *testing.T) {
		t.Parallel()

		pk := PubKeyMultisigThreshold{K: 0, PubKeys: pubkeys}
		multisignature := NewMultisig(len(pubkeys))
		require.False(t, pk.VerifyBytes(msg, amino.MustMarshal(multisignature)))

		require.NoError(t, multisignature.AddSignatureFromPubKey(sigs[0], pubkeys[0], pubkeys))
		require.False(t, pk.VerifyBytes(msg, amino.MustMarshal(multisignature)))
	})

	t.Run("threshold above key count", func(t *testing.T) {
		t.Parallel()

		pk := PubKeyMultisigThreshold{K: uint(len(pubkeys)) + 1, PubKeys: pubkeys}
		multisignature := NewMultisig(len(pubkeys))
		for i := range pubkeys {
			require.NoError(t, multisignature.AddSignatureFromPubKey(sigs[i], pubkeys[i], pubkeys))
		}
		require.False(t, pk.VerifyBytes(msg, amino.MustMarshal(multisignature)))
	})

	t.Run("duplicate constituent keys", func(t *testing.T) {
		t.Parallel()

		// Repeating a key lets its single holder fill two threshold positions.
		duplicated := []crypto.PubKey{pubkeys[0], pubkeys[0], pubkeys[1]}
		pk := PubKeyMultisigThreshold{K: 2, PubKeys: duplicated}

		multisignature := NewMultisig(len(duplicated))
		multisignature.AddSignature(sigs[0], 0)
		multisignature.AddSignature(sigs[0], 1)
		require.False(t, pk.VerifyBytes(msg, amino.MustMarshal(multisignature)))
	})

	t.Run("nil constituent key", func(t *testing.T) {
		t.Parallel()

		// Reachable over the wire: amino JSON decodes a null element into a nil
		// interface, which used to be dereferenced by the verification loop.
		var pk crypto.PubKey
		require.NoError(t, amino.UnmarshalJSON(
			[]byte(`{"@type":"/tm.PubKeyMultisig","threshold":"1","pubkeys":[null]}`), &pk))
		require.Equal(t, []crypto.PubKey{nil}, pk.(PubKeyMultisigThreshold).PubKeys)

		multisignature := NewMultisig(1)
		multisignature.AddSignature(sigs[0], 0)
		require.NotPanics(t, func() {
			require.False(t, pk.VerifyBytes(msg, amino.MustMarshal(multisignature)))
		})
	})
}

func TestVerifyBytesRejectsRepeatedLeafAcrossNestedPositions(t *testing.T) {
	t.Parallel()

	msg := []byte("repeated leaf")
	privA := ed25519.GenPrivKey()
	privB := ed25519.GenPrivKey()
	privC := ed25519.GenPrivKey()
	pubA, pubB, pubC := privA.PubKey(), privB.PubKey(), privC.PubKey()
	sigA, err := privA.Sign(msg)
	require.NoError(t, err)
	sigB, err := privB.Sign(msg)
	require.NoError(t, err)

	wrap := func(pub crypto.PubKey, sig []byte, levels int) (crypto.PubKey, []byte) {
		for range levels {
			pub = PubKeyMultisigThreshold{K: 1, PubKeys: []crypto.PubKey{pub}}
			wrapped := NewMultisig(1)
			require.NoError(t, wrapped.AddSignature(sig, 0))
			sig = wrapped.Marshal()
		}
		return pub, sig
	}
	assemble := func(sigs ...[]byte) []byte {
		assembled := NewMultisig(len(sigs))
		for i, sig := range sigs {
			require.NoError(t, assembled.AddSignature(sig, i))
		}
		return assembled.Marshal()
	}

	oneA, oneASig := wrap(pubA, sigA, 1)
	twoA, twoASig := wrap(pubA, sigA, 2)
	threeA, threeASig := wrap(pubA, sigA, 3)
	oneB, oneBSig := wrap(pubB, sigB, 1)
	deepA, deepASig := wrap(pubA, sigA, MaxNestingDepth-1)
	oneOfN := PubKeyMultisigThreshold{K: 1, PubKeys: []crypto.PubKey{pubA, pubB}}
	oneOfNSig := func() []byte {
		sig := NewMultisig(2)
		require.NoError(t, sig.AddSignature(sigA, 0))
		return sig.Marshal()
	}()
	left := oneOfN
	right := PubKeyMultisigThreshold{K: 1, PubKeys: []crypto.PubKey{pubA, pubC}}

	tests := []struct {
		name string
		pub  PubKeyMultisigThreshold
		sig  []byte
	}{
		{
			name: "leaf and one-of-one wrapper",
			pub:  PubKeyMultisigThreshold{K: 2, PubKeys: []crypto.PubKey{pubA, oneA}},
			sig:  assemble(sigA, oneASig),
		},
		{
			name: "leaf and one-of-n subtree",
			pub:  PubKeyMultisigThreshold{K: 2, PubKeys: []crypto.PubKey{pubA, oneOfN}},
			sig:  assemble(sigA, oneOfNSig),
		},
		{
			name: "same leaf in different subtrees",
			pub:  PubKeyMultisigThreshold{K: 2, PubKeys: []crypto.PubKey{left, right}},
			sig:  assemble(oneOfNSig, oneOfNSig),
		},
		{
			name: "maximum depth wrapper",
			pub:  PubKeyMultisigThreshold{K: 2, PubKeys: []crypto.PubKey{pubA, deepA}},
			sig:  assemble(sigA, deepASig),
		},
		{
			name: "four positions backed by one leaf",
			pub:  PubKeyMultisigThreshold{K: 4, PubKeys: []crypto.PubKey{pubA, oneA, twoA, threeA}},
			sig:  assemble(sigA, oneASig, twoASig, threeASig),
		},
		{
			name: "four positions backed by two leaves",
			pub:  PubKeyMultisigThreshold{K: 4, PubKeys: []crypto.PubKey{pubA, oneA, pubB, oneB}},
			sig:  assemble(sigA, oneASig, sigB, oneBSig),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.NoError(t, tt.pub.ValidateStructure())
			require.False(t, tt.pub.VerifyBytes(msg, tt.sig))
		})
	}
}

// Empty threshold nodes hold no leaves of their own, so validate treats each as
// a terminal leaf and two of them collide. Without that, {K:2,[e,e]} with
// e={K:1} slips past the duplicate scan: the per-level Equals check this scan
// replaced did catch it, and the leaf walk has to keep catching it.
func TestValidateRejectsDuplicateEmptySubtrees(t *testing.T) {
	t.Parallel()

	e := PubKeyMultisigThreshold{K: 1, PubKeys: nil}
	pk := PubKeyMultisigThreshold{K: 2, PubKeys: []crypto.PubKey{e, e}}

	require.NoError(t, pk.ValidateStructure(),
		"structure alone does not reject it; the duplicate scan must")
	require.EqualError(t, pk.validate(),
		"threshold k of n multisignature: duplicate pubkey")
	_, err := NewPubKeyMultisigThresholdChecked(2, []crypto.PubKey{e, e})
	require.EqualError(t, err,
		"threshold k of n multisignature: duplicate pubkey")
}

// A small-order ed25519 key verifies a forged signature, so it counts as a
// signer that no private key stands behind. The identity accepts S = 0 and
// R = identity for every message, and so does its encoding with the x sign bit
// set, which Equals treats as a different key. Without the small-order check,
// {K:3,[A, id, id']} passes the duplicate scan and one real signature satisfies
// it.
func TestValidateRejectsSmallOrderEd25519Leaf(t *testing.T) {
	t.Parallel()

	var id, idNeg ed25519.PubKeyEd25519
	id[0], idNeg[0] = 1, 1
	idNeg[31] = 0x80
	require.False(t, id.Equals(idNeg), "premise: the two spellings are distinct keys")

	msg := []byte("any sign bytes")
	forged := make([]byte, ed25519.SignatureSize)
	forged[0] = 1 // R = identity, S = 0
	require.True(t, stded25519.Verify(id[:], msg, forged),
		"premise: crypto/ed25519 accepts the forgery")
	require.False(t, id.VerifyBytes(msg, forged),
		"premise: PubKeyEd25519.VerifyBytes rejects small-order keys")

	privA := ed25519.GenPrivKey()
	pk := PubKeyMultisigThreshold{K: 3, PubKeys: []crypto.PubKey{privA.PubKey(), id, idNeg}}
	require.EqualError(t, pk.validate(),
		"threshold k of n multisignature: small-order ed25519 pubkey")

	sigA, err := privA.Sign(msg)
	require.NoError(t, err)
	multisig := NewMultisig(3)
	require.NoError(t, multisig.AddSignature(sigA, 0))
	require.NoError(t, multisig.AddSignature(forged, 1))
	require.NoError(t, multisig.AddSignature(forged, 2))
	require.False(t, pk.VerifyBytes(msg, multisig.Marshal()),
		"one real signer must not satisfy a 3 of 3")

	// Nested below the top level, the leaf is still reached.
	nested := PubKeyMultisigThreshold{K: 2, PubKeys: []crypto.PubKey{
		privA.PubKey(),
		PubKeyMultisigThreshold{K: 1, PubKeys: []crypto.PubKey{id}},
	}}
	require.EqualError(t, nested.validate(),
		"threshold k of n multisignature: small-order ed25519 pubkey")
}

// Leaf distinctness is sufficient for honest K of N but not necessary. Both
// shapes need K distinct signers (A+B, and A+B+C) yet reuse a leaf, so the rule
// rejects them. This deliberately conservative bound is documented in
// adr/pr6504_multisig_unique_leaf_keys.md. The test pins it so a later narrowing
// stays a conscious choice.
func TestValidateOverRejectsGenuineKofN(t *testing.T) {
	t.Parallel()

	pubA := ed25519.GenPrivKey().PubKey()
	pubB := ed25519.GenPrivKey().PubKey()
	pubC := ed25519.GenPrivKey().PubKey()

	// {K:2, [A, {K:2,[A,B]}]} is satisfied only by A and B together.
	andOfAnd := PubKeyMultisigThreshold{K: 2, PubKeys: []crypto.PubKey{
		pubA,
		PubKeyMultisigThreshold{K: 2, PubKeys: []crypto.PubKey{pubA, pubB}},
	}}
	// {K:2, [{K:2,[A,B]}, {K:2,[A,C]}]} is satisfied only by A, B and C together.
	andOfTwoAnds := PubKeyMultisigThreshold{K: 2, PubKeys: []crypto.PubKey{
		PubKeyMultisigThreshold{K: 2, PubKeys: []crypto.PubKey{pubA, pubB}},
		PubKeyMultisigThreshold{K: 2, PubKeys: []crypto.PubKey{pubA, pubC}},
	}}

	for _, pk := range []PubKeyMultisigThreshold{andOfAnd, andOfTwoAnds} {
		require.NoError(t, pk.ValidateStructure())
		require.EqualError(t, pk.validate(),
			"threshold k of n multisignature: duplicate pubkey")
	}
}

func TestNewPubKeyMultisigThresholdInvalid(t *testing.T) {
	t.Parallel()

	pubkeys, _ := generatePubKeysAndSignatures(2, []byte("dummy"))

	require.Panics(t, func() { NewPubKeyMultisigThreshold(0, pubkeys) })
	require.Panics(t, func() { NewPubKeyMultisigThreshold(-1, pubkeys) })
	require.Panics(t, func() { NewPubKeyMultisigThreshold(3, pubkeys) })
	require.Panics(t, func() { NewPubKeyMultisigThreshold(2, []crypto.PubKey{pubkeys[0], nil}) })
	require.Panics(t, func() { NewPubKeyMultisigThreshold(2, []crypto.PubKey{pubkeys[0], pubkeys[0]}) })
}

// The verification loop advances one signature per set bit, so a bit array
// with more set bits than there are signatures used to index off the end of
// Sigs. The earlier signatures have to be valid for the loop to reach the
// overrun at all, which is why this is not covered by the malformed cases
// above: a bad signature returns false before the index is ever reached.
func TestVerifyBytesSignatureCountMismatch(t *testing.T) {
	t.Parallel()

	msg := []byte{1, 2, 3, 4}
	pubkeys, sigs := generatePubKeysAndSignatures(3, msg)
	// K matches the set bits, so the threshold is not what refuses it.
	pk := PubKeyMultisigThreshold{K: 3, PubKeys: pubkeys}

	multisignature := NewMultisig(3)
	for i := range 3 {
		multisignature.BitArray.SetIndex(i, true)
	}
	multisignature.Sigs = sigs[:2] // three set bits, two valid signatures

	require.NotPanics(t, func() {
		require.False(t, pk.VerifyBytes(msg, amino.MustMarshal(multisignature)))
	})
}

// The other side of the same count: signatures past the last set bit are never
// indexed by the verification loop, so accepting them would leave the encoding
// non-canonical. Anyone who sees the transaction could append some, changing
// its bytes and so its hash while leaving what it authorizes untouched, and the
// sender — who is waiting on the hash they submitted — would never see their
// own transaction commit.
//
// The trailing signatures here are valid ones, so that what rejects the second
// multisignature is its shape and not its contents.
func TestVerifyBytesTrailingSignatures(t *testing.T) {
	t.Parallel()

	msg := []byte{1, 2, 3, 4}
	pubkeys, sigs := generatePubKeysAndSignatures(3, msg)
	pk := PubKeyMultisigThreshold{K: 1, PubKeys: pubkeys}

	multisignature := NewMultisig(3)
	require.NoError(t, multisignature.AddSignatureFromPubKey(sigs[0], pubkeys[0], pubkeys))
	canonical := multisignature.Marshal()
	require.True(t, pk.VerifyBytes(msg, canonical),
		"precondition: the canonical encoding of this signature verifies")

	multisignature.Sigs = append(multisignature.Sigs, sigs[1], sigs[2])
	malleated := multisignature.Marshal()
	require.NotEqual(t, canonical, malleated, "precondition: the two encodings differ")
	require.False(t, pk.VerifyBytes(msg, malleated))
}

// The bit array is decoded from the same untrusted bytes as the rest of the
// signature, and Size() is computed from ExtraBitsStored without reference to
// how many bytes Elems holds. Nine bits over one stored byte reports a size of
// nine, matching a nine-key threshold, and the walk over it indexes Elems[1].
func TestVerifyBytesMalformedBitArray(t *testing.T) {
	t.Parallel()

	msg := []byte{1, 2, 3, 4}
	pubkeys, _ := generatePubKeysAndSignatures(9, msg)
	pk := PubKeyMultisigThreshold{K: 1, PubKeys: pubkeys}

	sig := Multisignature{
		BitArray: &bitarray.CompactBitArray{ExtraBitsStored: 9, Elems: []byte{0x00}},
	}
	require.Equal(t, len(pubkeys), sig.BitArray.Size(),
		"precondition: the size check against the constituent keys passes")

	require.NotPanics(t, func() {
		require.False(t, pk.VerifyBytes(msg, amino.MustMarshal(sig)))
	})
}

// The third way the same encoding is malleable, after trailing signatures and
// alongside them: the bits of the final byte past Size() are never read, so
// anyone who sees the transaction can set them, changing its bytes and so its
// hash while leaving what it authorizes untouched.
func TestVerifyBytesBitArrayPadding(t *testing.T) {
	t.Parallel()

	msg := []byte{1, 2, 3, 4}
	pubkeys, sigs := generatePubKeysAndSignatures(3, msg)
	pk := PubKeyMultisigThreshold{K: 1, PubKeys: pubkeys}

	multisignature := NewMultisig(3)
	require.NoError(t, multisignature.AddSignatureFromPubKey(sigs[0], pubkeys[0], pubkeys))
	canonical := multisignature.Marshal()
	require.True(t, pk.VerifyBytes(msg, canonical),
		"precondition: the canonical encoding of this signature verifies")

	// Size() is 3, so the low five bits of the single stored byte are the ones
	// no part of verification ever looks at.
	multisignature.BitArray.Elems[0] |= 0b0001_1111
	malleated := multisignature.Marshal()
	require.Equal(t, len(canonical), len(malleated),
		"precondition: the two encodings differ only in the bits nothing reads")
	require.NotEqual(t, canonical, malleated)
	require.False(t, pk.VerifyBytes(msg, malleated))
}

// nestKey wraps pub in depth-1 further threshold keys, so that the result
// nests depth levels deep in total.
func nestKey(pub crypto.PubKey, depth int) crypto.PubKey {
	for range depth - 1 {
		pub = PubKeyMultisigThreshold{K: 1, PubKeys: []crypto.PubKey{pub}}
	}
	return pub
}

// A chain of threshold keys costs one amino decode of the remaining signature
// per level and is not bounded by auth's TxSigLimit, so the depth is bounded
// here instead.
func TestNestingDepth(t *testing.T) {
	t.Parallel()

	msg := []byte{1, 2, 3, 4}
	pubkeys, sigs := generatePubKeysAndSignatures(2, msg)

	t.Run("within the bound", func(t *testing.T) {
		t.Parallel()

		for depth := 1; depth <= MaxNestingDepth; depth++ {
			inner := PubKeyMultisigThreshold{K: 1, PubKeys: pubkeys}
			pk := nestKey(inner, depth).(PubKeyMultisigThreshold)
			require.NoError(t, pk.ValidateStructure(), "depth %d", depth)

			// And it still verifies: wrap a signature by pubkeys[0] once per
			// level of nesting added on top of inner.
			sig := sigs[0]
			multisignature := NewMultisig(len(pubkeys))
			require.NoError(t, multisignature.AddSignatureFromPubKey(sig, pubkeys[0], pubkeys))
			sig = multisignature.Marshal()
			for range depth - 1 {
				outer := NewMultisig(1)
				outer.AddSignature(sig, 0)
				sig = outer.Marshal()
			}
			require.True(t, pk.VerifyBytes(msg, sig), "depth %d", depth)
		}
	})

	t.Run("beyond the bound", func(t *testing.T) {
		t.Parallel()

		inner := PubKeyMultisigThreshold{K: 1, PubKeys: pubkeys}
		pk := nestKey(inner, MaxNestingDepth+1).(PubKeyMultisigThreshold)

		require.Error(t, pk.ValidateStructure())
		// A key decoded from a transaction signature never goes through the
		// constructor, so VerifyBytes has to reject it on its own.
		require.False(t, pk.VerifyBytes(msg, amino.MustMarshal(NewMultisig(1))))
		require.Panics(t, func() { NewPubKeyMultisigThreshold(1, pk.PubKeys) })
	})

	t.Run("an overlong chain is not walked to the bottom", func(t *testing.T) {
		t.Parallel()

		// The rejection has to happen at the bound, not after descending the
		// whole chain, or the check costs what it is meant to prevent.
		deep := nestKey(PubKeyMultisigThreshold{K: 1, PubKeys: pubkeys}, 100_000).(PubKeyMultisigThreshold)
		require.Error(t, deep.ValidateStructure())
	})
}

// wideKey returns a K=1 threshold key over n constituent threshold keys that
// hold no keys of their own. They are pairwise distinct, because their K
// differs, and each is a leaf of its own.
func wideKey(n int) PubKeyMultisigThreshold {
	subs := make([]crypto.PubKey, n)
	for i := range n {
		subs[i] = PubKeyMultisigThreshold{K: uint(i + 1)}
	}
	return PubKeyMultisigThreshold{K: 1, PubKeys: subs}
}

// validate compares every pair of the key's leaves to reject duplicates, so an
// unbounded key is quadratic in its own width. auth's TxSigLimit bounds the
// width no better than it bounds the depth, and for the same reason:
// std.CountSubKeys counts leaves, and a constituent holding no keys at all
// contributes none however many of them are listed, so a key of them counts
// zero however wide it is. MaxTotalKeys is what bounds it.
func TestTotalKeys(t *testing.T) {
	t.Parallel()

	msg := []byte{1, 2, 3, 4}

	t.Run("at the bound", func(t *testing.T) {
		t.Parallel()

		// The key itself, plus MaxTotalKeys-1 constituents.
		require.NoError(t, wideKey(MaxTotalKeys-1).ValidateStructure())
	})

	t.Run("beyond the bound", func(t *testing.T) {
		t.Parallel()

		pk := wideKey(MaxTotalKeys)
		require.Error(t, pk.ValidateStructure())
		// A key decoded from a transaction signature never goes through the
		// constructor, so VerifyBytes has to reject it on its own.
		require.False(t, pk.VerifyBytes(msg, amino.MustMarshal(NewMultisig(len(pk.PubKeys)))))
		require.Panics(t, func() { NewPubKeyMultisigThreshold(1, pk.PubKeys) })
	})

	t.Run("a very wide key is not scanned pairwise", func(t *testing.T) {
		t.Parallel()

		// The bound has to be enforced before validate's duplicate scan rather
		// than after it, or the check costs what it is meant to prevent: this
		// key takes minutes to reject if every pair of constituents is compared.
		pk := wideKey(100_000)
		require.False(t, pk.VerifyBytes(msg, amino.MustMarshal(NewMultisig(len(pk.PubKeys)))))
	})

	t.Run("every key TxSigLimit permits still validates", func(t *testing.T) {
		t.Parallel()

		// TxSigLimit allows 7 leaves: as a flat 7 of 7,
		pubkeys, _ := generatePubKeysAndSignatures(7, msg)
		require.NotPanics(t, func() { NewPubKeyMultisigThreshold(7, pubkeys) })

		// and as the same 7 spread over a key that branches at every level all
		// the way down to MaxNestingDepth.
		branching := PubKeyMultisigThreshold{K: 1, PubKeys: []crypto.PubKey{
			PubKeyMultisigThreshold{K: 1, PubKeys: []crypto.PubKey{
				PubKeyMultisigThreshold{K: 2, PubKeys: pubkeys[0:2]},
				PubKeyMultisigThreshold{K: 2, PubKeys: pubkeys[2:4]},
			}},
			PubKeyMultisigThreshold{K: 1, PubKeys: pubkeys[4:7]},
		}}
		require.NoError(t, branching.ValidateStructure())
	})
}

func generatePubKeysAndSignatures(n int, msg []byte) (pubkeys []crypto.PubKey, signatures [][]byte) {
	pubkeys = make([]crypto.PubKey, n)
	signatures = make([][]byte, n)
	for i := range n {
		var privkey crypto.PrivKey
		if rand.Int63()%2 == 0 {
			privkey = ed25519.GenPrivKey()
		} else {
			privkey = secp256k1.GenPrivKey()
		}
		pubkeys[i] = privkey.PubKey()
		signatures[i], _ = privkey.Sign(msg)
	}
	return
}

func TestPubKeyMultisigThreshold_String(t *testing.T) {
	t.Parallel()

	t.Run("empty set", func(t *testing.T) {
		t.Parallel()

		pk := PubKeyMultisigThreshold{
			PubKeys: make([]crypto.PubKey, 0), // empty
		}

		assert.Equal(t, "[]", pk.String())
	})

	t.Run("multiple keys", func(t *testing.T) {
		t.Parallel()

		var (
			keys, _ = generatePubKeysAndSignatures(10, []byte("dummy"))
			pk      = NewPubKeyMultisigThreshold(5, keys)
		)

		output := pk.String()

		for _, key := range keys {
			assert.Contains(t, output, key.String())
		}
	})
}

// validate nil-checked only its own constituents, but Equals recurses into the
// constituents of the keys it compares, so a nil nested below the top level was
// reached by the duplicate scan and dereferenced. ValidateStructure is the only
// check that visits every node, so it owns the nil rejection.
func TestNilConstituentAtAnyDepth(t *testing.T) {
	t.Parallel()

	msg := []byte{1, 2, 3, 4}
	pubkeys, _ := generatePubKeysAndSignatures(1, msg)

	// Equal K and equal len(PubKeys) so Equals gets past its cheap comparisons,
	// and the nil in the later constituent, which is the receiver of Equals in
	// the scan.
	nested := PubKeyMultisigThreshold{
		K: 1,
		PubKeys: []crypto.PubKey{
			PubKeyMultisigThreshold{K: 1, PubKeys: pubkeys},
			PubKeyMultisigThreshold{K: 1, PubKeys: []crypto.PubKey{nil}},
		},
	}
	// A bit array of the right size with no bits set and no signatures: the
	// counts line up, so nothing before the duplicate scan rejects it.
	sig := amino.MustMarshal(&Multisignature{
		BitArray: &bitarray.CompactBitArray{ExtraBitsStored: 2, Elems: []byte{0x00}},
	})

	t.Run("rejected rather than dereferenced", func(t *testing.T) {
		t.Parallel()

		require.Error(t, nested.ValidateStructure())
		require.NotPanics(t, func() {
			require.False(t, nested.VerifyBytes(msg, sig))
		})
		require.Panics(t, func() { NewPubKeyMultisigThreshold(1, nested.PubKeys) })
		_, err := NewPubKeyMultisigThresholdChecked(1, nested.PubKeys)
		require.Error(t, err)
	})

	t.Run("a nil at the top level is still rejected", func(t *testing.T) {
		t.Parallel()

		flat := PubKeyMultisigThreshold{K: 1, PubKeys: []crypto.PubKey{nil}}
		require.Error(t, flat.ValidateStructure())
		require.NotPanics(t, func() {
			require.False(t, flat.VerifyBytes(msg, amino.MustMarshal(NewMultisig(1))))
		})
	})

	t.Run("reachable over the wire in both encodings", func(t *testing.T) {
		t.Parallel()

		// Binary is the transaction-signature path: amino writes a nil interface
		// element and reads it back, so this shape arrives intact from the wire.
		var fromBinary crypto.PubKey
		require.NoError(t, amino.UnmarshalAny(amino.MustMarshalAny(nested), &fromBinary))
		require.Equal(t,
			[]crypto.PubKey{nil},
			fromBinary.(PubKeyMultisigThreshold).PubKeys[1].(PubKeyMultisigThreshold).PubKeys)
		require.NotPanics(t, func() {
			require.False(t, fromBinary.VerifyBytes(msg, sig))
		})

		// JSON is the genesis-document path, which gnogenesis verify and
		// gnogenesis fork call VerifyBytes on with no recover of their own.
		var fromJSON crypto.PubKey
		require.NoError(t, amino.UnmarshalJSON(fmt.Appendf(nil,
			`{"@type":"/tm.PubKeyMultisig","threshold":"1","pubkeys":[`+
				`{"@type":"/tm.PubKeyMultisig","threshold":"1","pubkeys":[%s]},`+
				`{"@type":"/tm.PubKeyMultisig","threshold":"1","pubkeys":[null]}]}`,
			amino.MustMarshalJSONAny(pubkeys[0])), &fromJSON))
		require.NotPanics(t, func() {
			require.False(t, fromJSON.VerifyBytes(msg, sig))
		})
	})
}

// MaxNestingDepth admits every key that branches at every level and spends no
// more than TxSigLimit's leaves. The deep shape inside a leaf budget is the
// unbalanced one: each level holds one leaf and one subkey, so no level is a
// degenerate wrapper, and d levels spend d+1 leaves.
func TestBranchingKeyDepth(t *testing.T) {
	t.Parallel()

	msg := []byte{1, 2, 3, 4}
	// auth.DefaultTxSigLimit. std cannot be imported here: it imports multisig.
	const txSigLimit = 7

	// countLeaves is std.CountSubKeys, which is what TxSigLimit is applied to.
	var countLeaves func(crypto.PubKey) int
	countLeaves = func(p crypto.PubKey) int {
		v, ok := p.(PubKeyMultisigThreshold)
		if !ok {
			return 1
		}
		n := 0
		for _, sub := range v.PubKeys {
			n += countLeaves(sub)
		}
		return n
	}

	// caterpillar(d) is d threshold levels, each holding one leaf and (except at
	// the bottom, which holds two leaves) one nested threshold key.
	caterpillar := func(d int) PubKeyMultisigThreshold {
		pubkeys, _ := generatePubKeysAndSignatures(d+1, msg)
		pk := PubKeyMultisigThreshold{K: 2, PubKeys: pubkeys[:2]}
		for i := range d - 1 {
			pk = PubKeyMultisigThreshold{K: 2, PubKeys: []crypto.PubKey{pubkeys[i+2], pk}}
		}
		return pk
	}

	for d := 2; d <= txSigLimit-1; d++ {
		pk := caterpillar(d)

		// No level delegates to a lone subkey, so none of these is degenerate.
		var levels func(crypto.PubKey) int
		levels = func(p crypto.PubKey) int {
			v, ok := p.(PubKeyMultisigThreshold)
			if !ok {
				return 0
			}
			require.GreaterOrEqual(t, len(v.PubKeys), 2, "level with a lone subkey")
			deepest := 0
			for _, sub := range v.PubKeys {
				deepest = max(deepest, levels(sub))
			}
			return deepest + 1
		}
		require.Equal(t, d, levels(pk))
		require.LessOrEqual(t, countLeaves(pk), txSigLimit, "depth %d", d)
		require.NoError(t, pk.ValidateStructure(), "depth %d", d)
	}

	// And the deepest such key is still what the bound stops: one more level
	// needs one more leaf than TxSigLimit allows, so MaxNestingDepth and
	// MaxTotalKeys bite at the same shape.
	require.Equal(t, MaxNestingDepth, txSigLimit-1)
	require.Error(t, caterpillar(MaxNestingDepth+1).ValidateStructure())
}
