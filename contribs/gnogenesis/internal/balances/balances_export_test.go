package balances

import (
	"bufio"
	"context"
	"testing"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getDummyBalances generates dummy balance lines
func getDummyBalances(t *testing.T, count int) []gnoland.Balance {
	t.Helper()

	dummyKeys := common.DummyKeys(t, count)
	amount := std.NewCoins(std.NewCoin(ugnot.Denom, 10))

	balances := make([]gnoland.Balance, len(dummyKeys))

	for index, key := range dummyKeys {
		balances[index] = gnoland.Balance{
			Address: key.Address(),
			Amount:  amount,
		}
	}

	return balances
}

func TestGenesis_Balances_Export(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis file", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := NewBalancesCmd(commands.NewTestIO())
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
		cmd := NewBalancesCmd(commands.NewTestIO())
		args := []string{
			"export",
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, common.ErrAppStateNotSet.Error())
	})

	t.Run("no output file specified", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		genesis.AppState = gnoland.GnoGenesisState{
			Balances: getDummyBalances(t, 1),
		}
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := NewBalancesCmd(commands.NewTestIO())
		args := []string{
			"export",
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, common.ErrNoOutputFile.Error())
	})

	t.Run("valid balances export", func(t *testing.T) {
		t.Parallel()

		// Generate dummy balances
		balances := getDummyBalances(t, 10)
		gnoland.SortBalances(balances)

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		genesis.AppState = gnoland.GnoGenesisState{
			Balances: balances,
		}
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Prepare the output file
		outputFile, outputCleanup := testutils.NewTestFile(t)
		t.Cleanup(outputCleanup)

		// Create the command
		cmd := NewBalancesCmd(commands.NewTestIO())
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

		outputBalances := make([]gnoland.Balance, 0)
		for scanner.Scan() {
			var balance gnoland.Balance
			err := balance.Parse(scanner.Text())
			require.NoError(t, err)

			outputBalances = append(outputBalances, balance)
		}

		require.NoError(t, scanner.Err())

		assert.Len(t, outputBalances, len(balances))

		for index, balance := range outputBalances {
			assert.Equal(t, balances[index], balance)
		}
	})
}
