package main

import (
	"context"
	"flag"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Txs_Generate(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"txs",
			"generate",
			"./",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.ErrorContains(t, cmdErr, errUnableToLoadGenesis.Error())
	})

	t.Run("missing args", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"txs",
			"generate",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.ErrorContains(t, cmdErr, flag.ErrHelp.Error())
	})

	t.Run("invalid dir", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"txs",
			"generate",
			"invalid",
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.ErrorContains(t, cmdErr, errUnableToLoadPackages.Error())
	})

	t.Run("packages from dir", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"txs",
			"generate",
			"../../../examples/gno.land/p/demo/avl",
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Validate the genesis was updated
		genesis, loadErr := types.GenesisDocFromFile(tempGenesis.Name())
		require.NoError(t, loadErr)

		require.NotNil(t, genesis.AppState)

		state, ok := genesis.AppState.(gnoland.GnoGenesisState)
		require.True(t, ok)

		require.Equal(t, 1, len(state.Txs))
		require.Equal(t, 1, len(state.Txs[0].Msgs))

		msgAddPkg, ok := state.Txs[0].Msgs[0].(vmm.MsgAddPackage)
		require.True(t, ok)

		assert.Equal(t, "gno.land/p/demo/avl", msgAddPkg.Package.Path)
	})
}
