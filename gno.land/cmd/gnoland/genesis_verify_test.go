package main

import (
	"context"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/mock"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Verify(t *testing.T) {
	t.Parallel()

	getValidTestGenesis := func() *types.GenesisDoc {
		key := mock.GenPrivKey().PubKey()

		return &types.GenesisDoc{
			GenesisTime:     time.Now(),
			ChainID:         "valid-chain-id",
			ConsensusParams: types.DefaultConsensusParams(),
			Validators: []types.GenesisValidator{
				{
					Address: key.Address(),
					PubKey:  key,
					Power:   1,
					Name:    "valid validator",
				},
			},
		}
	}

	t.Run("invalid txs", func(t *testing.T) {
		t.Parallel()

		tempFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		g := getValidTestGenesis()

		g.AppState = gnoland.GnoGenesisState{
			Balances: []gnoland.Balance{},
			Txs: []std.Tx{
				{},
			},
		}

		require.NoError(t, g.SaveAs(tempFile.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"verify",
			"--genesis-path",
			tempFile.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.Error(t, cmdErr)
	})

	t.Run("invalid balances", func(t *testing.T) {
		t.Parallel()

		tempFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		g := getValidTestGenesis()

		g.AppState = gnoland.GnoGenesisState{
			Balances: []gnoland.Balance{
				{},
			},
			Txs: []std.Tx{},
		}

		require.NoError(t, g.SaveAs(tempFile.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"verify",
			"--genesis-path",
			tempFile.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.Error(t, cmdErr)
	})

	t.Run("valid genesis", func(t *testing.T) {
		t.Parallel()

		tempFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		g := getValidTestGenesis()
		g.AppState = gnoland.GnoGenesisState{
			Balances: []gnoland.Balance{},
			Txs:      []std.Tx{},
		}

		require.NoError(t, g.SaveAs(tempFile.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"verify",
			"--genesis-path",
			tempFile.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)
	})

	t.Run("valid genesis, no state", func(t *testing.T) {
		t.Parallel()

		tempFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		g := getValidTestGenesis()
		require.NoError(t, g.SaveAs(tempFile.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"verify",
			"--genesis-path",
			tempFile.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)
	})

	t.Run("invalid genesis state", func(t *testing.T) {
		t.Parallel()

		tempFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		g := getValidTestGenesis()
		g.AppState = "Totally invalid state"
		require.NoError(t, g.SaveAs(tempFile.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"verify",
			"--genesis-path",
			tempFile.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.Error(t, cmdErr)
	})
}
