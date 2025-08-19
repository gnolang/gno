package validator

import (
	"context"
	"testing"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Validator_Remove(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis file", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := NewValidatorCmd(commands.NewTestIO())
		args := []string{
			"remove",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, common.ErrUnableToLoadGenesis.Error())
	})

	t.Run("invalid validator address", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := NewValidatorCmd(commands.NewTestIO())
		args := []string{
			"remove",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			"dummyaddress",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "invalid validator address")
	})

	t.Run("validator not found", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		dummyKeys := common.DummyKeys(t, 2)
		genesis := common.DefaultGenesis()

		// Set an existing validator
		genesis.Validators = append(genesis.Validators, types.GenesisValidator{
			Address: dummyKeys[0].Address(),
			PubKey:  dummyKeys[0],
			Power:   1,
			Name:    "example",
		})

		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := NewValidatorCmd(commands.NewTestIO())
		args := []string{
			"remove",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			dummyKeys[1].Address().String(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errValidatorNotPresent.Error())
	})

	t.Run("validator removed", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		dummyKey := common.DummyKey(t)

		genesis := common.DefaultGenesis()

		// Set an existing validator
		genesis.Validators = append(genesis.Validators, types.GenesisValidator{
			Address: dummyKey.Address(),
			PubKey:  dummyKey,
			Power:   1,
			Name:    "example",
		})

		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := NewValidatorCmd(commands.NewTestIO())
		args := []string{
			"remove",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			dummyKey.Address().String(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.NoError(t, cmdErr)
	})
}
