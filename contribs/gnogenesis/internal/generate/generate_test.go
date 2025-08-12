package generate

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Generate(t *testing.T) {
	t.Parallel()

	t.Run("default genesis", func(t *testing.T) {
		t.Parallel()

		tempDir, cleanup := testutils.NewTestCaseDir(t)
		t.Cleanup(cleanup)

		genesisPath := filepath.Join(tempDir, "genesis.json")

		// Create the command
		cmd := NewGenerateCmd(commands.NewTestIO())
		args := []string{
			"--output-path",
			genesisPath,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Load the genesis
		genesis, readErr := types.GenesisDocFromFile(genesisPath)
		require.NoError(t, readErr)

		// Make sure the default configuration is set
		defaultGenesis := common.DefaultGenesis()
		defaultGenesis.GenesisTime = genesis.GenesisTime

		assert.Equal(t, defaultGenesis, genesis)
	})

	t.Run("set chain ID", func(t *testing.T) {
		t.Parallel()

		chainID := "example-chain-ID"

		tempDir, cleanup := testutils.NewTestCaseDir(t)
		t.Cleanup(cleanup)

		genesisPath := filepath.Join(tempDir, "genesis.json")

		// Create the command
		cmd := NewGenerateCmd(commands.NewTestIO())
		args := []string{
			"--chain-id",
			chainID,
			"--output-path",
			genesisPath,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Load the genesis
		genesis, readErr := types.GenesisDocFromFile(genesisPath)
		require.NoError(t, readErr)

		assert.Equal(t, genesis.ChainID, chainID)
	})

	t.Run("set block max tx bytes", func(t *testing.T) {
		t.Parallel()

		blockMaxTxBytes := int64(100)

		tempDir, cleanup := testutils.NewTestCaseDir(t)
		t.Cleanup(cleanup)

		genesisPath := filepath.Join(tempDir, "genesis.json")

		// Create the command
		cmd := NewGenerateCmd(commands.NewTestIO())
		args := []string{
			"--block-max-tx-bytes",
			fmt.Sprintf("%d", blockMaxTxBytes),
			"--output-path",
			genesisPath,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Load the genesis
		genesis, readErr := types.GenesisDocFromFile(genesisPath)
		require.NoError(t, readErr)

		assert.Equal(
			t,
			genesis.ConsensusParams.Block.MaxTxBytes,
			blockMaxTxBytes,
		)
	})

	t.Run("set block max data bytes", func(t *testing.T) {
		t.Parallel()

		blockMaxDataBytes := int64(100)

		tempDir, cleanup := testutils.NewTestCaseDir(t)
		t.Cleanup(cleanup)

		genesisPath := filepath.Join(tempDir, "genesis.json")

		// Create the command
		cmd := NewGenerateCmd(commands.NewTestIO())
		args := []string{
			"--block-max-data-bytes",
			fmt.Sprintf("%d", blockMaxDataBytes),
			"--output-path",
			genesisPath,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Load the genesis
		genesis, readErr := types.GenesisDocFromFile(genesisPath)
		require.NoError(t, readErr)

		assert.Equal(
			t,
			genesis.ConsensusParams.Block.MaxDataBytes,
			blockMaxDataBytes,
		)
	})

	t.Run("set block max gas", func(t *testing.T) {
		t.Parallel()

		blockMaxGas := int64(100)

		tempDir, cleanup := testutils.NewTestCaseDir(t)
		t.Cleanup(cleanup)

		genesisPath := filepath.Join(tempDir, "genesis.json")

		// Create the command
		cmd := NewGenerateCmd(commands.NewTestIO())
		args := []string{
			"--block-max-gas",
			fmt.Sprintf("%d", blockMaxGas),
			"--output-path",
			genesisPath,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Load the genesis
		genesis, readErr := types.GenesisDocFromFile(genesisPath)
		require.NoError(t, readErr)

		assert.Equal(
			t,
			genesis.ConsensusParams.Block.MaxGas,
			blockMaxGas,
		)
	})

	t.Run("set block time iota", func(t *testing.T) {
		t.Parallel()

		blockTimeIota := int64(10)

		tempDir, cleanup := testutils.NewTestCaseDir(t)
		t.Cleanup(cleanup)

		genesisPath := filepath.Join(tempDir, "genesis.json")

		// Create the command
		cmd := NewGenerateCmd(commands.NewTestIO())
		args := []string{
			"--block-time-iota",
			fmt.Sprintf("%d", blockTimeIota),
			"--output-path",
			genesisPath,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Load the genesis
		genesis, readErr := types.GenesisDocFromFile(genesisPath)
		require.NoError(t, readErr)

		assert.Equal(
			t,
			genesis.ConsensusParams.Block.TimeIotaMS,
			blockTimeIota,
		)
	})

	t.Run("invalid genesis config (chain ID)", func(t *testing.T) {
		t.Parallel()

		invalidChainID := "thischainidisunusuallylongsoitwillcausethetesttofail"

		tempDir, cleanup := testutils.NewTestCaseDir(t)
		t.Cleanup(cleanup)

		genesisPath := filepath.Join(tempDir, "genesis.json")

		// Create the command
		cmd := NewGenerateCmd(commands.NewTestIO())
		args := []string{
			"--chain-id",
			invalidChainID,
			"--output-path",
			genesisPath,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.Error(t, cmdErr)
	})
}
