package txs

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var errInvalidPackageDir = errors.New("invalid package directory")

var genesisDeployFee = std.NewFee(50000, std.MustParseCoin(ugnot.ValueString(1000000)))

type addPkgCfg struct {
	txsCfg           *txsCfg
	deployerMnemonic string
}

func (c *addPkgCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.deployerMnemonic,
		"deployer-mnemonic",
		"source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast", // test1
		"The mnemonic of the wallet that will create packages on the transaction genesis",
	)
}

// newTxsAddPackagesCmd creates the genesis txs add packages subcommand
func newTxsAddPackagesCmd(txsCfg *txsCfg, io commands.IO) *commands.Command {
	cfg := &addPkgCfg{
		txsCfg: txsCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "packages",
			ShortUsage: "txs add packages <package-path ...>",
			ShortHelp:  "imports transactions from the given packages into the genesis.json",
			LongHelp:   "Imports the transactions from a given package directory recursively to the genesis.json",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execTxsAddPackages(cfg, io, args)
		},
	)
}

func execTxsAddPackages(
	cfg *addPkgCfg,
	io commands.IO,
	args []string,
) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.txsCfg.GenesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Make sure the package dir is set
	if len(args) == 0 {
		return errInvalidPackageDir
	}

	parsedTxs := make([]gnoland.TxWithMetadata, 0)
	for _, path := range args {
		// Generate transactions from the packages (recursively)
		txs, err := gnoland.LoadPackagesFromDir(path, cfg.deployerMnemonic, genesisDeployFee)
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
	if err := genesis.SaveAs(cfg.txsCfg.GenesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	io.Printfln(
		"Saved %d transactions to genesis.json",
		len(parsedTxs),
	)

	return nil
}
