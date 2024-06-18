package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var errInvalidPackageDir = errors.New("invalid package directory")

var (
	genesisDeployAddress = crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // test1
	genesisDeployFee     = std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
)

// newTxsAddPackagesCmd creates the genesis txs add packages subcommand
func newTxsAddPackagesCmd(txsCfg *txsCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "packages",
			ShortUsage: "txs add packages <package-path ...>",
			ShortHelp:  "imports transactions from the given packages into the genesis.json",
			LongHelp:   "Imports the transactions from a given package directory recursively to the genesis.json",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execTxsAddPackages(txsCfg, io, args)
		},
	)
}

func execTxsAddPackages(
	cfg *txsCfg,
	io commands.IO,
	args []string,
) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.homeDir.GenesisFilePath())
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Make sure the package dir is set
	if len(args) == 0 {
		return errInvalidPackageDir
	}

	parsedTxs := make([]std.Tx, 0)
	for _, path := range args {
		// Generate transactions from the packages (recursively)
		txs, err := gnoland.LoadPackagesFromDir(path, genesisDeployAddress, genesisDeployFee)
		if err != nil {
			return fmt.Errorf("unable to load txs from directory, %w", err)
		}

		parsedTxs = append(parsedTxs, txs...)
	}

	// Save the txs to the genesis.json
	if err := appendGenesisTxs(genesis, parsedTxs); err != nil {
		return fmt.Errorf("unable to append genesis transactions, %w", err)
	}

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.homeDir.GenesisFilePath()); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	io.Printfln(
		"Saved %d transactions to genesis.json",
		len(parsedTxs),
	)

	return nil
}
