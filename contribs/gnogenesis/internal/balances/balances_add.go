package balances

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
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
			ShortHelp:  "adds balances to the genesis.json",
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
		"the path to the balance file containing addresses in the format <address>=<amount>"+ugnot.Denom,
	)

	fs.Var(
		&c.singleEntries,
		"single",
		"the direct balance addition in the format <address>=<amount>"+ugnot.Denom,
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
	genesis, loadErr := types.GenesisDocFromFile(cfg.rootCfg.GenesisPath)
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

	finalBalances := gnoland.NewBalances()

	// Get the balance sheet from the source
	if singleEntriesSet {
		balances, err := gnoland.GetBalancesFromEntries(cfg.singleEntries...)
		if err != nil {
			return fmt.Errorf("unable to get balances from entries, %w", err)
		}

		finalBalances.LeftMerge(balances)
	}

	if balanceSheetSet {
		// Open the balance sheet
		file, loadErr := os.Open(cfg.balanceSheet)
		if loadErr != nil {
			return fmt.Errorf("unable to open balance sheet, %w", loadErr)
		}

		balances, err := gnoland.GetBalancesFromSheet(file)
		if err != nil {
			return fmt.Errorf("unable to get balances from balance sheet, %w", err)
		}

		finalBalances.LeftMerge(balances)
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

		finalBalances.LeftMerge(balances)
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
	finalBalances.LeftMerge(genesisBalances)

	// Save the balances
	sortedBalances := finalBalances.List()

	state.Balances = sortedBalances
	genesis.AppState = state

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.rootCfg.GenesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	for _, balance := range sortedBalances {
		io.Printfln("%s=%s", balance.Address.String(), balance.Amount.String())
	}

	io.Println()

	io.Printfln(
		"%d balances saved",
		len(finalBalances),
	)

	return nil
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
) (gnoland.Balances, error) {
	balances := gnoland.NewBalances()

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
			if feeAmount.AmountOf(ugnot.Denom) <= 0 {
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
				if sendAmount.AmountOf(ugnot.Denom) <= 0 {
					io.ErrPrintfln(
						"invalid send amount encountered: %s",
						msgSend.Amount.String(),
					)
					continue
				}

				// This way of determining final account balances is not really valid,
				// because we take into account only the ugnot transfer messages (MsgSend)
				// and not other message types (like MsgCall), that can also
				// initialize accounts with some gnoland. Because of this,
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
					from = std.NewCoins(std.NewCoin(ugnot.Denom, 0))
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
func mapGenesisBalancesFromState(state gnoland.GnoGenesisState) (gnoland.Balances, error) {
	// Construct the initial genesis balance sheet
	genesisBalances := gnoland.NewBalances()

	for _, balance := range state.Balances {
		genesisBalances[balance.Address] = balance
	}

	return genesisBalances, nil
}
