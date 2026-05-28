package secp256k1

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

func X_verify(publicKey []byte, message []byte, signature []byte) bool {
	if len(publicKey) != 33 {
		panic(fmt.Sprintf(
			"invalid public key length: expected 33, got %d",
			len(publicKey)))
	}
	pub := secp256k1.PubKeySecp256k1{}
	copy(pub[:], publicKey)
	return pub.VerifyBytes(message, signature)
}
