package balances

import (
	"context"
	"fmt"
	"os"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// newBalancesExportCmd creates the genesis balances export subcommand
func newBalancesExportCmd(balancesCfg *balancesCfg, io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "export",
			ShortUsage: "balances export [flags] <output-path>",
			ShortHelp:  "exports the balances from the genesis.json",
			LongHelp:   "Exports the balances from the genesis.json to an output file",
		},
		commands.NewEmptyConfig(),
		func(_ context.Context, args []string) error {
			return execBalancesExport(balancesCfg, io, args)
		},
	)
}

func execBalancesExport(cfg *balancesCfg, io commands.IO, args []string) error {
	// Load the genesis
	genesis, loadErr := types.GenesisDocFromFile(cfg.GenesisPath)
	if loadErr != nil {
		return fmt.Errorf("unable to load genesis, %w", loadErr)
	}

	// Load the genesis state
	if genesis.AppState == nil {
		return common.ErrAppStateNotSet
	}

	state := genesis.AppState.(gnoland.GnoGenesisState)
	if len(state.Balances) == 0 {
		io.Println("No genesis balances to export")

		return nil
	}

	// Make sure the output file path is specified
	if len(args) == 0 {
		return common.ErrNoOutputFile
	}

	// Open output file
	outputFile, err := os.OpenFile(
		args[0],
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0o755,
	)
	if err != nil {
		return fmt.Errorf("unable to create output file, %w", err)
	}
	defer outputFile.Close()

	// Save the balances
	for _, balance := range state.Balances {
		if _, err = fmt.Fprintf(outputFile, "%s\n", balance); err != nil {
			return fmt.Errorf("unable to write to output, %w", err)
		}
	}

	io.Printfln(
		"Exported %d balances",
		len(state.Balances),
	)

	return nil
}
