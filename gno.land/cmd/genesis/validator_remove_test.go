package main

import (
	"context"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Validator_Remove(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis file", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"validator",
			"remove",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "unable to load genesis")
	})

	t.Run("invalid validator address", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"validator",
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

		dummyKeys := []keys.Info{
			getDummyKey(t),
			getDummyKey(t),
		}

		genesis := getDefaultGenesis()

		// Set an existing validator
		genesis.Validators = append(genesis.Validators, types.GenesisValidator{
			Address: dummyKeys[0].GetAddress(),
			PubKey:  dummyKeys[0].GetPubKey(),
			Power:   1,
			Name:    "example",
		})

		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"validator",
			"remove",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			dummyKeys[1].GetPubKey().Address().String(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errValidatorNotPresent.Error())
	})

	t.Run("validator removed", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		dummyKey := getDummyKey(t)

		genesis := getDefaultGenesis()

		// Set an existing validator
		genesis.Validators = append(genesis.Validators, types.GenesisValidator{
			Address: dummyKey.GetAddress(),
			PubKey:  dummyKey.GetPubKey(),
			Power:   1,
			Name:    "example",
		})

		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"validator",
			"remove",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			dummyKey.GetPubKey().Address().String(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.NoError(t, cmdErr)
	})
}
