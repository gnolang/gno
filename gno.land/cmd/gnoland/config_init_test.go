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

func TestConfig_Init(t *testing.T) {
	t.Parallel()

	t.Run("invalid output path", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"config",
			"init",
			"--config-path",
			"",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errInvalidConfigOutputPath.Error())
	})

	t.Run("default config initialized", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		tempDir := t.TempDir()
		path := filepath.Join(tempDir, "config.toml")

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"config",
			"init",
			"--config-path",
			path,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Verify the config is valid
		cfg, err := config.LoadConfigFile(path)
		require.NoError(t, err)

		assert.NoError(t, cfg.ValidateBasic())
		assert.Equal(t, cfg, config.DefaultConfig())
	})
}
