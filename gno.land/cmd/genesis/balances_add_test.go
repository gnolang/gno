package main

import (
	"context"
	"fmt"
	"math"
	"strconv"
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

	t.Run("no sources selected", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"balances",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errNoBalanceSource.Error())
	})

	t.Run("more than one source selected", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := getDefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"balances",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
			"--balance-sheet",
			"dummy-sheet",
			"--single",
			"single-entry",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, errMultipleBalanceSources.Error())
	})

	t.Run("invalid genesis path", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"balances",
			"add",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "unable to load genesis")
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
			"balances",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
		}

		amount := int64(10)

		for _, dummyKey := range dummyKeys {
			args = append(args, "--single")
			args = append(
				args,
				fmt.Sprintf(
					"%s=%dugnot",
					dummyKey.Address().String(),
					amount,
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

		for _, entry := range state.Balances {
			accountBalance, err := getBalanceFromEntry(entry)
			require.NoError(t, err)

			// Find the appropriate key
			// (the genesis is saved as amino JSON, which is sorted,
			// meaning the order is not guaranteed)
			found := false
			for _, dummyKey := range dummyKeys {
				if dummyKey.Address().String() == accountBalance.address.String() {
					assert.Equal(t, amount, accountBalance.amount)

					found = true
					break
				}
			}

			if !found {
				t.Fatalf("unexpected entry with address %s found", accountBalance.address.String())
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
		amount := int64(10)

		balances := make([]string, len(dummyKeys))

		for index, key := range dummyKeys {
			balances[index] = fmt.Sprintf(
				"%s=%dugnot",
				key.Address().String(),
				amount,
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

		for _, entry := range state.Balances {
			accountBalance, err := getBalanceFromEntry(entry)
			require.NoError(t, err)

			// Find the appropriate key
			// (the genesis is saved as amino JSON, which is sorted,
			// meaning the order is not guaranteed)
			found := false
			for _, dummyKey := range dummyKeys {
				if dummyKey.Address().String() == accountBalance.address.String() {
					assert.Equal(t, amount, accountBalance.amount)

					found = true
					break
				}
			}

			if !found {
				t.Fatalf("unexpected entry with address %s found", accountBalance.address.String())
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
			amount      = int64(10)
			amountCoins = std.NewCoins(std.NewCoin("ugnot", amount))
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

		for _, entry := range state.Balances {
			accountBalance, err := getBalanceFromEntry(entry)
			require.NoError(t, err)

			// Find the appropriate key
			// (the genesis is saved as amino JSON, which is sorted,
			// meaning the order is not guaranteed)
			found := false
			for index, dummyKey := range dummyKeys {
				checkAmount := amount
				if index == 0 {
					// the first address should
					// have a balance of 0
					checkAmount = 0
				}

				if dummyKey.Address().String() == accountBalance.address.String() {
					assert.Equal(t, checkAmount, accountBalance.amount)

					found = true
					break
				}
			}

			if !found {
				t.Fatalf("unexpected entry with address %s found", accountBalance.address.String())
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
			Balances: []string{
				fmt.Sprintf(
					"%s=%dugnot",
					dummyKeys[0].Address().String(),
					100,
				),
			},
		}
		genesis.AppState = state
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"balances",
			"add",
			"--genesis-path",
			tempGenesis.Name(),
		}

		amount := int64(10)

		for _, dummyKey := range dummyKeys {
			args = append(args, "--single")
			args = append(
				args,
				fmt.Sprintf(
					"%s=%dugnot",
					dummyKey.Address().String(),
					amount,
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

		for _, entry := range state.Balances {
			accountBalance, err := getBalanceFromEntry(entry)
			require.NoError(t, err)

			// Find the appropriate key
			// (the genesis is saved as amino JSON, which is sorted,
			// meaning the order is not guaranteed)
			found := false
			for _, dummyKey := range dummyKeys {
				if dummyKey.Address().String() == accountBalance.address.String() {
					assert.Equal(t, amount, accountBalance.amount)

					found = true
					break
				}
			}

			if !found {
				t.Fatalf("unexpected entry with address %s found", accountBalance.address.String())
			}
		}
	})
}

func TestBalances_GetBalancesFromEntries(t *testing.T) {
	t.Parallel()

	t.Run("valid balances", func(t *testing.T) {
		t.Parallel()

		// Generate dummy keys
		dummyKeys := getDummyKeys(t, 2)
		amount := int64(10)

		balances := make([]string, len(dummyKeys))

		for index, key := range dummyKeys {
			balances[index] = fmt.Sprintf(
				"%s=%dugnot",
				key.Address().String(),
				amount,
			)
		}

		balanceMap, err := getBalancesFromEntries(balances)
		require.NoError(t, err)

		// Validate the balance map
		assert.Len(t, balanceMap, len(dummyKeys))
		for _, key := range dummyKeys {
			assert.Equal(t, amount, balanceMap[key.Address()])
		}
	})

	t.Run("malformed balance, invalid format", func(t *testing.T) {
		t.Parallel()

		balances := []string{
			"malformed balance",
		}

		balanceMap, err := getBalancesFromEntries(balances)

		assert.Nil(t, balanceMap)
		assert.ErrorContains(t, err, errInvalidBalanceFormat.Error())
	})

	t.Run("malformed balance, invalid address", func(t *testing.T) {
		t.Parallel()

		balances := []string{
			"dummyaddress=10ugnot",
		}

		balanceMap, err := getBalancesFromEntries(balances)

		assert.Nil(t, balanceMap)
		assert.ErrorContains(t, err, errInvalidAddress.Error())
	})

	t.Run("malformed balance, invalid amount", func(t *testing.T) {
		t.Parallel()

		dummyKey := getDummyKey(t)

		balances := []string{
			fmt.Sprintf(
				"%s=%sugnot",
				dummyKey.Address().String(),
				strconv.FormatUint(math.MaxUint64, 10),
			),
		}

		balanceMap, err := getBalancesFromEntries(balances)

		assert.Nil(t, balanceMap)
		assert.ErrorContains(t, err, errInvalidAmount.Error())
	})
}

func TestBalances_GetBalancesFromSheet(t *testing.T) {
	t.Parallel()

	t.Run("valid balances", func(t *testing.T) {
		t.Parallel()

		// Generate dummy keys
		dummyKeys := getDummyKeys(t, 2)
		amount := int64(10)

		balances := make([]string, len(dummyKeys))

		for index, key := range dummyKeys {
			balances[index] = fmt.Sprintf(
				"%s=%dugnot",
				key.Address().String(),
				amount,
			)
		}

		reader := strings.NewReader(strings.Join(balances, "\n"))
		balanceMap, err := getBalancesFromSheet(reader)
		require.NoError(t, err)

		// Validate the balance map
		assert.Len(t, balanceMap, len(dummyKeys))
		for _, key := range dummyKeys {
			assert.Equal(t, amount, balanceMap[key.Address()])
		}
	})

	t.Run("malformed balance, invalid amount", func(t *testing.T) {
		t.Parallel()

		dummyKey := getDummyKey(t)

		balances := []string{
			fmt.Sprintf(
				"%s=%sugnot",
				dummyKey.Address().String(),
				strconv.FormatUint(math.MaxUint64, 10),
			),
		}

		reader := strings.NewReader(strings.Join(balances, "\n"))

		balanceMap, err := getBalancesFromSheet(reader)

		assert.Nil(t, balanceMap)
		assert.ErrorContains(t, err, errInvalidAmount.Error())
	})
}

func TestBalances_GetBalancesFromTransactions(t *testing.T) {
	t.Parallel()

	t.Run("valid transactions", func(t *testing.T) {
		t.Parallel()

		var (
			dummyKeys   = getDummyKeys(t, 10)
			amount      = int64(10)
			amountCoins = std.NewCoins(std.NewCoin("ugnot", amount))
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

		reader := strings.NewReader(strings.Join(marshalledTxs, "\n"))
		balanceMap, err := getBalancesFromTransactions(context.Background(), reader)
		require.NoError(t, err)

		// Validate the balance map
		assert.Len(t, balanceMap, len(dummyKeys))
		for _, key := range dummyKeys[1:] {
			assert.Equal(t, amount, balanceMap[key.Address()])
		}

		assert.Equal(t, int64(0), balanceMap[sender.Address()])
	})

	t.Run("malformed transaction, invalid fee amount", func(t *testing.T) {
		t.Parallel()

		var (
			dummyKeys   = getDummyKeys(t, 10)
			amount      = int64(10)
			amountCoins = std.NewCoins(std.NewCoin("ugnot", amount))
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

		reader := strings.NewReader(strings.Join(marshalledTxs, "\n"))
		balanceMap, err := getBalancesFromTransactions(context.Background(), reader)

		assert.Nil(t, balanceMap)
		assert.ErrorContains(t, err, "invalid gas fee amount")
	})

	t.Run("malformed transaction, invalid send amount", func(t *testing.T) {
		t.Parallel()

		var (
			dummyKeys   = getDummyKeys(t, 10)
			amount      = int64(10)
			amountCoins = std.NewCoins(std.NewCoin("gnogno", amount)) // invalid send amount
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

		reader := strings.NewReader(strings.Join(marshalledTxs, "\n"))
		balanceMap, err := getBalancesFromTransactions(context.Background(), reader)

		assert.Nil(t, balanceMap)
		assert.ErrorContains(t, err, "invalid send amount")
	})
}
