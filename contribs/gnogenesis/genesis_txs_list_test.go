package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/gnolang/contribs/gnogenesis/internal/balances"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/testutils"
)

func TestGenesis_List_All(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis path", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newGenesisCmd(commands.NewTestIO())
		args := []string{
			"txs",
			"list",
			"--genesis-path",
			"",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, balances.errUnableToLoadGenesis)
	})

	t.Run("list all txs", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		// Generate dummy txs
		txs := generateDummyTxs(t, 10)

		genesis := GetDefaultGenesis()
		genesis.AppState = gnoland.GnoGenesisState{
			Txs: txs,
		}
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		cio := commands.NewTestIO()
		buf := bytes.NewBuffer(nil)
		cio.SetOut(commands.WriteNopCloser(buf))

		cmd := newGenesisCmd(cio)
		args := []string{
			"txs",
			"list",
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		require.Len(t, buf.String(), 4442)
	})
}
