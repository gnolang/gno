// Package secp256k1 implements signature verification over the secp256k1
// elliptic curve, as used by Bitcoin, Ethereum and the gno.land chain.
package secp256k1

// PubKeySize is the length, in bytes, of a compressed secp256k1 public key
// accepted by [Verify].
const PubKeySize = 33

// SignatureSize is the length, in bytes, of a secp256k1 signature accepted by
// [Verify]. Signatures are encoded as the concatenation of the R and S
// scalars (32 bytes each), in lower-S form.
const SignatureSize = 64

// Verify reports whether sig is a valid signature of msg by the given
// compressed public key. The signature must be in R || S form (64 bytes) and
// lower-S; the message is hashed with SHA-256 before verification, mirroring
// the convention used by gno.land's transaction signing.
//
// Verify returns false for any malformed input (wrong-length key or
// signature, invalid encoding) rather than panicking.
func Verify(publicKey, message, signature []byte) bool {
	if len(publicKey) != PubKeySize || len(signature) != SignatureSize {
		return false
	}
	return verify(publicKey, message, signature)
}

func verify(publicKey, message, signature []byte) bool // injected
