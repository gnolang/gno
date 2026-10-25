package ed25519

import (
	"crypto/ed25519"

	gnoed25519 "github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

func X_verify(publicKey []byte, message []byte, signature []byte) bool {
	if len(publicKey) != gnoed25519.PubKeyEd25519Size {
		return ed25519.Verify(publicKey, message, signature) // panics on the length, as before
	}
	// A small-order public key verifies forged signatures (see
	// PubKeyEd25519.IsSmallOrder). Realms call this native with a
	// caller-supplied key and often treat a passing signature as proof that
	// the caller holds that key, so it goes through the chain's VerifyBytes
	// and its gate.
	return gnoed25519.PubKeyEd25519(publicKey).VerifyBytes(message, signature)
}
