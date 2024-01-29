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

func TestSecrets_Init_All(t *testing.T) {
	t.Parallel()

	t.Run("invalid data directory", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"secrets",
			"init",
			"all",
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
			"all",
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
}

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
