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

	var s bls.G2Affine
	if _, err := s.SetBytes(sig); err != nil {
		return false
	}

	h, err := bls.HashToG2(msg, ethDST)
	if err != nil {
		return false
	}

	ok, err := bls.PairingCheck(
		[]bls.G1Affine{aggAff, negG1Gen},
		[]bls.G2Affine{h, s},
	)
	return err == nil && ok
}
