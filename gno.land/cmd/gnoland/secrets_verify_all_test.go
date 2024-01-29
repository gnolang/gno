package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"
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
			"all",
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
			"all",
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
			"all",
			"--data-dir",
			tempDir,
		}

		// Run the init command
		require.NoError(t, cmd.ParseAndRun(context.Background(), initArgs))

		// Modify the signature
		statePath := filepath.Join(tempDir, defaultValidatorStateName)
		state, err := readSecretData[privval.FilePVLastSignState](statePath)
		require.NoError(t, err)

		state.SignBytes = []byte("something totally random")
		state.Signature = []byte("signature")

		require.NoError(t, saveSecretData(state, statePath))

		cmd = newRootCmd(commands.NewTestIO())

		// Run the verify command
		verifyArgs := []string{
			"secrets",
			"verify",
			"all",
			"--data-dir",
			tempDir,
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
			"all",
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
			"all",
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
				"all",
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
				"all",
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
			"all",
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
			"all",
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
