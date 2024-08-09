package main

import (
	"context"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Init(t *testing.T) {
	t.Parallel()

	t.Run("default config initialized", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir())

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"config",
			"init",
			"--home",
			homeDir.Path(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Verify the config is valid
		cfg, err := config.LoadConfigFile(homeDir.ConfigFile())
		require.NoError(t, err)

		assert.NoError(t, cfg.ValidateBasic())
		assert.Equal(t, cfg, config.DefaultConfig())
	})

	t.Run("unable to overwrite config", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir())

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"config",
			"init",
			"--home",
			homeDir.Path(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Verify the config is valid
		cfg, err := config.LoadConfigFile(homeDir.ConfigFile())
		require.NoError(t, err)

		assert.NoError(t, cfg.ValidateBasic())
		assert.Equal(t, cfg, config.DefaultConfig())

		// Try to initialize again, expecting failure
		cmd = newRootCmd(commands.NewTestIO())

		cmdErr = cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errOverwriteNotEnabled)
	})

	t.Run("config overwritten", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir())

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"config",
			"init",
			"--force",
			"--home",
			homeDir.Path(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Verify the config is valid
		cfg, err := config.LoadConfigFile(homeDir.ConfigFile())
		require.NoError(t, err)

		assert.NoError(t, cfg.ValidateBasic())
		assert.Equal(t, cfg, config.DefaultConfig())

		// Try to initialize again, expecting success
		cmd = newRootCmd(commands.NewTestIO())

		cmdErr = cmd.ParseAndRun(context.Background(), args)
		assert.NoError(t, cmdErr)
	})
}
