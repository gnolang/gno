package main

import (
	"context"
	"flag"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Packages_Del(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"packages",
			"del",
			"gno.land/p/demo/avl",
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
			"packages",
			"del",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.ErrorContains(t, cmdErr, flag.ErrHelp.Error())
	})

	t.Run("del genesis packages", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		dummyPackagePath := "gno.land/r/demo/dummy"

		genesis := getDefaultGenesis()
		state := gnoland.GnoGenesisState{
			// Set an initial addpkg tx
			Txs: []std.Tx{
				{
					Msgs: []std.Msg{
						vmm.MsgAddPackage{
							Creator: getDummyKey(t).Address(),
							Package: &std.MemPackage{
								Name: "dummy",
								Path: dummyPackagePath + "0",
							},
						},
					},
				},
				{
					Msgs: []std.Msg{
						vmm.MsgAddPackage{
							Creator: getDummyKey(t).Address(),
							Package: &std.MemPackage{
								Name: "dummy",
								Path: dummyPackagePath + "1",
							},
						},
					},
				},
			},
		}
		genesis.AppState = state
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))
		require.Equal(t, 2, len(state.Txs))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"packages",
			"del",
			dummyPackagePath + "0",
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
	})
}
