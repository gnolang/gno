package secp256k1

// RecoverPubkey returns the public key that made signature over digest, as
// Ethereum's ecrecover does.
//
// digest must be 32 bytes. It is not hashed again: hash the message being
// authorized yourself (for Ethereum, with Keccak-256) and pass that hash. Never
// take the digest from untrusted input: for a digest of their choosing, anyone
// can build a signature that recovers any public key they like (a zero digest
// is the simplest case), so a matching key would prove nothing.
//
// signature must be 65 bytes: r and s as 32-byte big-endian integers, then a
// recovery ID of 0 or 1. Ethereum's v of 27/28 and EIP-155 values are not
// accepted; convert them first.
//
// The key is returned uncompressed: 0x04 || X || Y. The signer's Ethereum
// address is the last 20 bytes of the Keccak-256 hash of pubkey[1:].
//
// A wrong digest or recovery ID still recovers a key, only not the signer's,
// so the caller must compare the result with the key or address it expects.
// High-S signatures are accepted, as ecrecover accepts them: (r, n-s) with the
// other recovery ID recovers the same key, so a signature must not serve as a
// replay key.
//
// RecoverPubkey returns a zero key and false when it recovers no key: for a
// digest or signature of the wrong length, r or s outside [1, n-1], a recovery
// ID other than 0 or 1, an r that is not the x-coordinate of a curve point, or
// a signature that recovers the point at infinity.
func RecoverPubkey(digest []byte, signature []byte) (pubkey [65]byte, ok bool) {
	// The length check and the fixed-size copies keep the native argument
	// conversion bounded: the native never sees caller-sized slices.
	if len(digest) != 32 || len(signature) != 65 {
		return pubkey, false
	}
	var d [32]byte
	var sig [65]byte
	copy(d[:], digest)
	copy(sig[:], signature)
	return recoverPubkey(d, sig)
}

func recoverPubkey(digest [32]byte, signature [65]byte) ([65]byte, bool) // injected
