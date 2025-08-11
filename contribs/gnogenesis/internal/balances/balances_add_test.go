package balances

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
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
		cmd := NewBalancesCmd(commands.NewTestIO())
		args := []string{
			"add",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.ErrorContains(t, cmdErr, common.ErrUnableToLoadGenesis.Error())
	})

	t.Run("no sources selected", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := NewBalancesCmd(commands.NewTestIO())
		args := []string{
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
		cmd := NewBalancesCmd(commands.NewTestIO())
		args := []string{
			"add",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, common.ErrUnableToLoadGenesis.Error())
	})

	t.Run("balances from entries", func(t *testing.T) {
		t.Parallel()

		dummyKeys := common.DummyKeys(t, 2)

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := NewBalancesCmd(commands.NewTestIO())
		args := []string{
			"add",
			"--genesis-path",
			tempGenesis.Name(),
		}

		amount := std.NewCoins(std.NewCoin(ugnot.Denom, 10))

		for _, dummyKey := range dummyKeys {
			args = append(args, "--single")
			args = append(
				args,
				fmt.Sprintf(
					"%s=%s",
					dummyKey.Address().String(),
					ugnot.ValueString(amount.AmountOf(ugnot.Denom)),
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

		genesis := common.DefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		dummyKeys := common.DummyKeys(t, 10)
		amount := std.NewCoins(std.NewCoin(ugnot.Denom, 10))

		balances := make([]string, len(dummyKeys))

		// Add a random comment to the balances file output
		balances = append(balances, "#comment\n")

		for index, key := range dummyKeys {
			balances[index] = fmt.Sprintf(
				"%s=%s",
				key.Address().String(),
				ugnot.ValueString(amount.AmountOf(ugnot.Denom)),
			)
		}

		// Write the balance sheet to a file
		balanceSheet, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		_, err := balanceSheet.WriteString(strings.Join(balances, "\n"))
		require.NoError(t, err)

		// Create the command
		cmd := NewBalancesCmd(commands.NewTestIO())
		args := []string{
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

	t.Run("deterministic balances from sheet", func(t *testing.T) {
		t.Parallel()

		const (
			nKeys      = 100
			nRuns      = 10
			coinAmount = int64(10)
		)

		equalBalances := func(b1, b2 gnoland.Balance) bool {
			return b1.Address.Compare(b2.Address) == 0 && b1.Amount.IsEqual(b2.Amount)
		}

		dummyKeys := common.DummyKeys(t, nKeys)
		amount := std.NewCoins(std.NewCoin(ugnot.Denom, coinAmount))

		// Prepare the balance sheet
		lines := make([]string, 0, nKeys+1)
		lines = append(lines, "#comment") // random comment on top

		for _, key := range dummyKeys {
			lines = append(
				lines,
				fmt.Sprintf(
					"%s=%s",
					key.Address().String(),
					ugnot.ValueString(amount.AmountOf(ugnot.Denom)),
				),
			)
		}

		balanceSheet, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		_, err := balanceSheet.WriteString(strings.Join(lines, "\n"))
		require.NoError(t, err)

		var referenceBalances []gnoland.Balance

		for run := 0; run < nRuns; run++ {
			// Create a fresh genesis file
			tempGenesis, cleanupGen := testutils.NewTestFile(t)
			t.Cleanup(cleanupGen)

			genesis := common.DefaultGenesis()
			require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

			// Add the balance sheet
			cmd := NewBalancesCmd(commands.NewTestIO())
			args := []string{
				"add",
				"--genesis-path", tempGenesis.Name(),
				"--balance-sheet", balanceSheet.Name(),
			}

			require.NoError(t, cmd.ParseAndRun(context.Background(), args))

			// Load the modified genesis
			genesisDoc, err := types.GenesisDocFromFile(tempGenesis.Name())
			require.NoError(t, err)
			require.NotNil(t, genesisDoc.AppState)

			state, ok := genesisDoc.AppState.(gnoland.GnoGenesisState)
			require.True(t, ok)
			require.Len(t, state.Balances, nKeys)

			// The first run should be the reference one
			if run == 0 {
				referenceBalances = state.Balances

				continue
			}

			for index, balance := range state.Balances {
				assert.True(t, equalBalances(referenceBalances[index], balance))
			}
		}
	})

	t.Run("balances from transactions", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		var (
			dummyKeys   = common.DummyKeys(t, 10)
			amount      = std.NewCoins(std.NewCoin(ugnot.Denom, 10))
			amountCoins = std.NewCoins(std.NewCoin(ugnot.Denom, 10))
			gasFee      = std.NewCoin(ugnot.Denom, 1000000)
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
		cmd := NewBalancesCmd(commands.NewTestIO())
		args := []string{
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
					checkAmount = std.NewCoins(std.NewCoin(ugnot.Denom, 0))
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

		dummyKeys := common.DummyKeys(t, 10)

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.DefaultGenesis()
		state := gnoland.GnoGenesisState{
			// Set an initial balance value
			Balances: []gnoland.Balance{
				{
					Address: dummyKeys[0].Address(),
					Amount:  std.NewCoins(std.NewCoin(ugnot.Denom, 100)),
				},
			},
		}
		genesis.AppState = state
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := NewBalancesCmd(commands.NewTestIO())
		args := []string{
			"add",
			"--genesis-path",
			tempGenesis.Name(),
		}

		amount := std.NewCoins(std.NewCoin(ugnot.Denom, 10))

		for _, dummyKey := range dummyKeys {
			args = append(args, "--single")
			args = append(
				args,
				fmt.Sprintf(
					"%s=%s",
					dummyKey.Address().String(),
					ugnot.ValueString(amount.AmountOf(ugnot.Denom)),
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
			dummyKeys   = common.DummyKeys(t, 10)
			amount      = std.NewCoins(std.NewCoin(ugnot.Denom, 10))
			amountCoins = std.NewCoins(std.NewCoin(ugnot.Denom, 10))
			gasFee      = std.NewCoin(ugnot.Denom, 1000000)
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
			dummyKeys   = common.DummyKeys(t, 10)
			amountCoins = std.NewCoins(std.NewCoin(ugnot.Denom, 10))
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
			dummyKeys   = common.DummyKeys(t, 10)
			amountCoins = std.NewCoins(std.NewCoin("gnogno", 10)) // invalid send amount
			gasFee      = std.NewCoin(ugnot.Denom, 1)
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
