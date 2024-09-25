package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Download(t *testing.T) {
	t.Parallel()

	t.Run("no chain ID specified", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"download",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errNoChainID.Error())
	})

	t.Run("invalid chain ID specified", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"download",
			"random-testnet",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errChainNotSupported.Error())
	})

	t.Run("valid genesis.json download", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		require.NoError(t, os.MkdirAll(tempDir, 0o755))

		genesisPath := fmt.Sprintf("%s/genesis.json", t.TempDir())

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"download",
			test4ID,
			"--genesis-path",
			genesisPath,
		}

		// Run the command
		require.NoError(t, cmd.ParseAndRun(context.Background(), args))

		// Make sure the genesis.json is present
		sha, err := computeSHA256(genesisPath)
		require.NoError(t, err)

		assert.EqualValues(t, genesisSHAMap[test4ID], sha)
	})
}
