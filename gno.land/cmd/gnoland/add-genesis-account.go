package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type addGenesisAccountCfg struct {
	rootDir string

	genesisBalancesFile string
}

func newAddGenesisAccountCmd(io *commands.IO) *commands.Command {
	cfg := &addGenesisAccountCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "add-genesis-account",
			ShortUsage: "add-genesis-account [address] [coin] [flags]",
			ShortHelp:  "Add a genesis account to genesis.json",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execAddGenesisAccount(cfg, args, io)
		},
	)
}

func (c *addGenesisAccountCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"testdir",
		"directory for config and data",
	)

	fs.StringVar(
		&c.genesisBalancesFile,
		"from-file",
		"",
		"parse genesis balances from file",
	)
}

func execAddGenesisAccount(c *addGenesisAccountCfg, args []string, io *commands.IO) error {
	rootDir := c.rootDir
	genesisFile := rootDir + "/config/genesis.json"

	gen, err := bft.GenesisDocFromFile(genesisFile)
	if err != nil {
		return err
	}

	appState, ok := gen.AppState.(gnoland.GnoGenesisState)
	if !ok {
		panic("failed to parse genesis state")
	}

	var balances []string

	if c.genesisBalancesFile != "" {
		balances = loadGenesisBalances(c.genesisBalancesFile)
	} else {
		balances = append(balances, args[0]+"="+args[1])
	}

	for _, balance := range balances {
		for _, line := range appState.Balances {
			if strings.HasPrefix(line, balance) {
				return fmt.Errorf("cannot add account at existing address: %s", balance)
			}
		}
		appState.Balances = append(appState.Balances, balance)
	}

	gen.AppState = appState

	return gen.SaveAs(genesisFile)
}
