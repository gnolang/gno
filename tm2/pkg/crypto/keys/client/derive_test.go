package client

import (
	"bytes"
	"math"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_execDerive(t *testing.T) {
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

	t.Run("valid accounts generated", func(t *testing.T) {
		t.Parallel()

		// Generate a dummy mnemonic
		entropy, entropyErr := bip39.NewEntropy(mnemonicEntropySize)
		require.NoError(t, entropyErr)

		mnemonic, mnemonicErr := bip39.NewMnemonic(entropy)
		require.NoError(t, mnemonicErr)

		cfg := &deriveCfg{
			numAccounts:  1,
			accountIndex: 0,
			mnemonic:     mnemonic,
		}

		// Create a test IO so we can capture output
		mockOut := bytes.NewBufferString("")

		testIO := commands.NewTestIO()
		testIO.SetOut(commands.WriteNopCloser(mockOut))

		require.NoError(t, execDerive(cfg, testIO))

		// Grab the output
		deriveOutput := mockOut.String()

		// Verify the addresses are derived correctly
		expectedAccounts := generateAccounts(
			mnemonic,
			cfg.accountIndex,
			cfg.numAccounts,
		)

		for _, expectedAccount := range expectedAccounts {
			assert.Contains(t, deriveOutput, expectedAccount.String())
		}
	})
}
