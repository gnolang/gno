package multisig

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig/bitarray"
)

// Multisignature is used to represent the signature object used in the multisigs.
// Sigs is a list of signatures, sorted by corresponding index.
type Multisignature struct {
	BitArray *bitarray.CompactBitArray
	Sigs     [][]byte
}

// NewMultisig returns a new Multisignature of size n.
func NewMultisig(n int) *Multisignature {
	// Default the signature list to have a capacity of two, since we can
	// expect that most multisigs will require multiple signers.
	return &Multisignature{bitarray.NewCompactBitArray(n), make([][]byte, 0, 2)}
}

// GetIndex returns the index of pk in keys. Returns -1 if not found
func getIndex(pk crypto.PubKey, keys []crypto.PubKey) int {
	for i := range keys {
		if pk.Equals(keys[i]) {
			return i
		}
	}
	return -1
}

// AddSignature adds a signature to the multisig, at the corresponding index.
// If the signature already exists, replace it.
//
// It reports an error rather than silently producing an unverifiable
// multisignature when index is not one of mSig's own slots. Verification
// requires exactly one signature per set bit — see Multisignature.ValidateBasic
// — and an index at or past the bit array's size sets no bit, so appending for
// it would break that equality and the result would be rejected on chain with
// nothing here having complained.
func (mSig *Multisignature) AddSignature(sig []byte, index int) error {
	if index < 0 || index >= mSig.BitArray.Size() {
		return fmt.Errorf(
			"multisignature: index %d out of range for a %d-bit multisignature", index, mSig.BitArray.Size())
	}
	newSigIndex := mSig.BitArray.NumTrueBitsBefore(index)
	// Signature already exists, just replace the value there
	if mSig.BitArray.GetIndex(index) {
		mSig.Sigs[newSigIndex] = sig
		return nil
	}
	mSig.BitArray.SetIndex(index, true)
	// Optimization if the index is the greatest index
	if newSigIndex == len(mSig.Sigs) {
		mSig.Sigs = append(mSig.Sigs, sig)
		return nil
	}
	// Expand slice by one with a dummy element, move all elements after i
	// over by one, then place the new signature in that gap.
	mSig.Sigs = append(mSig.Sigs, make([]byte, 0))
	copy(mSig.Sigs[newSigIndex+1:], mSig.Sigs[newSigIndex:])
	mSig.Sigs[newSigIndex] = sig
	return nil
}

// AddSignatureFromPubKey adds a signature to the multisig, at the index in
// keys corresponding to the provided pubkey.
func (mSig *Multisignature) AddSignatureFromPubKey(sig []byte, pubkey crypto.PubKey, keys []crypto.PubKey) error {
	index := getIndex(pubkey, keys)
	if index == -1 {
		keysStr := make([]string, len(keys))
		for i, k := range keys {
			keysStr[i] = fmt.Sprintf("%X", k.Bytes())
		}

		return fmt.Errorf("provided key %X doesn't exist in pubkeys: \n%s", pubkey.Bytes(), strings.Join(keysStr, "\n"))
	}

	// keys and mSig are supplied by the caller from separate places, so a key
	// found in keys is not necessarily a slot mSig has.
	return mSig.AddSignature(sig, index)
}

// ValidateBasic returns an error unless mSig is a well-formed multisignature for
// a threshold key over nKeys constituents.
//
// Both places that walk a multisignature call this before trusting its shape:
// PubKeyMultisigThreshold.VerifyBytes and auth's signature gas consumer, which
// recurses over the constituent keys itself and so reaches them before
// VerifyBytes is called. Sharing one implementation is the point — the two must
// not disagree about which shapes are walkable, and the equality below is a
// consensus rule.
func (mSig Multisignature) ValidateBasic(nKeys int) error {
	// The bit array is as caller-chosen as the rest of the signature, and it
	// decides both how far the walk runs and which signatures it indexes, so its
	// own shape has to hold before Size() is trusted for either: an
	// ExtraBitsStored past the end of Elems would make the walk index off the end
	// of it.
	if err := mSig.BitArray.ValidateBasic(); err != nil {
		return err
	}
	size := mSig.BitArray.Size()
	if size != nKeys {
		return fmt.Errorf(
			"multisignature bit array size %d does not match the %d constituent keys", size, nKeys)
	}
	// Exactly one signature per set bit. The walk advances one signature per set
	// bit, so fewer signatures than set bits runs off the end of Sigs, and more
	// leaves trailing signatures that nothing ever indexes.
	if nSigs := mSig.BitArray.NumTrueBitsBefore(size); nSigs != len(mSig.Sigs) {
		return fmt.Errorf(
			"multisignature has %d set bits but %d signatures", nSigs, len(mSig.Sigs))
	}
	return nil
}

// Marshal the multisignature with amino
func (mSig *Multisignature) Marshal() []byte {
	return amino.MustMarshal(mSig)
}
