package multisig

import (
	"errors"
	"fmt"
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

// MaxNestingDepth is how many levels of threshold keys a
// PubKeyMultisigThreshold may contain, counting itself. A key holding only
// non-multisig constituents is depth 1; a multisig of multisigs is depth 2.
//
// The bound exists because verifying a nested key costs one amino decode of
// the whole remaining signature blob per level, so an unbounded chain is
// quadratic in the size of a transaction that pays for it linearly. It is not
// bounded by auth's TxSigLimit either: std.CountSubKeys counts leaves, and a
// chain of 1-of-1 keys has exactly one however deep it runs.
//
// 6 is the deepest a key that branches at every level can nest while spending
// no more than TxSigLimit's 7 leaves. The deep shape that stays inside a leaf
// budget is the unbalanced one — each level holding one leaf and one subkey —
// and d such levels spend d+1 leaves and 2d+1 keys, so 7 leaves reach depth 6
// at 13 keys. That is the same shape MaxTotalKeys admits, which is why the two
// bounds agree here: neither is the binding one for a branching key.
//
// Anything deeper needs a level holding a single subkey, which delegates its
// whole threshold to that subkey and so expresses nothing the subkey did not
// already express. Those are reachable within MaxTotalKeys alone — a 1-of-1
// chain is one key per level — which is why the depth bound is a separate
// check and not a consequence of the key budget.
const MaxNestingDepth = 6

// MaxTotalKeys is how many keys a PubKeyMultisigThreshold may contain in
// total, counting itself, every threshold key nested within it, and every leaf
// key at the bottom.
//
// The bound exists because validate scans each level's constituents pairwise to
// reject duplicates, so an unbounded key is quadratic in the size of a
// transaction that pays for it linearly. TxSigLimit bounds the width no better
// than it bounds the depth, and for the same reason: std.CountSubKeys counts
// leaves, and a constituent holding no keys at all contributes none however
// many of them are listed.
//
// 14 is twice auth's default TxSigLimit of 7, which is the most leaves a
// transaction may present. A key that spends all 7 on leaves and branches at
// every level needs at most 6 threshold keys above them, so 14 admits every
// branching key TxSigLimit permits.
//
// It does not admit every degenerate one: 7 leaves each behind a 1-of-1 wrapper
// of its own is 15 keys, and is rejected here although it counts 7 against
// TxSigLimit. Wrapping a lone subkey expresses nothing that subkey did not
// already express, which is the same reasoning MaxNestingDepth rests on.
//
// It is deliberately the tighter of the two bounds and can be raised later:
// a chain configuring TxSigLimit above 7 has to raise this with it.
const MaxTotalKeys = 14

var _ crypto.PubKey = PubKeyMultisigThreshold{}

// NewPubKeyMultisigThreshold returns a new PubKeyMultisigThreshold.
// Panics unless k and pubkeys describe a valid threshold key; see
// NewPubKeyMultisigThresholdChecked for the fallible form.
func NewPubKeyMultisigThreshold(k int, pubkeys []crypto.PubKey) crypto.PubKey {
	pk, err := NewPubKeyMultisigThresholdChecked(k, pubkeys)
	if err != nil {
		panic(err.Error())
	}
	return pk
}

// NewPubKeyMultisigThresholdChecked is NewPubKeyMultisigThreshold, reporting an
// error rather than panicking.
//
// A caller taking k or pubkeys from user input should use this one. What makes
// a threshold key valid has grown over time — a positive threshold, at least k
// constituents, no duplicates among them, and the structural bounds above — and
// a caller that pre-checks those conditions itself, one at a time, drifts out of
// sync with validate silently: the next condition added here becomes a panic on
// ordinary input at a call site nobody thought to revisit.
func NewPubKeyMultisigThresholdChecked(k int, pubkeys []crypto.PubKey) (crypto.PubKey, error) {
	if k <= 0 {
		return nil, errors.New("threshold k of n multisignature: k <= 0")
	}
	pk := PubKeyMultisigThreshold{uint(k), pubkeys}
	if err := pk.validate(); err != nil {
		return nil, err
	}
	return pk, nil
}

// validate returns an error if pk does not satisfy the invariants of a
// threshold multisig key.
//
// NewPubKeyMultisigThreshold is not the only way a PubKeyMultisigThreshold
// comes into existence: amino decodes one straight from the bytes of an
// untrusted, caller-chosen transaction signature, populating K and PubKeys
// without going through the constructor. Anything using those fields to make
// an authorization decision must check them first.
func (pk PubKeyMultisigThreshold) validate() error {
	if pk.K == 0 {
		return errors.New("threshold k of n multisignature: k <= 0")
	}
	if uint(len(pk.PubKeys)) < pk.K {
		return errors.New("threshold k of n multisignature: len(pubkeys) < k")
	}
	// Bound the key before the pairwise duplicate scan below, which is
	// quadratic in the number of constituents at each level, and reject nils
	// before it, which is what stops Equals dereferencing one. ValidateStructure
	// owns both: it is the only check here that visits every node, so a
	// per-constituent nil check at this level would cover only this level while
	// Equals compares whole subtrees.
	if err := pk.ValidateStructure(); err != nil {
		return err
	}
	for i, pubkey := range pk.PubKeys {
		// A key listed twice occupies two positions of the threshold but only
		// ever needs the one private key behind it, so K distinct signers is
		// not what such a key actually requires.
		if slices.ContainsFunc(pk.PubKeys[:i], pubkey.Equals) {
			return errors.New("threshold k of n multisignature: duplicate pubkey")
		}
	}
	return nil
}

// ValidateStructure returns an error if pk nests threshold keys more than
// MaxNestingDepth levels deep, or holds more than MaxTotalKeys keys in total.
//
// It is separate from validate because auth's signature gas consumer recurses
// over the constituent keys itself, before VerifyBytes is ever reached, and so
// has to establish the same bounds on its own recursion. Both callers agreeing
// on one implementation is the point: these bounds are what keep that
// recursion, the amino decode it performs at each level, and validate's own
// duplicate scan linear in the transaction.
func (pk PubKeyMultisigThreshold) ValidateStructure() error {
	budget := MaxTotalKeys
	return validateStructure(pk, MaxNestingDepth, &budget)
}

// validateStructure returns an error unless pub fits within levels of nesting
// and budget further keys. Both budgets are spent on pub before its own
// constituents are looked at, so a key that exhausts either one is rejected
// without being walked to the bottom: the check never costs what it is meant to
// prevent.
func validateStructure(pub crypto.PubKey, levels int, budget *int) error {
	// Rejecting nils here, rather than only among validate's own constituents,
	// is what keeps validate's duplicate scan from dereferencing one: Equals
	// recurses into the constituents of the keys it compares, so a nil nested
	// below the top level is reached by pubkey.Equals rather than by the nil
	// check above it. This walk is the only place that visits every node.
	if pub == nil {
		return errors.New("threshold k of n multisignature: nil pubkey")
	}
	if *budget <= 0 {
		return fmt.Errorf(
			"threshold k of n multisignature: more than %d keys in total", MaxTotalKeys)
	}
	*budget--

	v, ok := pub.(PubKeyMultisigThreshold)
	if !ok {
		return nil
	}
	if levels <= 0 {
		return fmt.Errorf(
			"threshold k of n multisignature: nested more than %d levels deep", MaxNestingDepth)
	}
	for _, subkey := range v.PubKeys {
		if err := validateStructure(subkey, levels-1, budget); err != nil {
			return err
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
	// Establish the shape of the signature against this key before walking it:
	// the bit array size against the constituent count, and exactly one
	// signature per set bit. Shared with auth's gas consumer, which reaches the
	// same structure first. See Multisignature.ValidateBasic.
	if sig.ValidateBasic(len(pk.PubKeys)) != nil {
		return false
	}
	size := sig.BitArray.Size()
	// ensure at least k signatures are set. ValidateBasic has already equated the
	// set bits with len(Sigs), so this also subsumes a separate length check on
	// Sigs: nSigs is at most size by construction and at least K here.
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
