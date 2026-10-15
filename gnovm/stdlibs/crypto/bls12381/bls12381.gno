// Package bls12381 exposes BLS signature verification over the BLS12-381
// curve as a native binding, with the parameters Ethereum's consensus layer
// uses: the IETF proof-of-possession ciphersuite
// (BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_) with public keys in G1 and
// signatures in G2, both in compressed (ZCash) encoding.
package bls12381

const (
	// PubKeySize is the length of a compressed G1 public key.
	PubKeySize = 48
	// SignatureSize is the length of a compressed G2 signature.
	SignatureSize = 96
)

// FastAggregateVerify reports whether sig is a valid aggregate signature of
// msg by all of pubkeys, as specified by FastAggregateVerify in the IETF BLS
// signature draft. pubkeys is the concatenation of at least one PubKeySize
// public key. Malformed keys or signatures, or points outside the prime-order
// subgroup, make it return false.
func FastAggregateVerify(pubkeys, msg, sig []byte) bool {
	// Reject wrong lengths before the native call so a malformed input is
	// not copied into Go memory first; see adr/native_input_bounds.md.
	if len(pubkeys) == 0 || len(pubkeys)%PubKeySize != 0 || len(sig) != SignatureSize {
		return false
	}
	return fastAggregateVerify(pubkeys, msg, sig)
}

func fastAggregateVerify(pubkeys, msg, sig []byte) bool // injected
