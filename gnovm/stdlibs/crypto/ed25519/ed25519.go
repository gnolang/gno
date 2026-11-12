package ed25519

import (
	"crypto/ed25519"
	"reflect"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	gnoed25519 "github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

// X_verify reports whether signature is valid for message under publicKey.
func X_verify(m *gno.Machine, publicKey gno.TypedValue, message []byte, signature gno.TypedValue) bool {
	// Only the message is priced, so the key and the signature are copied
	// only at the one length each can have.
	publicKey.DeepFill(m.Store)
	signature.DeepFill(m.Store)
	if publicKey.GetLength() != ed25519.PublicKeySize || signature.GetLength() != ed25519.SignatureSize {
		return false
	}
	var pk, sig []byte
	gno.Gno2GoValue(&publicKey, reflect.ValueOf(&pk).Elem())
	gno.Gno2GoValue(&signature, reflect.ValueOf(&sig).Elem())
	// A small-order public key verifies forged signatures (see
	// PubKeyEd25519.IsSmallOrder). Realms call this native with a
	// caller-supplied key and often treat a passing signature as proof that
	// the caller holds that key, so it goes through the chain's VerifyBytes
	// and its gate.
	return gnoed25519.PubKeyEd25519(pk).VerifyBytes(message, sig)
}
