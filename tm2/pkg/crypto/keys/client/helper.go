package client

import (
	"crypto/sha256"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
)

// GenerateMnemonic generates a new BIP39 mnemonic using the
// provided entropy size
func GenerateMnemonic(entropySize int) (string, error) {
	// Generate the entropy seed
	entropySeed, err := bip39.NewEntropy(entropySize)
	if err != nil {
		return "", err
	}

	// Generate the actual mnemonic
	return bip39.NewMnemonic(entropySeed[:])
}

// GenerateMnemonicWithCustomEntropy generates a BIP39 mnemonic using
// user-provided entropy instead of computer PRNG
func GenerateMnemonicWithCustomEntropy(io commands.IO) (string, error) {
	// Display entropy advice
	io.Println("")
	io.Println("=== MANUAL ENTROPY GENERATION ===")
	io.Println("")
	io.Println("Provide at least 160 bits of entropy from a true random source:")
	io.Println("- Dice: 38+ d20 rolls (e.g., 18 7 3 12 5 19 8 2 14 11...)")
	io.Println("- Coins: 160+ flips (e.g., HTTHHTTHHHTTHHTHTTHHTHHT...)")
	io.Println("- Cards: 31+ draws (e.g., 7H 2C KS 9D 4H JS QC 3S...)")
	io.Println("- Other: keyboard mashing, environmental noise, etc.")
	io.Println("")

	// Get the entropy input
	inputEntropy, err := io.GetString("Enter your entropy (any length, will be hashed with SHA-256):")
	if err != nil {
		return "", err
	}

	if len(inputEntropy) < 27 {
		return "", fmt.Errorf("entropy too short (%d characters). Please provide at least 27 characters for 160-bit security", len(inputEntropy))
	}

	// Hash the input entropy to create deterministic seed
	hashedEntropy := sha256.Sum256([]byte(inputEntropy))
	
	// Show what we're using as entropy (first 16 bytes as hex)
	io.Printf("\nDerived entropy (SHA-256): %x...\n", hashedEntropy[:16])
	io.Printf("Input length: %d characters\n", len(inputEntropy))
	
	// Confirm before proceeding
	conf, err := io.GetConfirmation("Generate mnemonic from this entropy?")
	if err != nil {
		return "", err
	}
	if !conf {
		return "", fmt.Errorf("mnemonic generation cancelled")
	}

	// Generate the mnemonic from the hashed entropy
	return bip39.NewMnemonic(hashedEntropy[:])
}
