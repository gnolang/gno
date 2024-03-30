package ed25519

// Verify returns true if the signature is valid for the message and public key.
func VerifyED25519(publicKey []byte, message []byte, signature []byte) bool {
	return verifyED25519(publicKey, message, signature)
}

func verifyED25519(publicKey []byte, message []byte, signature []byte) bool // injected
