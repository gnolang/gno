package main

import (
	"context"
	"path/filepath"
	"testing"

	signer "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	fstate "github.com/gnolang/gno/tm2/pkg/bft/privval/state"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func verifyValidatorKey(t *testing.T, path string) {
	t.Helper()

	_, err := signer.LoadFileKey(path)
	require.NoError(t, err)
}

func verifyValidatorState(t *testing.T, path string) {
	t.Helper()

	validatorState, err := fstate.LoadFileState(path)
	require.NoError(t, err)

	assert.Zero(t, validatorState.Height)
	assert.Zero(t, validatorState.Round)
	assert.Zero(t, validatorState.Step)
	assert.Nil(t, validatorState.Signature)
	assert.Nil(t, validatorState.SignBytes)
}

func verifyNodeKey(t *testing.T, path string) {
	t.Helper()

	_, err := types.LoadNodeKey(path)
	require.NoError(t, err)
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

	t.Run("validator key type honored", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()

		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"init",
			"--data-dir",
			tempDir,
			"--validator-key-type",
			string(signer.KeyTypeEd25519),
		}

		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		fk, err := signer.LoadFileKey(filepath.Join(tempDir, defaultValidatorKeyName))
		require.NoError(t, err)
		assert.IsType(t, ed25519.PubKeyEd25519{}, fk.PubKey)
	})
}

// TestSecrets_Init_InvalidKeyType checks that a rejected -validator-key-type
// writes no secret at all, on the all-keys path ("") and on each single-key
// path: a partial write would make the corrected re-run fail on the overwrite
// guard.
func TestSecrets_Init_InvalidKeyType(t *testing.T) {
	t.Parallel()

	for _, key := range []string{"", validatorPrivateKeyKey, validatorStateKey, nodeIDKey} {
		name := key
		if name == "" {
			name = "all"
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()

			cmd := newRootCmd(commands.NewTestIO())
			args := []string{
				"secrets",
				"init",
				"--data-dir",
				tempDir,
				"--validator-key-type",
				"rsa",
			}
			if key != "" {
				args = append(args, key)
			}

			cmdErr := cmd.ParseAndRun(context.Background(), args)
			require.ErrorContains(t, cmdErr, "unknown validator key type")

			assert.NoFileExists(t, filepath.Join(tempDir, defaultValidatorKeyName))
			assert.NoFileExists(t, filepath.Join(tempDir, defaultValidatorStateName))
			assert.NoFileExists(t, filepath.Join(tempDir, defaultNodeKeyName))
		})
	}
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
			nodeIDKey,
			defaultNodeKeyName,
			verifyNodeKey,
		},
	}

	for _, testCase := range testTable {
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
			nodeIDKey,
			defaultNodeKeyName,
		},
	}

	for _, testCase := range testTable {
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
