package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"

	_ "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
)

var (
	balanceRegex = regexp.MustCompile(`^(\w+)=(\d+)ugnot$`)
	amountRegex  = regexp.MustCompile(`^(\d+)ugnot$`)
)

var (
	errNoBalanceSource        = errors.New("at least one balance source must be set")
	errMultipleBalanceSources = errors.New("only one mode can be set at a time")
	errBalanceParsingAborted  = errors.New("balance parsing aborted")
	errInvalidBalanceFormat   = errors.New("invalid balance format encountered")
	errInvalidAddress         = errors.New("invalid address encountered")
	errInvalidAmount          = errors.New("invalid amount encountered")
)

type balancesAddCfg struct {
	rootCfg *balancesCfg

	balanceSheet  string
	singleEntries commands.StringArr
	parseExport   string
}

// newBalancesAddCmd creates the genesis balances add subcommand
func newBalancesAddCmd(rootCfg *balancesCfg, io *commands.IO) *commands.Command {
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

func execBalancesAdd(ctx context.Context, cfg *balancesAddCfg, io *commands.IO) error {
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

	if err := validateSetModes(
		[]bool{
			singleEntriesSet,
			balanceSheetSet,
			txFileSet,
		},
	); err != nil {
		return fmt.Errorf("invalid modes set, %w", err)
	}

	var (
		balances accountBalances
		err      error
	)

	// Get the balance sheet from the source
	switch {
	case singleEntriesSet:
		balances, err = getBalancesFromEntries(cfg.singleEntries)
	case balanceSheetSet:
		// Open the balance sheet
		file, loadErr := os.Open(cfg.balanceSheet)
		if loadErr != nil {
			return fmt.Errorf("unable to open balance sheet, %w", loadErr)
		}

		balances, err = getBalancesFromSheet(file)
	default:
		// Open the transactions file
		file, loadErr := os.Open(cfg.parseExport)
		if loadErr != nil {
			return fmt.Errorf("unable to open transactions file, %w", loadErr)
		}

		balances, err = getBalancesFromTransactions(ctx, file)
	}

	if err != nil {
		return fmt.Errorf("unable to get balances, %w", err)
	}

	// Initialize genesis app state if it is not initialized already
	if genesis.AppState == nil {
		genesis.AppState = gnoland.GnoGenesisState{}
	}

	// Construct the initial genesis balance sheet
	state := genesis.AppState.(gnoland.GnoGenesisState)
	genesisBalances, err := extractGenesisBalances(state)
	if err != nil {
		return err
	}

	// Merge the two balance sheets, with the input
	// having precedence over the genesis transactions
	balances.leftMerge(genesisBalances)

	// Save the balances
	state.Balances = balances.toList()
	genesis.AppState = state

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.rootCfg.genesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	io.Printfln(
		"%d pre-mines saved",
		len(balances),
	)

	io.Println()

	for address, balance := range balances {
		io.Printfln("%s:%dugnot", address.String(), balance)
	}

	return nil
}

// validateSetModes validates a good mode was
// set for the balance addition
func validateSetModes(modes []bool) error {
	anySet := false

	for _, mode := range modes {
		if !mode {
			continue
		}

		if anySet {
			return errMultipleBalanceSources
		}

		anySet = true
	}

	if !anySet {
		return errNoBalanceSource
	}

	return nil
}

// getBalancesFromEntries extracts the balance entries
// from the array of balance
func getBalancesFromEntries(entries []string) (accountBalances, error) {
	balances := make(accountBalances)

	for _, entry := range entries {
		accountBalance, err := getBalanceFromEntry(entry)
		if err != nil {
			return nil, fmt.Errorf("unable to extract balance data, %w", err)
		}

		balances[accountBalance.address] = accountBalance.amount
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

		accountBalance, err := getBalanceFromEntry(entry)
		if err != nil {
			return nil, fmt.Errorf("unable to extract balance data, %w", err)
		}

		balances[accountBalance.address] = accountBalance.amount
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
func getBalancesFromTransactions(ctx context.Context, reader io.Reader) (accountBalances, error) {
	balances := make(accountBalances)

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, errBalanceParsingAborted
		default:
			// Parse the amino JSON
			var tx std.Tx

			if err := amino.UnmarshalJSON(scanner.Bytes(), &tx); err != nil {
				return nil, fmt.Errorf(
					"unable to unmarshal amino JSON, %w",
					err,
				)
			}

			feeAmount, err := getAmountFromEntry(tx.Fee.GasFee.String())
			if err != nil {
				return nil, fmt.Errorf(
					"invalid gas fee amount, %s",
					tx.Fee.GasFee.String(),
				)
			}

			for _, msg := range tx.Msgs {
				if msg.Type() != "send" {
					continue
				}

				msgSend := msg.(bank.MsgSend)

				sendAmount, err := getAmountFromEntry(msgSend.Amount.String())
				if err != nil {
					return nil, fmt.Errorf(
						"%s, %s",
						"invalid send amount",
						msgSend.Amount.String(),
					)
				}

				// This way of determining final account balances is not really valid,
				// because we take into account only the ugnot transfer messages (MsgSend)
				// and not other message types (like MsgCall), that can also
				// initialize accounts with some balances. Because of this,
				// we can run into a situation where a message send amount or fee
				// causes an accounts balance to go < 0. In these cases,
				// we initialize the account (it is present in the balance sheet), but
				// with the balance of 0
				from := balances[msgSend.FromAddress]
				to := balances[msgSend.ToAddress]

				to += sendAmount

				if from < sendAmount || from < feeAmount {
					// Account cannot cover send amount / fee
					// (see message above)
					from = 0
				}

				if from > sendAmount {
					from -= sendAmount
				}

				if from > feeAmount {
					from -= feeAmount
				}

				balances[msgSend.FromAddress] = from
				balances[msgSend.ToAddress] = to
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

// getAmountFromEntry
func getAmountFromEntry(entry string) (int64, error) {
	matches := amountRegex.FindStringSubmatch(entry)

	// Check if there is a match
	if len(matches) != 2 {
		return 0, fmt.Errorf(
			"invalid amount, %s",
			entry,
		)
	}

	amount, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount, %s", matches[1])
	}

	return amount, nil
}

// getBalanceFromEntry extracts the account balance information
// from a single line in the form of: <address>=<amount>ugnot
func getBalanceFromEntry(entry string) (*accountBalance, error) {
	matches := balanceRegex.FindStringSubmatch(entry)
	if len(matches) != 3 {
		return nil, fmt.Errorf("%w, %s", errInvalidBalanceFormat, entry)
	}

	// Validate the address
	address, err := crypto.AddressFromString(matches[1])
	if err != nil {
		return nil, fmt.Errorf("%w, %w", errInvalidAddress, err)
	}

	// Validate the amount
	amount, err := strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w, %w", errInvalidAmount, err)
	}

	return &accountBalance{
		address: address,
		amount:  amount,
	}, nil
}

// extractGenesisBalances extracts the initial account balances from the
// genesis app state
func extractGenesisBalances(state gnoland.GnoGenesisState) (accountBalances, error) {
	// Construct the initial genesis balance sheet
	genesisBalances := make(accountBalances)

	for _, entry := range state.Balances {
		accountBalance, err := getBalanceFromEntry(entry)
		if err != nil {
			return nil, fmt.Errorf("invalid genesis balance entry, %w", err)
		}

		genesisBalances[accountBalance.address] = accountBalance.amount
	}

	return genesisBalances, nil
}
