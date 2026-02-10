package txs

import (
	"bytes"
	"context"
	"testing"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
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
		cmd := NewTxsCmd(commands.NewTestIO())
		args := []string{
			"list",
			"--genesis-path",
			"",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, common.ErrUnableToLoadGenesis)
	})

	t.Run("list all txs", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		// Generate dummy txs
		txs := generateDummyTxs(t, 10)

		genesis := common.DefaultGenesis()
		genesis.AppState = gnoland.GnoGenesisState{
			Txs: txs,
		}
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		cio := commands.NewTestIO()
		buf := bytes.NewBuffer(nil)
		cio.SetOut(commands.WriteNopCloser(buf))

		cmd := NewTxsCmd(cio)
		args := []string{
			"list",
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		require.Len(t, buf.String(), 5262)
	})
}
