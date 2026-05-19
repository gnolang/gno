package secp256k1

import (
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

// X_verify is the native binding for crypto/secp256k1.verify. Length checks
// are performed by the Gno wrapper; here we trust the inputs and delegate to
// the production verifier in tm2.
func X_verify(publicKey, message, signature []byte) bool {
	var pub secp256k1.PubKeySecp256k1
	copy(pub[:], publicKey)
	return pub.VerifyBytes(message, signature)
}
