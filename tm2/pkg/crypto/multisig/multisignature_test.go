package multisig

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig/bitarray"
)

// Verification requires exactly one signature per set bit, so an index that sets
// no bit must not append a signature. AddSignature discarded SetIndex's failure
// return and appended anyway, leaving a multisignature that nothing here
// complained about and the chain rejects.
func TestAddSignatureOutOfRange(t *testing.T) {
	t.Parallel()

	sig := []byte("signature")

	for _, tt := range []struct {
		name  string
		size  int
		index int
	}{
		{"one past the end", 3, 3},
		{"well past the end", 3, 5},
		{"negative", 3, -1},
		{"zero-size multisignature", 0, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mSig := NewMultisig(tt.size)
			require.Error(t, mSig.AddSignature(sig, tt.index))
			// And it left the multisignature alone rather than half-updating it.
			require.Empty(t, mSig.Sigs)
			require.Equal(t, 0, mSig.BitArray.NumTrueBitsBefore(mSig.BitArray.Size()))
		})
	}

	t.Run("every in-range index keeps the equality", func(t *testing.T) {
		t.Parallel()

		const size = 5
		mSig := NewMultisig(size)
		// Out of order, and one index twice, to cover the insert and the replace
		// branches as well as the append.
		for _, index := range []int{4, 0, 2, 0} {
			require.NoError(t, mSig.AddSignature(sig, index))
			require.Equal(t,
				mSig.BitArray.NumTrueBitsBefore(mSig.BitArray.Size()), len(mSig.Sigs),
				"after adding index %d", index)
		}
		require.NoError(t, mSig.ValidateBasic(size))
	})

	t.Run("AddSignatureFromPubKey reports it", func(t *testing.T) {
		t.Parallel()

		pubkeys, sigs := generatePubKeysAndSignatures(3, []byte("dummy"))

		// A multisignature sized from somewhere other than the key list it is
		// indexed against: the index is valid for keys and not for the bit array.
		mSig := NewMultisig(2)
		require.Error(t, mSig.AddSignatureFromPubKey(sigs[2], pubkeys[2], pubkeys))
		require.Empty(t, mSig.Sigs)

		// Sized consistently, it works.
		mSig = NewMultisig(len(pubkeys))
		require.NoError(t, mSig.AddSignatureFromPubKey(sigs[2], pubkeys[2], pubkeys))
		require.NoError(t, mSig.ValidateBasic(len(pubkeys)))
	})
}

// ValidateBasic is shared by VerifyBytes and auth's signature gas consumer, so
// that the two cannot come to disagree about which shapes are walkable.
func TestMultisignatureValidateBasic(t *testing.T) {
	t.Parallel()

	sig := []byte("signature")

	for _, tt := range []struct {
		name   string
		mSig   Multisignature
		nKeys  int
		wantOK bool
	}{
		{
			"canonical",
			Multisignature{
				BitArray: &bitarray.CompactBitArray{ExtraBitsStored: 2, Elems: []byte{0b1000_0000}},
				Sigs:     [][]byte{sig},
			},
			2, true,
		},
		{
			"bit array size disagrees with the key count",
			Multisignature{
				BitArray: &bitarray.CompactBitArray{ExtraBitsStored: 2, Elems: []byte{0b1000_0000}},
				Sigs:     [][]byte{sig},
			},
			3, false,
		},
		{
			"fewer signatures than set bits would run off the end of Sigs",
			Multisignature{
				BitArray: &bitarray.CompactBitArray{ExtraBitsStored: 2, Elems: []byte{0b1100_0000}},
				Sigs:     [][]byte{sig},
			},
			2, false,
		},
		{
			"more signatures than set bits leaves some never indexed",
			Multisignature{
				BitArray: &bitarray.CompactBitArray{ExtraBitsStored: 2, Elems: []byte{0b1000_0000}},
				Sigs:     [][]byte{sig, sig},
			},
			2, false,
		},
		{
			// The bit array's own shape, which the two clauses above take on
			// trust: nine bits claimed over one stored byte makes the walk index
			// Elems[1].
			"bit array claiming more bits than it stores",
			Multisignature{
				BitArray: &bitarray.CompactBitArray{ExtraBitsStored: 1, Elems: []byte{0x80}},
				Sigs:     [][]byte{sig},
			},
			9, false,
		},
		{
			"bit array with bits set past its end",
			Multisignature{
				BitArray: &bitarray.CompactBitArray{ExtraBitsStored: 2, Elems: []byte{0b1000_0001}},
				Sigs:     [][]byte{sig},
			},
			2, false,
		},
		{
			"nil bit array against a key that has constituents",
			Multisignature{},
			1, false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.mSig.ValidateBasic(tt.nKeys)
			if tt.wantOK {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// Every multisignature a signer assembles through the exported API validates
// against the key it was built for. This is the direction that matters for not
// rejecting real signers.
func TestMultisignatureValidateBasicAcceptsAssembledSignatures(t *testing.T) {
	t.Parallel()

	for n := 1; n <= 20; n++ {
		pubkeys := make([]crypto.PubKey, n)
		for i := range pubkeys {
			pubkeys[i] = PubKeyMultisigThreshold{K: uint(i + 1)}
		}
		for signers := 1; signers <= n; signers++ {
			mSig := NewMultisig(n)
			for i := range signers {
				require.NoError(t, mSig.AddSignatureFromPubKey([]byte("sig"), pubkeys[i], pubkeys))
			}
			require.NoError(t, mSig.ValidateBasic(n), "n=%d signers=%d", n, signers)
		}
	}
}
