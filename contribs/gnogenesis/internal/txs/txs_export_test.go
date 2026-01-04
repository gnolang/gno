package txs

import (
	"bufio"
	"context"
	"testing"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Txs_Export(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis file", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := NewTxsCmd(commands.NewTestIO())
		args := []string{
			"export",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, common.ErrUnableToLoadGenesis.Error())
	})

	t.Run("invalid genesis app state", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		genesis.AppState = nil // no app state
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := NewTxsCmd(commands.NewTestIO())
		args := []string{
			"export",
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errAppStateNotSet.Error())
	})

	t.Run("no output file specified", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		genesis.AppState = gnoland.GnoGenesisState{
			Txs: generateDummyTxs(t, 1),
		}
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := NewTxsCmd(commands.NewTestIO())
		args := []string{
			"export",
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, common.ErrNoOutputFile.Error())
	})

	t.Run("valid txs export", func(t *testing.T) {
		t.Parallel()

		// Generate dummy txs
		txs := generateDummyTxs(t, 10)

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		genesis.AppState = gnoland.GnoGenesisState{
			Txs: txs,
		}
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Prepare the output file
		outputFile, outputCleanup := testutils.NewTestFile(t)
		t.Cleanup(outputCleanup)

		// Create the command
		cmd := NewTxsCmd(commands.NewTestIO())
		args := []string{
			"export",
			"--genesis-path",
			tempGenesis.Name(),
			outputFile.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Validate the transactions were written down
		scanner := bufio.NewScanner(outputFile)

		outputTxs := make([]gnoland.TxWithMetadata, 0)
		for scanner.Scan() {
			var tx gnoland.TxWithMetadata

			require.NoError(t, amino.UnmarshalJSON(scanner.Bytes(), &tx))

			outputTxs = append(outputTxs, tx)
		}

		require.NoError(t, scanner.Err())

		assert.Len(t, outputTxs, len(txs))

		for index, tx := range outputTxs {
			assert.Equal(t, txs[index], tx)
		}
	})
}
