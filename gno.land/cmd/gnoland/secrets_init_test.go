package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func verifyValidatorKey(t *testing.T, path string) {
	t.Helper()

	validatorKey, err := readSecretData[privval.FilePVKey](path)
	require.NoError(t, err)

	assert.NoError(t, validateValidatorKey(validatorKey))
}

func verifyValidatorState(t *testing.T, path string) {
	t.Helper()

	validatorState, err := readSecretData[privval.FilePVLastSignState](path)
	require.NoError(t, err)

	assert.Zero(t, validatorState.Height)
	assert.Zero(t, validatorState.Round)
	assert.Zero(t, validatorState.Step)
	assert.Nil(t, validatorState.Signature)
	assert.Nil(t, validatorState.SignBytes)
}

func verifyNodeKey(t *testing.T, path string) {
	t.Helper()

	nodeKey, err := readSecretData[p2p.NodeKey](path)
	require.NoError(t, err)

	assert.NoError(t, validateNodeKey(nodeKey))
}

func TestSecrets_Init_All(t *testing.T) {
	t.Parallel()

	t.Run("invalid data directory", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"init",
			"--data-dir",
			"",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errInvalidDataDir.Error())
	})

	t.Run("all secrets initialized", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		tempDir := t.TempDir()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"init",
			"--data-dir",
			tempDir,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Verify the validator key is saved
		verifyValidatorKey(t, filepath.Join(tempDir, defaultValidatorKeyName))

		// Verify the last sign validator state is saved
		verifyValidatorState(t, filepath.Join(tempDir, defaultValidatorStateName))

		// Verify the node p2p key is saved
		verifyNodeKey(t, filepath.Join(tempDir, defaultNodeKeyName))
	})

	t.Run("no secrets overwritten", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		tempDir := t.TempDir()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"init",
			"--data-dir",
			tempDir,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Verify the validator key is saved
		verifyValidatorKey(t, filepath.Join(tempDir, defaultValidatorKeyName))

		// Verify the last sign validator state is saved
		verifyValidatorState(t, filepath.Join(tempDir, defaultValidatorStateName))

		// Verify the node p2p key is saved
		verifyNodeKey(t, filepath.Join(tempDir, defaultNodeKeyName))

		// Attempt to reinitialize the secrets, without the overwrite permission
		cmdErr = cmd.ParseAndRun(context.Background(), args)
		require.ErrorIs(t, cmdErr, errOverwriteNotEnabled)
	})
}

func TestSecrets_Init_Single(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name         string
		keyValue     string
		expectedFile string
		verifyFn     func(*testing.T, string)
	}{
		{
			"validator key initialized",
			validatorPrivateKeyKey,
			defaultValidatorKeyName,
			verifyValidatorKey,
		},
		{
			"validator state initialized",
			validatorStateKey,
			defaultValidatorStateName,
			verifyValidatorState,
		},
		{
			"node p2p key initialized",
			nodeKeyKey,
			defaultNodeKeyName,
			verifyNodeKey,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				tempDir = t.TempDir()

				expectedPath = filepath.Join(tempDir, testCase.expectedFile)
			)

			// Create the command
			cmd := newRootCmd(commands.NewTestIO())
			args := []string{
				"secrets",
				"init",
				"--data-dir",
				tempDir,
				testCase.keyValue,
			}

			// Run the command
			cmdErr := cmd.ParseAndRun(context.Background(), args)
			require.NoError(t, cmdErr)

			// Verify the validator key is saved
			testCase.verifyFn(t, expectedPath)
		})
	}
}

func TestSecrets_Init_Single_Overwrite(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name         string
		keyValue     string
		expectedFile string
	}{
		{
			"validator key not overwritten",
			validatorPrivateKeyKey,
			defaultValidatorKeyName,
		},
		{
			"validator state not overwritten",
			validatorStateKey,
			defaultValidatorStateName,
		},
		{
			"node p2p key not overwritten",
			nodeKeyKey,
			defaultNodeKeyName,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()

			// Create the command
			cmd := newRootCmd(commands.NewTestIO())
			args := []string{
				"secrets",
				"init",
				"--data-dir",
				tempDir,
				testCase.keyValue,
			}

			// Run the command
			cmdErr := cmd.ParseAndRun(context.Background(), args)
			require.NoError(t, cmdErr)

			// Attempt to reinitialize the secret, without the overwrite permission
			cmdErr = cmd.ParseAndRun(context.Background(), args)
			require.ErrorIs(t, cmdErr, errOverwriteNotEnabled)
		})
	}
}
