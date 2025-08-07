package balances

import (
	"context"
	"testing"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Balances_Remove(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis", func(t *testing.T) {
		// Create the command
		cmd := NewBalancesCmd(commands.NewTestIO())
		args := []string{
			"remove",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.ErrorContains(t, cmdErr, common.ErrUnableToLoadGenesis.Error())
	})

	t.Run("genesis app state not set", func(t *testing.T) {
		t.Parallel()

		dummyKey := common.DummyKey(t)

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		genesis.AppState = nil // not set
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := NewBalancesCmd(commands.NewTestIO())
		args := []string{
			"remove",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			dummyKey.Address().String(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.ErrorContains(t, cmdErr, common.ErrAppStateNotSet.Error())
	})

	t.Run("address is present", func(t *testing.T) {
		t.Parallel()

		dummyKey := common.DummyKey(t)

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		state := gnoland.GnoGenesisState{
			// Set an initial balance value
			Balances: []gnoland.Balance{
				{
					Address: dummyKey.Address(),
					Amount:  std.NewCoins(std.NewCoin(ugnot.Denom, 100)),
				},
			},
		}
		genesis.AppState = state
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := NewBalancesCmd(commands.NewTestIO())
		args := []string{
			"remove",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			dummyKey.Address().String(),
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

		assert.Len(t, state.Balances, 0)
	})

	t.Run("address not present", func(t *testing.T) {
		t.Parallel()

		dummyKey := common.DummyKey(t)

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		state := gnoland.GnoGenesisState{
			Balances: []gnoland.Balance{}, // Empty initial balance
		}
		genesis.AppState = state
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := NewBalancesCmd(commands.NewTestIO())
		args := []string{
			"remove",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			dummyKey.Address().String(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.ErrorContains(t, cmdErr, errBalanceNotFound.Error())
	})
}
