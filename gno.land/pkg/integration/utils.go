package integration

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/stretchr/testify/require"
)

// `unquote` takes a slice of strings, resulting from splitting a string block by spaces, and
// processes them. The function handles quoted phrases and escape characters within these strings.
func unquote(args []string) ([]string, error) {
	const quote = '"'

	parts := []string{}
	var inQuote bool

	var part strings.Builder
	for _, arg := range args {
		var escaped bool
		for _, c := range arg {
			if escaped {
				// If the character is meant to be escaped, it is processed with Unquote.
				// We use `Unquote` here for two main reasons:
				// 1. It will validate that the escape sequence is correct
				// 2. It converts the escaped string to its corresponding raw character.
				//    For example, "\\t" becomes '\t'.
				uc, err := strconv.Unquote(`"\` + string(c) + `"`)
				if err != nil {
					return nil, fmt.Errorf("unhandled escape sequence `\\%c`: %w", c, err)
				}

				part.WriteString(uc)
				escaped = false
				continue
			}

			// If we are inside a quoted string and encounter an escape character,
			// flag the next character as `escaped`
			if inQuote && c == '\\' {
				escaped = true
				continue
			}

			// Detect quote and toggle inQuote state
			if c == quote {
				inQuote = !inQuote
				continue
			}

			// Handle regular character
			part.WriteRune(c)
		}

		// If we're inside a quote, add a single space.
		// It reflects one or multiple spaces between args in the original string.
		if inQuote {
			part.WriteRune(' ')
			continue
		}

		// Finalize part, add to parts, and reset for next part
		parts = append(parts, part.String())
		part.Reset()
	}

	// Check if a quote is left open
	if inQuote {
		return nil, errors.New("unfinished quote")
	}

	return parts, nil
}

func buildGnoland(t *testing.T, rootdir string) string {
	t.Helper()

	bin := filepath.Join(t.TempDir(), "gnoland-test")

	t.Log("building gnoland integration binary...")

	// Build a fresh gno binary in a temp directory
	gnoArgsBuilder := []string{"build", "-o", bin}

	os.Executable()

	// Forward `-covermode` settings if set
	if coverMode := testing.CoverMode(); coverMode != "" {
		gnoArgsBuilder = append(gnoArgsBuilder,
			"-covermode", coverMode,
		)
	}

	// Append the path to the gno command source
	gnoArgsBuilder = append(gnoArgsBuilder, filepath.Join(rootdir,
		"gno.land", "pkg", "integration", "process"))

	t.Logf("build command: %s", strings.Join(gnoArgsBuilder, " "))

	cmd := exec.Command("go", gnoArgsBuilder...)

	var buff bytes.Buffer
	cmd.Stderr, cmd.Stdout = &buff, &buff
	defer buff.Reset()

	if err := cmd.Run(); err != nil {
		require.FailNowf(t, "unable to build binary", "%q\n%s",
			err.Error(), buff.String())
	}

	return bin
}

// GeneratePrivKeyFromMnemonic generates a crypto.PrivKey from a mnemonic.
func GeneratePrivKeyFromMnemonic(mnemonic, bip39Passphrase string, account, index uint32) (crypto.PrivKey, error) {
	// Generate Seed from Mnemonic
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to generate seed: %w", err)
	}

	// Derive Private Key
	coinType := crypto.CoinType // ensure this is set correctly in your context
	hdPath := hd.NewFundraiserParams(account, coinType, index)
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, hdPath.String())
	if err != nil {
		return nil, fmt.Errorf("failed to derive private key: %w", err)
	}

	// Convert to secp256k1 private key
	privKey := secp256k1.PrivKeySecp256k1(derivedPriv)
	return privKey, nil
}
