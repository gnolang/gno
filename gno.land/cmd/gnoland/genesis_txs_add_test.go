package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateDummyTxs generates dummy transactions
func generateDummyTxs(t *testing.T, count int) []std.Tx {
	t.Helper()

	txs := make([]std.Tx, count)

	for i := 0; i < count; i++ {
		txs[i] = std.Tx{
			Msgs: []std.Msg{
				bank.MsgSend{
					FromAddress: crypto.Address{byte(i)},
					ToAddress:   crypto.Address{byte((i + 1) % count)},
					Amount:      std.NewCoins(std.NewCoin("ugnot", 1)),
				},
			},
			Fee: std.Fee{
				GasWanted: 1,
				GasFee:    std.NewCoin("ugnot", 1000000),
			},
			Memo: fmt.Sprintf("tx %d", i),
		}
	}

	return txs
}

// encodeDummyTxs encodes the transactions into amino JSON
func encodeDummyTxs(t *testing.T, txs []std.Tx) []string {
	t.Helper()

	encodedTxs := make([]string, 0, len(txs))

	for _, tx := range txs {
		encodedTx, err := amino.MarshalJSON(tx)
		if err != nil {
			t.Fatalf("unable to marshal tx, %v", err)
		}

		encodedTxs = append(encodedTxs, string(encodedTx))
	}

	return encodedTxs
}

func TestGenesis_Txs_Add(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis file", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"txs",
			"add",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errUnableToLoadGenesis.Error())
	})

	t.Run("invalid txs file", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"txs",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			"dummy-tx-file",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errInvalidTxsFile.Error())
	})

	t.Run("no txs file", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"txs",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errNoTxsFileSpecified.Error())
	})

	t.Run("malformed txs file", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"txs",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			tempGenesis.Name(), // invalid txs file
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "unable to read file")
	})

	t.Run("valid txs file", func(t *testing.T) {
		t.Parallel()

		// Generate dummy txs
		txs := generateDummyTxs(t, 10)

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Prepare the transactions file
		txsFile, txsCleanup := testutils.NewTestFile(t)
		t.Cleanup(txsCleanup)

		_, err := txsFile.WriteString(
			strings.Join(
				encodeDummyTxs(t, txs),
				"\n",
			),
		)
		require.NoError(t, err)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"txs",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			txsFile.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Validate the transactions were written down
		updatedGenesis, err := types.GenesisDocFromFile(tempGenesis.Name())
		require.NoError(t, err)
		require.NotNil(t, updatedGenesis.AppState)

		// Fetch the state
		state := updatedGenesis.AppState.(gnoland.GnoGenesisState)

		assert.Len(t, state.Txs, len(txs))

		for index, tx := range state.Txs {
			assert.Equal(t, txs[index], tx)
		}
	})

	t.Run("existing genesis txs", func(t *testing.T) {
		t.Parallel()

		// Generate dummy txs
		txs := generateDummyTxs(t, 10)

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := getDefaultGenesis()
		genesisState := gnoland.GnoGenesisState{
			Txs: txs[0 : len(txs)/2],
		}

		genesis.AppState = genesisState
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Prepare the transactions file
		txsFile, txsCleanup := testutils.NewTestFile(t)
		t.Cleanup(txsCleanup)

		_, err := txsFile.WriteString(
			strings.Join(
				encodeDummyTxs(t, txs),
				"\n",
			),
		)
		require.NoError(t, err)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"txs",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			txsFile.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		// Validate the transactions were written down
		updatedGenesis, err := types.GenesisDocFromFile(tempGenesis.Name())
		require.NoError(t, err)
		require.NotNil(t, updatedGenesis.AppState)

		// Fetch the state
		state := updatedGenesis.AppState.(gnoland.GnoGenesisState)

		assert.Len(t, state.Txs, len(txs))

		for index, tx := range state.Txs {
			assert.Equal(t, txs[index], tx)
		}
	})
}
