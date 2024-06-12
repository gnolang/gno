package main

import (
	"context"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Txs_Remove(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis file", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"txs",
			"remove",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errUnableToLoadGenesis.Error())
	})

	t.Run("invalid genesis app state", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := getDefaultGenesis()
		genesis.AppState = nil // no app state
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"txs",
			"remove",
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errAppStateNotSet.Error())
	})
	t.Run("no transaction hash specified", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		// Generate dummy txs
		txs := generateDummyTxs(t, 10)

		genesis := getDefaultGenesis()
		genesis.AppState = gnoland.GnoGenesisState{
			Txs: txs,
		}
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"txs",
			"remove",
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errNoTxHashSpecified.Error())
	})

	t.Run("transaction removed", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		// Generate dummy txs
		txs := generateDummyTxs(t, 10)

		genesis := getDefaultGenesis()
		genesis.AppState = gnoland.GnoGenesisState{
			Txs: txs,
		}
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		txHash, err := getTxHash(txs[0])
		require.NoError(t, err)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"txs",
			"remove",
			"--genesis-path",
			tempGenesis.Name(),
			txHash,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Validate the transaction was removed
		updatedGenesis, err := types.GenesisDocFromFile(tempGenesis.Name())
		require.NoError(t, err)
		require.NotNil(t, updatedGenesis.AppState)

		// Fetch the state
		state := updatedGenesis.AppState.(gnoland.GnoGenesisState)

		assert.Len(t, state.Txs, len(txs)-1)

		for _, tx := range state.Txs {
			genesisTxHash, err := getTxHash(tx)
			require.NoError(t, err)

			assert.NotEqual(t, txHash, genesisTxHash)
		}
	})
}
