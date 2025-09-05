package client

import (
	"crypto/sha256"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
)

const (
	// MinEntropyChars is the minimum number of characters required for custom entropy
	// Calculated as: ceil(180 / log2(10+24)) = 36
	// This ensures at least 180 bits of entropy when using alphanumeric characters
	MinEntropyChars = 36
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
func GenerateMnemonicWithCustomEntropy(io commands.IO, masked bool) (string, error) {
	// Display entropy advice
	io.Println("")
	io.Println("=== MANUAL ENTROPY GENERATION ===")
	io.Println("")
	io.Println("Generate true random entropy using ONE of these methods:")
	io.Println("• Dice: Roll a D20 (20-sided die) exactly 42 times")
	io.Println("• Cards: Shuffle a standard 52-card deck 20 times, then record the full deck order")
	io.Println("")

	// Get the entropy input
	var inputEntropy string
	var err error

	prompt := "Enter your entropy:"
	if masked {
		inputEntropy, err = io.GetPassword(prompt, false)
	} else {
		inputEntropy, err = io.GetString(prompt)
	}
	if err != nil {
		return "", err
	}

	if len(inputEntropy) < MinEntropyChars {
		return "", fmt.Errorf("entropy too short (%d characters). Please provide at least %d characters", len(inputEntropy), MinEntropyChars)
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
