package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenesis_Balances_Add(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis", func(t *testing.T) {
		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"balances",
			"add",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.ErrorContains(t, cmdErr, errUnableToLoadGenesis.Error())
	})

	t.Run("no sources selected", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"balances",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errNoBalanceSource.Error())
	})

	t.Run("invalid genesis path", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"balances",
			"add",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errUnableToLoadGenesis.Error())
	})

	t.Run("balances from entries", func(t *testing.T) {
		t.Parallel()

		dummyKeys := getDummyKeys(t, 2)

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"balances",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
		}

		amount := std.NewCoins(std.NewCoin("ugnot", 10))

		for _, dummyKey := range dummyKeys {
			args = append(args, "--single")
			args = append(
				args,
				fmt.Sprintf(
					"%s=%dugnot",
					dummyKey.Address().String(),
					amount.AmountOf("ugnot"),
				),
			)
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

		require.Equal(t, len(dummyKeys), len(state.Balances))

		for _, balance := range state.Balances {
			// Find the appropriate key
			// (the genesis is saved with randomized balance order)
			found := false
			for _, dummyKey := range dummyKeys {
				if dummyKey.Address().String() == balance.Address.String() {
					assert.Equal(t, amount, balance.Amount)

					found = true
					break
				}
			}

			if !found {
				t.Fatalf("unexpected entry with address %s found", balance.Address.String())
			}
		}
	})

	t.Run("balances from sheet", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		dummyKeys := getDummyKeys(t, 10)
		amount := std.NewCoins(std.NewCoin("ugnot", 10))

		balances := make([]string, len(dummyKeys))

		// Add a random comment to the balances file output
		balances = append(balances, "#comment\n")

		for index, key := range dummyKeys {
			balances[index] = fmt.Sprintf(
				"%s=%dugnot",
				key.Address().String(),
				amount.AmountOf("ugnot"),
			)
		}

		// Write the balance sheet to a file
		balanceSheet, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		_, err := balanceSheet.WriteString(strings.Join(balances, "\n"))
		require.NoError(t, err)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"balances",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			"--balance-sheet",
			balanceSheet.Name(),
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

		require.Equal(t, len(dummyKeys), len(state.Balances))

		for _, balance := range state.Balances {
			// Find the appropriate key
			// (the genesis is saved with randomized balance order)
			found := false
			for _, dummyKey := range dummyKeys {
				if dummyKey.Address().String() == balance.Address.String() {
					assert.Equal(t, amount, balance.Amount)

					found = true
					break
				}
			}

			if !found {
				t.Fatalf("unexpected entry with address %s found", balance.Address.String())
			}
		}
	})

	t.Run("balances from transactions", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		var (
			dummyKeys   = getDummyKeys(t, 10)
			amount      = std.NewCoins(std.NewCoin("ugnot", 10))
			amountCoins = std.NewCoins(std.NewCoin("ugnot", 10))
			gasFee      = std.NewCoin("ugnot", 1000000)
			txs         = make([]std.Tx, 0)
		)

		sender := dummyKeys[0]
		for _, dummyKey := range dummyKeys[1:] {
			tx := std.Tx{
				Msgs: []std.Msg{
					bank.MsgSend{
						FromAddress: sender.Address(),
						ToAddress:   dummyKey.Address(),
						Amount:      amountCoins,
					},
				},
				Fee: std.Fee{
					GasWanted: 10,
					GasFee:    gasFee,
				},
				Signatures: make([]std.Signature, 0),
			}

			txs = append(txs, tx)
		}

		// Marshal the transactions into amino JSON
		marshalledTxs := make([]string, 0, len(txs))

		for _, tx := range txs {
			marshalledTx, err := amino.MarshalJSON(tx)
			require.NoError(t, err)

			marshalledTxs = append(marshalledTxs, string(marshalledTx))
		}

		// Write the transactions to a file
		txsFile, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		_, err := txsFile.WriteString(strings.Join(marshalledTxs, "\n"))
		require.NoError(t, err)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"balances",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			"--parse-export",
			txsFile.Name(),
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

		require.Equal(t, len(dummyKeys), len(state.Balances))

		for _, balance := range state.Balances {
			// Find the appropriate key
			// (the genesis is saved with randomized balance order)
			found := false
			for index, dummyKey := range dummyKeys {
				checkAmount := amount
				if index == 0 {
					// the first address should
					// have a balance of 0
					checkAmount = std.NewCoins(std.NewCoin("ugnot", 0))
				}

				if dummyKey.Address().String() == balance.Address.String() {
					assert.True(t, balance.Amount.IsEqual(checkAmount))

					found = true
					break
				}
			}

			if !found {
				t.Fatalf("unexpected entry with address %s found", balance.Address.String())
			}
		}
	})

	t.Run("balances overwrite", func(t *testing.T) {
		t.Parallel()

		dummyKeys := getDummyKeys(t, 10)

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := getDefaultGenesis()
		state := gnoland.GnoGenesisState{
			// Set an initial balance value
			Balances: []gnoland.Balance{
				{
					Address: dummyKeys[0].Address(),
					Amount:  std.NewCoins(std.NewCoin("ugnot", 100)),
				},
			},
		}
		genesis.AppState = state
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"genesis",
			"balances",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
		}

		amount := std.NewCoins(std.NewCoin("ugnot", 10))

		for _, dummyKey := range dummyKeys {
			args = append(args, "--single")
			args = append(
				args,
				fmt.Sprintf(
					"%s=%dugnot",
					dummyKey.Address().String(),
					amount.AmountOf("ugnot"),
				),
			)
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

		require.Equal(t, len(dummyKeys), len(state.Balances))

		for _, balance := range state.Balances {
			// Find the appropriate key
			// (the genesis is saved with randomized balance order)
			found := false
			for _, dummyKey := range dummyKeys {
				if dummyKey.Address().String() == balance.Address.String() {
					assert.Equal(t, amount, balance.Amount)

					found = true
					break
				}
			}

			if !found {
				t.Fatalf("unexpected entry with address %s found", balance.Address.String())
			}
		}
	})
}

func TestBalances_GetBalancesFromTransactions(t *testing.T) {
	t.Parallel()

	t.Run("valid transactions", func(t *testing.T) {
		t.Parallel()

		var (
			dummyKeys   = getDummyKeys(t, 10)
			amount      = std.NewCoins(std.NewCoin("ugnot", 10))
			amountCoins = std.NewCoins(std.NewCoin("ugnot", 10))
			gasFee      = std.NewCoin("ugnot", 1000000)
			txs         = make([]std.Tx, 0)
		)

		sender := dummyKeys[0]
		for _, dummyKey := range dummyKeys[1:] {
			tx := std.Tx{
				Msgs: []std.Msg{
					bank.MsgSend{
						FromAddress: sender.Address(),
						ToAddress:   dummyKey.Address(),
						Amount:      amountCoins,
					},
				},
				Fee: std.Fee{
					GasWanted: 10,
					GasFee:    gasFee,
				},
				Signatures: make([]std.Signature, 0),
			}

			txs = append(txs, tx)
		}

		// Marshal the transactions into amino JSON
		marshalledTxs := make([]string, 0, len(txs))

		for _, tx := range txs {
			marshalledTx, err := amino.MarshalJSON(tx)
			require.NoError(t, err)

			marshalledTxs = append(marshalledTxs, string(marshalledTx))
		}

		mockErr := new(bytes.Buffer)
		io := commands.NewTestIO()
		io.SetErr(commands.WriteNopCloser(mockErr))

		reader := strings.NewReader(strings.Join(marshalledTxs, "\n"))
		balanceMap, err := getBalancesFromTransactions(context.Background(), io, reader)
		require.NoError(t, err)

		// Validate the balance map
		assert.Len(t, balanceMap, len(dummyKeys))
		for _, key := range dummyKeys[1:] {
			assert.Equal(t, amount, balanceMap[key.Address()].Amount)
		}

		assert.Equal(t, std.Coins{}, balanceMap[sender.Address()].Amount)
	})

	t.Run("malformed transaction, invalid fee amount", func(t *testing.T) {
		t.Parallel()

		var (
			dummyKeys   = getDummyKeys(t, 10)
			amountCoins = std.NewCoins(std.NewCoin("ugnot", 10))
			gasFee      = std.NewCoin("gnos", 1) // invalid fee
			txs         = make([]std.Tx, 0)
		)

		sender := dummyKeys[0]
		for _, dummyKey := range dummyKeys[1:] {
			tx := std.Tx{
				Msgs: []std.Msg{
					bank.MsgSend{
						FromAddress: sender.Address(),
						ToAddress:   dummyKey.Address(),
						Amount:      amountCoins,
					},
				},
				Fee: std.Fee{
					GasWanted: 10,
					GasFee:    gasFee,
				},
				Signatures: make([]std.Signature, 0),
			}

			txs = append(txs, tx)
		}

		// Marshal the transactions into amino JSON
		marshalledTxs := make([]string, 0, len(txs))

		for _, tx := range txs {
			marshalledTx, err := amino.MarshalJSON(tx)
			require.NoError(t, err)

			marshalledTxs = append(marshalledTxs, string(marshalledTx))
		}

		mockErr := new(bytes.Buffer)
		io := commands.NewTestIO()
		io.SetErr(commands.WriteNopCloser(mockErr))

		reader := strings.NewReader(strings.Join(marshalledTxs, "\n"))
		balanceMap, err := getBalancesFromTransactions(context.Background(), io, reader)
		require.NoError(t, err)

		assert.NotNil(t, balanceMap)
		assert.Contains(t, mockErr.String(), "invalid gas fee amount")
	})

	t.Run("malformed transaction, invalid send amount", func(t *testing.T) {
		t.Parallel()

		var (
			dummyKeys   = getDummyKeys(t, 10)
			amountCoins = std.NewCoins(std.NewCoin("gnogno", 10)) // invalid send amount
			gasFee      = std.NewCoin("ugnot", 1)
			txs         = make([]std.Tx, 0)
		)

		sender := dummyKeys[0]
		for _, dummyKey := range dummyKeys[1:] {
			tx := std.Tx{
				Msgs: []std.Msg{
					bank.MsgSend{
						FromAddress: sender.Address(),
						ToAddress:   dummyKey.Address(),
						Amount:      amountCoins,
					},
				},
				Fee: std.Fee{
					GasWanted: 10,
					GasFee:    gasFee,
				},
				Signatures: make([]std.Signature, 0),
			}

			txs = append(txs, tx)
		}

		// Marshal the transactions into amino JSON
		marshalledTxs := make([]string, 0, len(txs))

		for _, tx := range txs {
			marshalledTx, err := amino.MarshalJSON(tx)
			require.NoError(t, err)

			marshalledTxs = append(marshalledTxs, string(marshalledTx))
		}

		mockErr := new(bytes.Buffer)
		io := commands.NewTestIO()
		io.SetErr(commands.WriteNopCloser(mockErr))

		reader := strings.NewReader(strings.Join(marshalledTxs, "\n"))
		balanceMap, err := getBalancesFromTransactions(context.Background(), io, reader)
		require.NoError(t, err)

		assert.NotNil(t, balanceMap)
		assert.Contains(t, mockErr.String(), "invalid send amount")
	})
}
