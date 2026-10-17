package bls12381

import (
	bls "github.com/consensys/gnark-crypto/ecc/bls12-381"
)

const (
	pubKeySize    = 48
	signatureSize = 96
)

// ethDST is the domain separation tag of the IETF BLS proof-of-possession
// ciphersuite used by the Ethereum consensus layer.
var ethDST = []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")

// negG1Gen is -G1, the generator negated once so that the verification
// equation e(pk, H(m)) == e(G1, sig) becomes a single pairing product check.
var negG1Gen bls.G1Affine

func init() {
	_, _, g1, _ := bls.Generators()
	negG1Gen.Neg(&g1)
}

// X_fastAggregateVerify implements FastAggregateVerify of the IETF BLS
// signature draft (POP ciphersuite, minimal-pubkey-size: public keys in G1,
// signatures in G2), as used by Ethereum sync committees.
//
// pubkeys is the concatenation of n >= 1 compressed 48-byte G1 public keys;
// sig is a compressed 96-byte G2 signature. Every point is decoded with its
// subgroup check. It returns false on any malformed input.
func X_fastAggregateVerify(pubkeys, msg, sig []byte) bool {
	if len(pubkeys) == 0 || len(pubkeys)%pubKeySize != 0 || len(sig) != signatureSize {
		return false
	}

	var agg bls.G1Jac
	for i := 0; i < len(pubkeys); i += pubKeySize {
		var pk bls.G1Affine
		if _, err := pk.SetBytes(pubkeys[i : i+pubKeySize]); err != nil {
			return false
		}
		// KeyValidate: the identity is not a valid public key.
		if pk.IsInfinity() {
			return false
		}
		agg.AddMixed(&pk)
	}
	var aggAff bls.G1Affine
	aggAff.FromJacobian(&agg)
	return verifyAggregate(aggAff, msg, sig)
}

// X_fastAggregateVerifyExcluding verifies sig over msg by the signers of a
// committee whose aggregate public key is aggregate, given the public keys of
// the members that did not sign: the signers' aggregate is aggregate minus
// excluded. It equals FastAggregateVerify over the signers' keys when
// aggregate is the sum of the committee's keys, as in an Ethereum
// SyncCommittee, at the cost of decoding only the non-signers' keys.
//
// aggregate is a compressed 48-byte G1 key, excluded the concatenation of
// zero or more such keys, sig a compressed 96-byte G2 signature. Every point
// is decoded with its subgroup check, and an aggregate of no signer fails.
func X_fastAggregateVerifyExcluding(aggregate, excluded, msg, sig []byte) bool {
	if len(aggregate) != pubKeySize || len(excluded)%pubKeySize != 0 || len(sig) != signatureSize {
		return false
	}
	var agg bls.G1Affine
	if _, err := agg.SetBytes(aggregate); err != nil || agg.IsInfinity() {
		return false
	}
	var sum bls.G1Jac
	sum.FromAffine(&agg)
	for i := 0; i < len(excluded); i += pubKeySize {
		var pk bls.G1Affine
		if _, err := pk.SetBytes(excluded[i : i+pubKeySize]); err != nil || pk.IsInfinity() {
			return false
		}
		var neg bls.G1Affine
		neg.Neg(&pk)
		sum.AddMixed(&neg)
	}
	var signers bls.G1Affine
	signers.FromJacobian(&sum)
	if signers.IsInfinity() {
		return false
	}
	return verifyAggregate(signers, msg, sig)
}

// verifyAggregate checks e(pk, H(msg)) == e(G1, sig).
func verifyAggregate(pk bls.G1Affine, msg, sig []byte) bool {
	var s bls.G2Affine
	if _, err := s.SetBytes(sig); err != nil {
		return false
	}
	h, err := bls.HashToG2(msg, ethDST)
	if err != nil {
		return false
	}
	ok, err := bls.PairingCheck(
		[]bls.G1Affine{pk, negG1Gen},
		[]bls.G2Affine{h, s},
	)
	return err == nil && ok
}
