package multisig

import (
	"errors"
	"slices"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// PubKeyMultisigThreshold implements a K of N threshold multisig.
type PubKeyMultisigThreshold struct {
	K       uint            `json:"threshold"`
	PubKeys []crypto.PubKey `json:"pubkeys"`
}

var _ crypto.PubKey = PubKeyMultisigThreshold{}

// NewPubKeyMultisigThreshold returns a new PubKeyMultisigThreshold.
// Panics if len(pubkeys) < k or 0 >= k.
func NewPubKeyMultisigThreshold(k int, pubkeys []crypto.PubKey) crypto.PubKey {
	if k <= 0 {
		panic("threshold k of n multisignature: k <= 0")
	}
	pk := PubKeyMultisigThreshold{uint(k), pubkeys}
	if err := pk.validate(); err != nil {
		panic(err.Error())
	}
	return pk
}

// validate returns an error if pk does not satisfy the invariants of a
// threshold multisig key.
//
// NewPubKeyMultisigThreshold is not the only way a PubKeyMultisigThreshold
// comes into existence: amino decodes one straight from the bytes of an
// untrusted, attacker-chosen transaction signature, populating K and PubKeys
// without going through the constructor. Anything using those fields to make
// an authorization decision must check them first.
func (pk PubKeyMultisigThreshold) validate() error {
	if pk.K == 0 {
		return errors.New("threshold k of n multisignature: k <= 0")
	}
	if uint(len(pk.PubKeys)) < pk.K {
		return errors.New("threshold k of n multisignature: len(pubkeys) < k")
	}
	for i, pubkey := range pk.PubKeys {
		if pubkey == nil {
			return errors.New("nil pubkey")
		}
		// A key listed twice occupies two positions of the threshold but only
		// ever needs the one private key behind it, so K distinct signers is
		// not what such a key actually requires.
		if slices.ContainsFunc(pk.PubKeys[:i], pubkey.Equals) {
			return errors.New("threshold k of n multisignature: duplicate pubkey")
		}
	}
	return nil
}

func (pk PubKeyMultisigThreshold) String() string {
	pubKeys := make([]string, len(pk.PubKeys))
	for i, key := range pk.PubKeys {
		pubKeys[i] = key.String()
	}

	return "[" + strings.Join(pubKeys, ", ") + "]"
}

// VerifyBytes expects sig to be an amino encoded version of a MultiSignature.
// Returns true iff the multisignature contains k or more signatures
// for the correct corresponding keys,
// and all signatures are valid. (Not just k of the signatures)
// The multisig uses a bitarray, so multiple signatures for the same key is not
// a concern.
func (pk PubKeyMultisigThreshold) VerifyBytes(msg []byte, marshalledSig []byte) bool {
	// pk is decoded from the signature of an untrusted transaction, so it has
	// not necessarily been through NewPubKeyMultisigThreshold. Without this,
	// a K of 0 verifies against no signatures at all.
	if pk.validate() != nil {
		return false
	}
	var sig Multisignature
	err := amino.Unmarshal(marshalledSig, &sig)
	if err != nil {
		return false
	}
	size := sig.BitArray.Size()
	// ensure bit array is the correct size
	if len(pk.PubKeys) != size {
		return false
	}
	// ensure size of signature list
	if len(sig.Sigs) < int(pk.K) || len(sig.Sigs) > size {
		return false
	}
	// ensure at least k signatures are set
	if sig.BitArray.NumTrueBitsBefore(size) < int(pk.K) {
		return false
	}
	// index in the list of signatures which we are concerned with.
	sigIndex := 0
	for i := range size {
		if sig.BitArray.GetIndex(i) {
			if !pk.PubKeys[i].VerifyBytes(msg, sig.Sigs[sigIndex]) {
				return false
			}
			sigIndex++
		}
	}
	return true
}

// Bytes returns the amino encoded version of the PubKeyMultisigThreshold
func (pk PubKeyMultisigThreshold) Bytes() []byte {
	return amino.MustMarshalAny(pk)
}

// Address returns tmhash(PubKeyMultisigThreshold.Bytes())
func (pk PubKeyMultisigThreshold) Address() crypto.Address {
	return crypto.AddressFromPreimage(pk.Bytes())
}

// Equals returns true iff pk and other both have the same number of keys, and
// all constituent keys are the same, and in the same order.
func (pk PubKeyMultisigThreshold) Equals(other crypto.PubKey) bool {
	otherKey, sameType := other.(PubKeyMultisigThreshold)
	if !sameType {
		return false
	}
	if pk.K != otherKey.K || len(pk.PubKeys) != len(otherKey.PubKeys) {
		return false
	}
	for i := range pk.PubKeys {
		if !pk.PubKeys[i].Equals(otherKey.PubKeys[i]) {
			return false
		}
	}
	return true
}
