package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type packagesDelCfg struct {
	rootCfg *packagesCfg
}

// newPackagesDelCmd creates the genesis packages list subcommand
func newPackagesDelCmd(rootCfg *packagesCfg, io commands.IO) *commands.Command {
	cfg := &packagesDelCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "del",
			ShortUsage: "packages del [flags] <pkgpath> [<pkgpath>...]",
			ShortHelp:  "removes the given addpkg transactions",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execPackagesDel(cfg, args, io)
		},
	)
}

func (c *packagesDelCfg) RegisterFlags(fs *flag.FlagSet) {}

func execPackagesDel(cfg *packagesDelCfg, args []string, io commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	// Load the genesis
	genesis, err := types.GenesisDocFromFile(cfg.rootCfg.genesisPath)
	if err != nil {
		return fmt.Errorf("unable to load genesis, %w", err)
	}

	state := genesis.AppState.(gnoland.GnoGenesisState)

	// Create map of given args
	toBeRemoved := make(map[string]struct{}, len(args))
	for _, arg := range args {
		toBeRemoved[arg] = struct{}{}
	}

	var txs []std.Tx
	removed := 0
	for _, tx := range state.Txs {
		include := true
		for _, msg := range tx.Msgs {
			if msg.Type() != "add_package" {
				continue
			}
			msgAddPkg := msg.(vmm.MsgAddPackage)
			if _, ok := toBeRemoved[msgAddPkg.Package.Path]; ok {
				include = false
				break
			}
		}
		if include {
			txs = append(txs, tx)
			continue
		}

		// Not Included
		removed++

		// Marshal tx
		m, err := amino.MarshalJSON(tx)
		if err != nil {
			return fmt.Errorf("unable to marshal amino JSON, %w", err)
		}

		// Print marshalled tx
		io.Printfln(string(m))
		io.Println()
	}

	// Save the txs
	state.Txs = txs
	genesis.AppState = state

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.rootCfg.genesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	io.Printfln(
		"%d txs removed!",
		removed,
	)

	return nil
}
