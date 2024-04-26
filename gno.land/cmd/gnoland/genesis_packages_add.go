package main

import (
	"context"
	"flag"
	"fmt"

	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type packagesAddCfg struct {
	rootCfg *packagesCfg
}

// newPackagesAddCmd creates the genesis packages add subcommand
func newPackagesAddCmd(rootCfg *packagesCfg, io commands.IO) *commands.Command {
	cfg := &packagesAddCfg{
		rootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "add",
			ShortUsage: "packages add [flags] <path> [<path>...]",
			LongHelp:   "Adds new package(s) to the genesis.json",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execPackagesAdd(ctx, cfg, args, io)
		},
	)
}

func (c *packagesAddCfg) RegisterFlags(fs *flag.FlagSet) {}

func execPackagesAdd(ctx context.Context, cfg *packagesAddCfg, args []string, io commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	// Load the genesis
	genesis, err := types.GenesisDocFromFile(cfg.rootCfg.genesisPath)
	if err != nil {
		return fmt.Errorf("unable to load genesis, %w", err)
	}

	test1 := crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // TODO(hariom): flags?
	defaultFee := std.NewFee(50000, std.MustParseCoin("1000000ugnot"))                // TODO(hariom): flags?
	txs, err := gnoland.LoadPackagesFromDirs(args, test1, defaultFee, nil)
	if err != nil {
		return fmt.Errorf("unable to load packages: %w", err)
	}

	state := genesis.AppState.(gnoland.GnoGenesisState)
	state.Txs = append(state.Txs, txs...)

	// Save the txs
	genesis.AppState = state

	// Save the updated genesis
	if err := genesis.SaveAs(cfg.rootCfg.genesisPath); err != nil {
		return fmt.Errorf("unable to save genesis.json, %w", err)
	}

	// Print packages and files
	for _, tx := range txs {
		for _, msg := range tx.Msgs {
			msgAddPkg := msg.(vmm.MsgAddPackage)
			io.Println(msgAddPkg.Package.Path)
			for _, file := range msgAddPkg.Package.Files {
				io.Printfln("\t- %s", file.Name)
			}
		}
	}

	io.Println()

	io.Printfln(
		"%d txs added!",
		len(txs),
	)

	return nil
}
