package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	t.Parallel()

	verifyDataCommon := func(nodeDir string) {
		// Verify the config is valid
		cfg, err := config.LoadConfigFile(constructConfigPath(nodeDir))
		require.NoError(t, err)

		assert.NoError(t, cfg.ValidateBasic())
		assert.Equal(t, cfg, config.DefaultConfig())

		// Verify the validator key is saved
		verifyValidatorKey(t, filepath.Join(constructSecretsPath(nodeDir), defaultValidatorKeyName))

		// Verify the last sign validator state is saved
		verifyValidatorState(t, filepath.Join(constructSecretsPath(nodeDir), defaultValidatorStateName))

		// Verify the node p2p key is saved
		verifyNodeKey(t, filepath.Join(constructSecretsPath(nodeDir), defaultNodeKeyName))
	}

	t.Run("config and secrets initialized", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		tempDir := t.TempDir()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"init",
			"--data-dir",
			tempDir,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Verify the generated data
		verifyDataCommon(tempDir)
	})

	t.Run("config and secrets not overwritten", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		tempDir := t.TempDir()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"init",
			"--data-dir",
			tempDir,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Verify the generated data
		verifyDataCommon(tempDir)

		// Try to run the command again, expecting failure
		cmdErr = cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errOverwriteNotEnabled)
	})

	t.Run("config and secrets overwritten", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		tempDir := t.TempDir()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"init",
			"--force",
			"--data-dir",
			tempDir,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Verify the generated data
		verifyDataCommon(tempDir)

		// Try to run the command again, expecting success
		cmdErr = cmd.ParseAndRun(context.Background(), args)
		assert.NoError(t, cmdErr)
	})
}
