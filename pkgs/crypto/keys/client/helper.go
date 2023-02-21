package client

import "github.com/gnolang/gno/pkgs/crypto/bip39"

// generateMnemonic generates a new BIP39 mnemonic using the
// provided entropy size
func generateMnemonic(entropySize int) (string, error) {
	// Generate the entropy seed
	entropySeed, err := bip39.NewEntropy(entropySize)
	if err != nil {
		return "", err
	}

	// Generate the actual mnemonic
	return bip39.NewMnemonic(entropySeed[:])
}
