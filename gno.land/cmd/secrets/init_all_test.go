package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
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
			"init",
			"all",
			"--data-dir",
			tempDir,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Verify the validator key is saved
		validateValidatorKey(t, filepath.Join(tempDir, defaultValidatorKeyName))

		// Verify the last sign validator state is saved
		validateValidatorState(t, filepath.Join(tempDir, defaultValidatorStateName))

		// Verify the node p2p key is saved
		validateNodeKey(t, filepath.Join(tempDir, defaultNodeKeyName))
	})
}

func validateValidatorKey(t *testing.T, path string) {
	t.Helper()

	validatorKeyRaw, err := os.ReadFile(path)
	require.NoError(t, err)

	var validatorKey privval.FilePVKey
	require.NoError(t, amino.UnmarshalJSON(validatorKeyRaw, &validatorKey))

	assert.NotNil(t, validatorKey.Address)
	assert.NotEqual(t, types.Address{}, validatorKey.Address)
	assert.NotNil(t, validatorKey.PrivKey)
	assert.NotNil(t, validatorKey.PubKey)
}

func validateValidatorState(t *testing.T, path string) {
	t.Helper()

	validatorStateRaw, err := os.ReadFile(path)
	require.NoError(t, err)

	var validatorState privval.FilePVLastSignState
	require.NoError(t, amino.UnmarshalJSON(validatorStateRaw, &validatorState))

	assert.Zero(t, validatorState.Height)
	assert.Zero(t, validatorState.Round)
	assert.Zero(t, validatorState.Step)
	assert.Nil(t, validatorState.Signature)
	assert.Nil(t, validatorState.SignBytes)
}

func validateNodeKey(t *testing.T, path string) {
	t.Helper()

	nodeKeyRaw, err := os.ReadFile(path)
	require.NoError(t, err)

	var nodeKey p2p.NodeKey
	require.NoError(t, amino.UnmarshalJSON(nodeKeyRaw, &nodeKey))

	assert.NotNil(t, nodeKey.PrivKey)
}
