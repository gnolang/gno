package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecrets_Verify_All(t *testing.T) {
	t.Parallel()

	t.Run("signature mismatch", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir(), withSecrets)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())

		// Run the init command
		initArgs := []string{
			"secrets",
			"init",
			"--home",
			homeDir.Path(),
		}

		// Run the init command
		require.NoError(t, cmd.ParseAndRun(context.Background(), initArgs))

		// Modify the signature
		state, err := readSecretData[privval.FilePVLastSignState](homeDir.SecretsValidatorState())
		require.NoError(t, err)

		state.SignBytes = []byte("something totally random")
		state.Signature = []byte("signature")

		require.NoError(t, saveSecretData(state, homeDir.SecretsValidatorState()))

		cmd = newRootCmd(commands.NewTestIO())

		// Run the verify command
		verifyArgs := []string{
			"secrets",
			"verify",
			"--home",
			homeDir.Path(),
		}

		assert.ErrorContains(
			t,
			cmd.ParseAndRun(context.Background(), verifyArgs),
			errSignatureMismatch.Error(),
		)
	})

	t.Run("all secrets valid", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		tempDir := t.TempDir()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())

		// Run the init command
		initArgs := []string{
			"secrets",
			"init",
			"--home",
			tempDir,
		}

		// Run the init command
		require.NoError(t, cmd.ParseAndRun(context.Background(), initArgs))

		cmd = newRootCmd(commands.NewTestIO())

		// Run the verify command
		verifyArgs := []string{
			"secrets",
			"verify",
			"--home",
			tempDir,
		}

		assert.NoError(t, cmd.ParseAndRun(context.Background(), verifyArgs))
	})
}

func TestSecrets_Verify_All_Missing(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name                 string
		fileName             string
		expectedErrorMessage string
	}{
		{
			"invalid validator key path",
			defaultValidatorKeyName,
			"unable to read validator key",
		},
		{
			"invalid validator state path",
			defaultValidatorStateName,
			"unable to read last validator sign state",
		},
		{
			"invalid node p2p key path",
			defaultNodeKeyName,
			"unable to read node p2p key",
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create a temporary directory
			homeDir := newTestHomeDirectory(t, t.TempDir())

			// Create the command
			cmd := newRootCmd(commands.NewTestIO())

			// Run the init command
			initArgs := []string{
				"secrets",
				"init",
				"--home",
				homeDir.Path(),
			}

			// Run the init command
			require.NoError(t, cmd.ParseAndRun(context.Background(), initArgs))

			// Delete the validator key
			require.NoError(t, os.Remove(filepath.Join(homeDir.SecretsDir(), testCase.fileName)))

			cmd = newRootCmd(commands.NewTestIO())

			// Run the verify command
			verifyArgs := []string{
				"secrets",
				"verify",
				"--home",
				homeDir.Path(),
			}

			assert.ErrorContains(
				t,
				cmd.ParseAndRun(context.Background(), verifyArgs),
				testCase.expectedErrorMessage,
			)
		})
	}

	t.Run("invalid validator key path", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir())

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())

		// Run the init command
		initArgs := []string{
			"secrets",
			"init",
			"--home",
			homeDir.Path(),
		}

		// Run the init command
		require.NoError(t, cmd.ParseAndRun(context.Background(), initArgs))

		// Delete the validator key
		require.NoError(t, os.Remove(homeDir.SecretsValidatorKey()))

		cmd = newRootCmd(commands.NewTestIO())

		// Run the verify command
		verifyArgs := []string{
			"secrets",
			"verify",
			"--home",
			homeDir.Path(),
		}

		assert.ErrorContains(
			t,
			cmd.ParseAndRun(context.Background(), verifyArgs),
			"unable to read validator key",
		)
	})
}

func TestSecrets_Verify_Single(t *testing.T) {
	t.Parallel()

	t.Run("invalid validator key", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir(), withSecrets)

		invalidKey := &privval.FilePVKey{
			PrivKey: nil, // invalid
		}

		require.NoError(t, saveSecretData(invalidKey, homeDir.SecretsValidatorKey()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"verify",
			"--home",
			homeDir.Path(),
			validatorPrivateKeyKey,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errInvalidPrivateKey)
	})

	t.Run("invalid validator state", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir(), withSecrets)

		invalidState := &privval.FilePVLastSignState{
			Height: -1, // invalid
		}

		require.NoError(t, saveSecretData(invalidState, homeDir.SecretsValidatorState()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"verify",
			"--home",
			homeDir.Path(),
			validatorStateKey,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errInvalidSignStateHeight)
	})

	t.Run("invalid validator state signature", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir(), withSecrets)

		validKey := generateValidatorPrivateKey()
		validState := generateLastSignValidatorState()

		// Save an invalid signature
		validState.Signature = []byte("totally valid signature")

		require.NoError(t, saveSecretData(validKey, homeDir.SecretsValidatorKey()))
		require.NoError(t, saveSecretData(validState, homeDir.SecretsValidatorState()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"verify",
			"--home",
			homeDir.Path(),
			validatorStateKey,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errSignatureValuesMissing)
	})

	t.Run("invalid node key", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir(), withSecrets)

		invalidNodeKey := &p2p.NodeKey{
			PrivKey: nil, // invalid
		}

		require.NoError(t, saveSecretData(invalidNodeKey, homeDir.SecretsNodeKey()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"verify",
			"--home",
			homeDir.Path(),
			nodeIDKey,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errInvalidNodeKey)
	})
}
