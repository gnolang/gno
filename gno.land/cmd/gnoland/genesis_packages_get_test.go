package main

import (
	"context"
	"flag"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Packages_Get(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"packages",
			"get",
			"", // TODO
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
			"get",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.ErrorContains(t, cmdErr, flag.ErrHelp.Error())
	})

	t.Run("get genesis packages", func(t *testing.T) {
		t.Parallel()

		dummyPackagePath := "gno.land/r/demo/dummy"

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

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
								Path: dummyPackagePath,
							},
						},
					},
				},
				{
					Msgs: []std.Msg{
						vmm.MsgAddPackage{
							Creator: getDummyKey(t).Address(),
							Package: &std.MemPackage{
								Name: "dummy1",
								Path: dummyPackagePath + "1",
							},
						},
					},
				},
				{
					Msgs: []std.Msg{
						vmm.MsgCall{
							PkgPath: dummyPackagePath,
						},
					},
				},
			},
		}
		genesis.AppState = state
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"packages",
			"get",
			dummyPackagePath,
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)
	})
}
