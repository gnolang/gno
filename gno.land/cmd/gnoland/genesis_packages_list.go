package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type packagesListCfg struct {
	rootCfg *packagesCfg
}

// newPackagesListCmd creates the genesis packages list subcommand
func newPackagesListCmd(rootCfg *packagesCfg, io commands.IO) *commands.Command {
	cfg := &packagesListCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "list",
			ShortUsage: "packages list [flags]",
			LongHelp:   "List all the addpkg transactions",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execPackagesList(ctx, cfg, args, io)
		},
	)
}

func (c *packagesListCfg) RegisterFlags(fs *flag.FlagSet) {}

func execPackagesList(ctx context.Context, cfg *packagesListCfg, args []string, io commands.IO) error {
	if len(args) > 0 {
		return flag.ErrHelp
	}

	// Load the genesis
	genesis, err := types.GenesisDocFromFile(cfg.rootCfg.genesisPath)
	if err != nil {
		return fmt.Errorf("unable to load genesis, %w", err)
	}

	state := genesis.AppState.(gnoland.GnoGenesisState)

	count := 0
	var pkgList []string
	for _, tx := range state.Txs {
		for _, msg := range tx.Msgs {
			if msg.Type() != "add_package" {
				continue
			}

			count++

			msgAddPkg := msg.(vmm.MsgAddPackage)
			pkgList = append(pkgList, msgAddPkg.Package.Path)
		}
	}

	io.Println(strings.Join(pkgList, "\n"))
	io.Println()
	io.Printfln(
		"%d txs!",
		count,
	)

	return nil
}
