package secp256k1

import (
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// Header byte of a compact signature that recovers an uncompressed key; the
// recovery ID is added to it.
const compactUncompressedHeader = 27

// X_recoverPubkey is the native binding for recoverPubkey. RecoverPubkey has
// already checked the lengths, and the fixed-size array parameters keep the
// dispatcher's conversion bounded.
//
// It calls decred's secp256k1 directly, so every build runs the same code
// whatever the build tags.
func X_recoverPubkey(digest [32]byte, signature [65]byte) (pubkey [65]byte, ok bool) {
	recoveryID := signature[64]
	// ecrecover accepts recovery IDs 0 and 1 only. RecoverCompact would also
	// take 2 and 3 (an x-coordinate of r+n) and the compressed-key headers.
	if recoveryID > 1 {
		return pubkey, false
	}
	var compact [65]byte
	compact[0] = compactUncompressedHeader + recoveryID
	copy(compact[1:], signature[:64])
	// RecoverCompact checks that r and s are in [1, n-1], that r is the
	// x-coordinate of a curve point and that the key is not the point at
	// infinity.
	key, _, err := ecdsa.RecoverCompact(compact[:], digest[:])
	if err != nil {
		return pubkey, false
	}
	copy(pubkey[:], key.SerializeUncompressed())
	return pubkey, true
}
