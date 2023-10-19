package main

import (
	"context"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getDummyKey generates a random public key,
// and returns the key info
func getDummyKey(t *testing.T) keys.Info {
	t.Helper()

	mnemonic, err := client.GenerateMnemonic(256)
	require.NoError(t, err)

	kb := keys.NewInMemory()

	info, err := kb.CreateAccount(
		"dummy",
		mnemonic,
		"",
		"",
		uint32(0),
		uint32(0),
	)
	require.NoError(t, err)

	return info
}

func TestGenesis_Validator_Add(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis file", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"validator",
			"add",
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

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		key := getDummyKey(t)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"validator",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			key.GetPubKey().Address().String(),
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

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		key := getDummyKey(t)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"validator",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			key.GetPubKey().Address().String(),
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

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		key := getDummyKey(t)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"validator",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			key.GetPubKey().Address().String(),
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

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		dummyKeys := []keys.Info{
			getDummyKey(t),
			getDummyKey(t),
		}

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"validator",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			dummyKeys[0].GetPubKey().Address().String(),
			"--name",
			"example",
			"--pub-key",
			crypto.PubKeyToBech32(dummyKeys[1].GetPubKey()), // another key
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errPublicKeyMismatch.Error())
	})

	t.Run("validator with same address exists", func(t *testing.T) {
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
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			dummyKeys[0].GetPubKey().Address().String(),
			"--name",
			"example",
			"--pub-key",
			crypto.PubKeyToBech32(dummyKeys[0].GetPubKey()), // another key
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errAddressPresent.Error())
	})

	t.Run("valid genesis validator", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		key := getDummyKey(t)
		genesis := getDefaultGenesis()

		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"validator",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			"--address",
			key.GetPubKey().Address().String(),
			"--name",
			"example",
			"--pub-key",
			crypto.PubKeyToBech32(key.GetPubKey()), // another key
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)
	})
}
