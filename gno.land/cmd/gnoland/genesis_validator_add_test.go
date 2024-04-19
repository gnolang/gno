package main

import (
	"context"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getDummyKey generates a random public key,
// and returns the key info
func getDummyKey(t *testing.T) crypto.PubKey {
	t.Helper()

	mnemonic, err := client.GenerateMnemonic(256)
	require.NoError(t, err)

	seed := bip39.NewSeed(mnemonic, "")

	return generateKeyFromSeed(seed, 0).PubKey()
}

// generateKeyFromSeed generates a private key from
// the provided seed and index
func generateKeyFromSeed(seed []byte, index uint32) crypto.PrivKey {
	pathParams := hd.NewFundraiserParams(0, crypto.CoinType, index)

	masterPriv, ch := hd.ComputeMastersFromSeed(seed)

	//nolint:errcheck // This derivation can never error out, since the path params
	// are always going to be valid
	derivedPriv, _ := hd.DerivePrivateKeyForPath(masterPriv, ch, pathParams.String())

	return secp256k1.PrivKeySecp256k1(derivedPriv)
}

// getDummyKeys generates random keys for testing
func getDummyKeys(t *testing.T, count int) []crypto.PubKey {
	t.Helper()

	dummyKeys := make([]crypto.PubKey, count)

	for i := 0; i < count; i++ {
		dummyKeys[i] = getDummyKey(t)
	}

	return dummyKeys
}

func TestGenesis_Validator_Add(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis file", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"validator",
			"add",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errUnableToLoadGenesis.Error())
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
			"genesis",
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
			"genesis",
			"validator",
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

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		key := getDummyKey(t)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"validator",
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

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		key := getDummyKey(t)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"validator",
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

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		dummyKeys := getDummyKeys(t, 2)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"validator",
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

		dummyKeys := getDummyKeys(t, 2)
		genesis := getDefaultGenesis()

		// Set an existing validator
		genesis.Validators = append(genesis.Validators, types.GenesisValidator{
			Address: dummyKeys[0].Address(),
			PubKey:  dummyKeys[0],
			Power:   1,
			Name:    "example",
		})

		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"validator",
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

		key := getDummyKey(t)
		genesis := getDefaultGenesis()

		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"validator",
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
