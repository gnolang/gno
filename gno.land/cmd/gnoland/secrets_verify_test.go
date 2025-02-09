package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	signer "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	"github.com/gnolang/gno/tm2/pkg/bft/privval/state"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecrets_Verify_All(t *testing.T) {
	t.Parallel()

	t.Run("invalid data directory", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"verify",
			"--data-dir",
			"",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errInvalidDataDir.Error())
	})

	t.Run("invalid data directory", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "example.json")

		require.NoError(
			t,
			os.WriteFile(
				path,
				[]byte("hello"),
				0o644,
			),
		)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"verify",
			"--data-dir",
			path,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errInvalidDataDir.Error())
	})

	t.Run("signature mismatch", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		tempDir := t.TempDir()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())

		// Run the init command
		initArgs := []string{
			"secrets",
			"init",
			"--data-dir",
			tempDir,
		}

		// Run the init command
		require.NoError(t, cmd.ParseAndRun(context.Background(), initArgs))

		// Modify the signature
		statePath := filepath.Join(tempDir, defaultValidatorStateName)
		state, err := state.NewFileState(statePath)
		require.NoError(t, err)

		require.NoError(t, state.Update(
			state.Height,
			state.Round,
			state.Step,
			[]byte("something totally random"),
			[]byte("signature"),
		))

		cmd = newRootCmd(commands.NewTestIO())

		// Run the verify command
		verifyArgs := []string{
			"secrets",
			"verify",
			"--data-dir",
			tempDir,
		}

		assert.Error(t, cmd.ParseAndRun(context.Background(), verifyArgs))
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
			"--data-dir",
			tempDir,
		}

		// Run the init command
		require.NoError(t, cmd.ParseAndRun(context.Background(), initArgs))

		cmd = newRootCmd(commands.NewTestIO())

		// Run the verify command
		verifyArgs := []string{
			"secrets",
			"verify",
			"--data-dir",
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
			tempDir := t.TempDir()

			// Create the command
			cmd := newRootCmd(commands.NewTestIO())

			// Run the init command
			initArgs := []string{
				"secrets",
				"init",
				"--data-dir",
				tempDir,
			}

			// Run the init command
			require.NoError(t, cmd.ParseAndRun(context.Background(), initArgs))

			// Delete the validator key
			require.NoError(t, os.Remove(filepath.Join(tempDir, testCase.fileName)))

			cmd = newRootCmd(commands.NewTestIO())

			// Run the verify command
			verifyArgs := []string{
				"secrets",
				"verify",
				"--data-dir",
				tempDir,
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
		tempDir := t.TempDir()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())

		// Run the init command
		initArgs := []string{
			"secrets",
			"init",
			"--data-dir",
			tempDir,
		}

		// Run the init command
		require.NoError(t, cmd.ParseAndRun(context.Background(), initArgs))

		// Delete the validator key
		require.NoError(t, os.Remove(filepath.Join(tempDir, defaultValidatorKeyName)))

		cmd = newRootCmd(commands.NewTestIO())

		// Run the verify command
		verifyArgs := []string{
			"secrets",
			"verify",
			"--data-dir",
			tempDir,
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

	t.Run("invalid validator state signature", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		keyPath := filepath.Join(dirPath, defaultValidatorKeyName)
		statePath := filepath.Join(dirPath, defaultValidatorStateName)

		_, err := signer.GeneratePersistedFileKey(keyPath)
		require.NoError(t, err)
		validState, err := state.GeneratePersistedFileState(statePath)
		require.NoError(t, err)

		// Save an invalid signature
		err = validState.Update(
			validState.Height,
			validState.Round,
			validState.Step,
			validState.SignBytes,
			[]byte("totally invalid signature"),
		)
		require.NoError(t, err)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"verify",
			"--data-dir",
			dirPath,
			validatorStateKey,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.Error(t, cmdErr)
	})

	t.Run("invalid node key", func(t *testing.T) {
		t.Parallel()

		dirPath := t.TempDir()
		path := filepath.Join(dirPath, defaultNodeKeyName)

		invalidNodeKey := &types.NodeKey{
			PrivKey: nil, // invalid
		}

		require.NoError(t, saveNodeKey(invalidNodeKey, path))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"verify",
			"--data-dir",
			dirPath,
			nodeIDKey,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errInvalidNodeKey)
	})
}
