package client

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestMnemonic generates a random mnemonic
func generateTestMnemonic(t *testing.T) string {
	t.Helper()

	entropy, entropyErr := bip39.NewEntropy(256)
	require.NoError(t, entropyErr)

	mnemonic, mnemonicErr := bip39.NewMnemonic(entropy)
	require.NoError(t, mnemonicErr)

	return mnemonic
}

func TestDerive_InvalidDerive(t *testing.T) {
	t.Parallel()

	t.Run("invalid number of accounts, no accounts requested", func(t *testing.T) {
		t.Parallel()

		cfg := &deriveCfg{
			numAccounts: 0,
		}

		assert.ErrorIs(t, execDerive(cfg, nil), errInvalidNumAccounts)
	})

	t.Run("invalid number of accounts, > uint32", func(t *testing.T) {
		t.Parallel()

		cfg := &deriveCfg{
			numAccounts: math.MaxUint32 + 1, // > uint32
		}

		assert.ErrorIs(t, execDerive(cfg, nil), errInvalidNumAccounts)
	})

	t.Run("invalid account index", func(t *testing.T) {
		t.Parallel()

		cfg := &deriveCfg{
			numAccounts:  1,
			accountIndex: math.MaxUint32 + 1, // > uint32
		}

		assert.ErrorIs(t, execDerive(cfg, nil), errInvalidAccountIndex)
	})

	t.Run("invalid mnemonic", func(t *testing.T) {
		t.Parallel()

		cfg := &deriveCfg{
			numAccounts:  1,
			accountIndex: 0,
			mnemonic:     "one two",
		}

		assert.ErrorIs(t, execDerive(cfg, nil), errInvalidMnemonic)
	})
}

func TestDerive_ValidDerive(t *testing.T) {
	t.Parallel()

	// Generate a dummy mnemonic
	var (
		mnemonic            = generateTestMnemonic(t)
		accountIndex uint64 = 0
		numAccounts  uint64 = 10
	)

	// Set up the IO
	mockOut := bytes.NewBufferString("")

	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(mockOut))

	// Create the root command
	cmd := newDeriveCmd(io)

	// Prepare the args
	args := []string{
		"derive",
		"--mnemonic",
		mnemonic,
		"--num-accounts",
		fmt.Sprintf("%d", numAccounts),
		"--account-index",
		fmt.Sprintf("%d", accountIndex),
	}

	// Run the command
	require.NoError(t, cmd.ParseAndRun(context.Background(), args))

	// Verify the addresses are derived correctly
	expectedAccounts := generateAccounts(
		mnemonic,
		accountIndex,
		numAccounts,
	)

	// Grab the output
	deriveOutput := mockOut.String()

	for _, expectedAccount := range expectedAccounts {
		assert.Contains(t, deriveOutput, expectedAccount.String())
	}
}
