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
)

type packagesGetCfg struct {
	rootCfg *packagesCfg
}

// newPackagesGetCmd creates the genesis packages list subcommand
func newPackagesGetCmd(rootCfg *packagesCfg, io commands.IO) *commands.Command {
	cfg := &packagesGetCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "get",
			ShortUsage: "packages get [flags] <pkgpath> [<pkgpath>...]",
			ShortHelp:  "get the addpkg transactions for given package path",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execPackagesGet(cfg, args, io)
		},
	)
}

func (c *packagesGetCfg) RegisterFlags(fs *flag.FlagSet) {}

func execPackagesGet(cfg *packagesGetCfg, args []string, io commands.IO) error {
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
	want := make(map[string]struct{}, len(args))
	for _, arg := range args {
		want[arg] = struct{}{}
	}

	count := 0
	for _, tx := range state.Txs {
		for _, msg := range tx.Msgs {
			if msg.Type() != msgAddPkg {
				continue
			}
			msgAddPkg := msg.(vmm.MsgAddPackage)
			if _, ok := want[msgAddPkg.Package.Path]; !ok {
				continue
			}

			count++

			// Marshal tx
			m, err := amino.MarshalJSON(tx)
			if err != nil {
				return fmt.Errorf("unable to marshal amino JSON, %w", err)
			}

			// Print marshalled tx
			io.Printfln(string(m))
			io.Println()

			break
		}
	}

	io.Printfln(
		"%d txs found!",
		count,
	)

	return nil
}
