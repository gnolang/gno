package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Txs_Add_Packages(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis file", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"txs",
			"add",
			"packages",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errUnableToLoadGenesis.Error())
	})

	t.Run("invalid package dir", func(t *testing.T) {
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
			"add",
			"packages",
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errInvalidPackageDir.Error())
	})

	t.Run("valid package", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Prepare the package
		var (
			packagePath = "gno.land/p/demo/cuttlas"
			dir         = t.TempDir()
		)

		createFile := func(path, data string) {
			file, err := os.Create(path)
			require.NoError(t, err)

			_, err = file.WriteString(data)
			require.NoError(t, err)
		}

		// Create the gno.mod file
		createFile(
			filepath.Join(dir, "gno.mod"),
			fmt.Sprintf("module %s\n", packagePath),
		)

		// Create a simple main.gno
		createFile(
			filepath.Join(dir, "main.gno"),
			"package cuttlas\n\nfunc Example() string {\nreturn \"Manos arriba!\"\n}",
		)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"txs",
			"add",
			"packages",
			"--genesis-path",
			tempGenesis.Name(),
			dir,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Validate the transactions were written down
		updatedGenesis, err := types.GenesisDocFromFile(tempGenesis.Name())
		require.NoError(t, err)
		require.NotNil(t, updatedGenesis.AppState)

		// Fetch the state
		state := updatedGenesis.AppState.(gnoland.GnoGenesisState)

		require.Equal(t, 1, len(state.Txs))
		require.Equal(t, 1, len(state.Txs[0].Msgs))

		msgAddPkg, ok := state.Txs[0].Msgs[0].(vmm.MsgAddPackage)
		require.True(t, ok)

		assert.Equal(t, packagePath, msgAddPkg.Package.Path)
	})
}
