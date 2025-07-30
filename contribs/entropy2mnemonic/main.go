package main

import (
	"bufio"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/tyler-smith/go-bip39"
)

const (
	// MinEntropyLength is the minimum recommended entropy length for 160-bit security
	MinEntropyLength = 27

	// RecommendedEntropyLength is the recommended entropy length for good security
	RecommendedEntropyLength = 43
)

func main() {
	if err := run(os.Stdin, os.Stdout, os.Stderr, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(stdin io.Reader, stdout, stderr io.Writer, args []string) error {
	fs := flag.NewFlagSet("entropy2mnemonic", flag.ContinueOnError)
	fs.SetOutput(stderr)

	quiet := fs.Bool("quiet", false, "only output the mnemonic")

	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: entropy2mnemonic [options] [entropy]\n\n")
		fmt.Fprintf(stderr, "Generate BIP39 mnemonic from custom entropy.\n\n")
		fmt.Fprintf(stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(stderr, "\nExamples:\n")
		fmt.Fprintf(stderr, "  entropy2mnemonic \"dice rolls: 20 5 13 8 19 2 11 15 4 17 9 1 14 6 18 3 12 7 16 10\"\n")
		fmt.Fprintf(stderr, "  echo \"my entropy\" | entropy2mnemonic\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Get entropy from args or stdin
	var input string
	if fs.NArg() > 0 {
		input = strings.Join(fs.Args(), " ")
	} else {
		if !*quiet {
			fmt.Fprintln(stdout, "=== ENTROPY TO MNEMONIC CONVERTER ===")
			fmt.Fprintln(stdout)
			fmt.Fprintln(stdout, "This tool generates a BIP39 mnemonic from your custom entropy.")
			fmt.Fprintln(stdout, "The same entropy will always produce the same mnemonic.")
			fmt.Fprintln(stdout)
			fmt.Fprintln(stdout, "REQUIREMENTS:")
			fmt.Fprintf(stdout, "- Minimum %d characters for 160-bit security\n", MinEntropyLength)
			fmt.Fprintf(stdout, "- Recommended %d+ characters for better security\n", RecommendedEntropyLength)
			fmt.Fprintln(stdout)
			fmt.Fprintln(stdout, "GOOD ENTROPY SOURCES:")
			fmt.Fprintln(stdout, "- Dice rolls: 38+ d20 rolls (e.g., 18 7 3 12 5 19 8 2 14 11...)")
			fmt.Fprintln(stdout, "- Coin flips: 160+ flips (e.g., HTTHHTTHHHTTHHTHTTHHTHHT...)")
			fmt.Fprintln(stdout, "- Playing cards: 31+ draws (e.g., 7H 2C KS 9D 4H JS QC 3S...)")
			fmt.Fprintln(stdout, "- Random typing, environmental noise, etc.")
			fmt.Fprintln(stdout)
			fmt.Fprintln(stdout, "Enter your entropy (press Enter when done):")
		}

		scanner := bufio.NewScanner(stdin)
		if scanner.Scan() {
			input = scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return fmt.Errorf("no entropy provided")
	}

	// Validate entropy length
	if len(input) < MinEntropyLength {
		return fmt.Errorf("entropy too short (%d characters). Please provide at least %d characters for 160-bit security",
			len(input), MinEntropyLength)
	}

	// Hash the input entropy
	hashedEntropy := sha256.Sum256([]byte(input))

	// Generate the mnemonic from the hashed entropy
	mnemonic, err := bip39.NewMnemonic(hashedEntropy[:])
	if err != nil {
		return fmt.Errorf("generating mnemonic: %w", err)
	}

	if *quiet {
		fmt.Fprintln(stdout, mnemonic)
	} else {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Entropy received:")
		fmt.Fprintf(stdout, "  Length: %d characters\n", len(input))
		fmt.Fprintf(stdout, "  SHA-256: %x\n", hashedEntropy)
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Generated mnemonic (24 words):")
		fmt.Fprintln(stdout, mnemonic)
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "IMPORTANT: Store this mnemonic securely. It cannot be recovered!")
	}

	return nil
}
