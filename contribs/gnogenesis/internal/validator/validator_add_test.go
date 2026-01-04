package validator

import (
	"context"
	"testing"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Validator_Add(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis file", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := NewValidatorCmd(commands.NewTestIO())
		args := []string{
			"add",
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
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			"dummyaddress",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "invalid validator address")
	})

	t.Run("invalid voting power", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		key := common.DummyKey(t)

		// Create the command
		cmd := NewValidatorCmd(commands.NewTestIO())
		args := []string{
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			key.Address().String(),
			"--power",
			"-1", // invalid voting power
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errInvalidPower)
	})

	t.Run("invalid validator name", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		key := common.DummyKey(t)

		// Create the command
		cmd := NewValidatorCmd(commands.NewTestIO())
		args := []string{
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			key.Address().String(),
			"--name",
			"", // invalid validator name
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errInvalidName.Error())
	})

	t.Run("invalid public key", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		key := common.DummyKey(t)

		// Create the command
		cmd := NewValidatorCmd(commands.NewTestIO())
		args := []string{
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			key.Address().String(),
			"--name",
			"example",
			"--pub-key",
			"invalidkey", // invalid pub key
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "invalid validator public key")
	})

	t.Run("public key address mismatch", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		dummyKeys := common.DummyKeys(t, 2)

		// Create the command
		cmd := NewValidatorCmd(commands.NewTestIO())
		args := []string{
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			dummyKeys[0].Address().String(),
			"--name",
			"example",
			"--pub-key",
			crypto.PubKeyToBech32(dummyKeys[1]), // another key
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errPublicKeyAddressMismatch.Error())
	})

	t.Run("validator with same address exists", func(t *testing.T) {
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
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			dummyKeys[0].Address().String(),
			"--name",
			"example",
			"--pub-key",
			crypto.PubKeyToBech32(dummyKeys[0]), // another key
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errAddressPresent.Error())
	})

	t.Run("valid genesis validator", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		key := common.DummyKey(t)
		genesis := common.DefaultGenesis()

		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := NewValidatorCmd(commands.NewTestIO())
		args := []string{
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			key.Address().String(),
			"--name",
			"example",
			"--pub-key",
			crypto.PubKeyToBech32(key), // another key
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)
	})
}
