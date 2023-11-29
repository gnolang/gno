package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"

	_ "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
)

var (
	errNoBalanceSource       = errors.New("at least one balance source must be set")
	errBalanceParsingAborted = errors.New("balance parsing aborted")
	errInvalidAddress        = errors.New("invalid address encountered")
)

type balancesAddCfg struct {
	rootCfg *balancesCfg

	balanceSheet  string
	singleEntries commands.StringArr
	parseExport   string
}

// newBalancesAddCmd creates the genesis balances add subcommand
func newBalancesAddCmd(rootCfg *balancesCfg, io commands.IO) *commands.Command {
	cfg := &balancesAddCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "add",
			ShortUsage: "balances add [flags]",
			LongHelp:   "Adds a new validator to the genesis.json",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execBalancesAdd(ctx, cfg, io)
		},
	)
}

func (c *balancesAddCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.balanceSheet,
		"balance-sheet",
		"",
		"the path to the balance file containing addresses in the format <address>=<amount>ugnot",
	)

	fs.Var(
		&c.singleEntries,
		"single",
		"the direct balance addition in the format <address>=<amount>ugnot",
	)

	fs.StringVar(
		&c.parseExport,
		"parse-export",
		"",
		"the path to the transaction export containing a list of transactions (JSONL)",
	)
}

func execBalancesAdd(ctx context.Context, cfg *balancesAddCfg, io commands.IO) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.rootCfg.genesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Validate the source is set correctly
	var (
		singleEntriesSet = len(cfg.singleEntries) != 0
		balanceSheetSet  = cfg.balanceSheet != ""
		txFileSet        = cfg.parseExport != ""
	)

	if !singleEntriesSet && !balanceSheetSet && !txFileSet {
		return errNoBalanceSource
	}

	finalBalances := make(accountBalances)

	// Get the balance sheet from the source
	if singleEntriesSet {
		balances, err := getBalancesFromEntries(cfg.singleEntries)
		if err != nil {
			return fmt.Errorf("unable to get balances from entries, %w", err)
		}

		finalBalances.leftMerge(balances)
	}

	if balanceSheetSet {
		// Open the balance sheet
		file, loadErr := os.Open(cfg.balanceSheet)
		if loadErr != nil {
			return fmt.Errorf("unable to open balance sheet, %w", loadErr)
		}

		balances, err := getBalancesFromSheet(file)
		if err != nil {
			return fmt.Errorf("unable to get balances from balance sheet, %w", err)
		}

		finalBalances.leftMerge(balances)
	}

	if txFileSet {
		// Open the transactions file
		file, loadErr := os.Open(cfg.parseExport)
		if loadErr != nil {
			return fmt.Errorf("unable to open transactions file, %w", loadErr)
		}

		balances, err := getBalancesFromTransactions(ctx, io, file)
		if err != nil {
			return fmt.Errorf("unable to get balances from tx file, %w", err)
		}

		finalBalances.leftMerge(balances)
	}

	// Initialize genesis app state if it is not initialized already
	if genesis.AppState == nil {
		genesis.AppState = gnoland.GnoGenesisState{}
	}

	// Construct the initial genesis balance sheet
	state := genesis.AppState.(gnoland.GnoGenesisState)
	genesisBalances, err := mapGenesisBalancesFromState(state)
	if err != nil {
		return err
	}

	// Merge the two balance sheets, with the input
	// having precedence over the genesis balances
	finalBalances.leftMerge(genesisBalances)

	// Save the balances
	state.Balances = finalBalances.toList()
	genesis.AppState = state

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.rootCfg.genesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	io.Printfln(
		"%d pre-mines saved",
		len(finalBalances),
	)

	io.Println()

	for address, balance := range finalBalances {
		io.Printfln("%s:%dugnot", address.String(), balance)
	}

	return nil
}

// getBalancesFromEntries extracts the balance entries
// from the array of balance
func getBalancesFromEntries(entries []string) (accountBalances, error) {
	balances := make(accountBalances)

	for _, entry := range entries {
		var balance gnoland.Balance
		if err := balance.Parse(entry); err != nil {
			return nil, fmt.Errorf("unable to parse balance entry: %w", err)
		}
		balances[balance.Address] = balance
	}

	return balances, nil
}

// getBalancesFromSheet extracts the balance sheet from the passed in
// balance sheet file, that has the format of <address>=<amount>ugnot
func getBalancesFromSheet(sheet io.Reader) (accountBalances, error) {
	// Parse the balances
	balances := make(accountBalances)
	scanner := bufio.NewScanner(sheet)

	for scanner.Scan() {
		entry := scanner.Text()

		// Remove comments
		entry = strings.Split(entry, "#")[0]
		entry = strings.TrimSpace(entry)

		// Skip empty lines
		if entry == "" {
			continue
		}

		var balance gnoland.Balance
		if err := balance.Parse(entry); err != nil {
			return nil, fmt.Errorf("unable to extract balance data, %w", err)
		}

		balances[balance.Address] = balance
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error encountered while scanning, %w", err)
	}

	return balances, nil
}

// getBalancesFromTransactions constructs a balance map based on MsgSend messages.
// This way of determining the final balance sheet is not valid, since it doesn't take into
// account different message types (ex. MsgCall) that can initialize accounts with some balance values.
// The right way to do this sort of initialization is to spin up an in-memory node
// and execute the entire transaction history to determine touched accounts and final balances,
// and construct a balance sheet based off of this information
func getBalancesFromTransactions(
	ctx context.Context,
	io commands.IO,
	reader io.Reader,
) (accountBalances, error) {
	balances := make(accountBalances)

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, errBalanceParsingAborted
		default:
			// Parse the amino JSON
			var tx std.Tx

			line := scanner.Bytes()

			if err := amino.UnmarshalJSON(line, &tx); err != nil {
				io.ErrPrintfln(
					"invalid amino JSON encountered: %q",
					string(line),
				)

				continue
			}

			feeAmount := std.NewCoins(tx.Fee.GasFee)
			if feeAmount.AmountOf("ugnot") <= 0 {
				io.ErrPrintfln(
					"invalid gas fee amount encountered: %q",
					tx.Fee.GasFee.String(),
				)
			}

			for _, msg := range tx.Msgs {
				if msg.Type() != "send" {
					continue
				}

				msgSend := msg.(bank.MsgSend)

				sendAmount := msgSend.Amount
				if sendAmount.AmountOf("ugnot") <= 0 {
					io.ErrPrintfln(
						"invalid send amount encountered: %s",
						msgSend.Amount.String(),
					)
					continue
				}

				// This way of determining final account balances is not really valid,
				// because we take into account only the ugnot transfer messages (MsgSend)
				// and not other message types (like MsgCall), that can also
				// initialize accounts with some balances. Because of this,
				// we can run into a situation where a message send amount or fee
				// causes an accounts balance to go < 0. In these cases,
				// we initialize the account (it is present in the balance sheet), but
				// with the balance of 0

				from := balances[msgSend.FromAddress].Amount
				to := balances[msgSend.ToAddress].Amount

				to = to.Add(sendAmount)

				if from.IsAllLT(sendAmount) || from.IsAllLT(feeAmount) {
					// Account cannot cover send amount / fee
					// (see message above)
					from = std.NewCoins(std.NewCoin("ugnot", 0))
				}

				if from.IsAllGT(sendAmount) {
					from = from.Sub(sendAmount)
				}

				if from.IsAllGT(feeAmount) {
					from = from.Sub(feeAmount)
				}

				// Set new balance
				balances[msgSend.FromAddress] = gnoland.Balance{
					Address: msgSend.FromAddress,
					Amount:  from,
				}
				balances[msgSend.ToAddress] = gnoland.Balance{
					Address: msgSend.ToAddress,
					Amount:  to,
				}
			}
		}
	}

	// Check for scanning errors
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf(
			"error encountered while reading file, %w",
			err,
		)
	}

	return balances, nil
}

// mapGenesisBalancesFromState extracts the initial account balances from the
// genesis app state
func mapGenesisBalancesFromState(state gnoland.GnoGenesisState) (accountBalances, error) {
	// Construct the initial genesis balance sheet
	genesisBalances := make(accountBalances)

	for _, balance := range state.Balances {
		genesisBalances[balance.Address] = balance
	}

	return genesisBalances, nil
}
