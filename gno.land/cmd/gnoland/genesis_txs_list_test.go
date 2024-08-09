package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func TestGenesis_List_All(t *testing.T) {
	t.Parallel()

	t.Run("list all txs", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory
		homeDir := newTestHomeDirectory(t, t.TempDir())

		// Generate dummy txs
		txs := generateDummyTxs(t, 10)

		genesis := getDefaultGenesis()
		genesis.AppState = gnoland.GnoGenesisState{
			Txs: txs,
		}
		require.NoError(t, genesis.SaveAs(homeDir.GenesisFilePath()))

		cio := commands.NewTestIO()
		buf := bytes.NewBuffer(nil)
		cio.SetOut(commands.WriteNopCloser(buf))

		cmd := newRootCmd(cio)
		args := []string{
			"genesis",
			"txs",
			"list",
			"--home",
			homeDir.Path(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		require.Len(t, buf.String(), 4442)
	})
}
