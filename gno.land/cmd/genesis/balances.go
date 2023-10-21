package main

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type balancesCfg struct {
	genesisPath string
}

// newBalancesCmd creates the genesis balances subcommand
func newBalancesCmd(io *commands.IO) *commands.Command {
	cfg := &balancesCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "balances",
			ShortUsage: "balances <subcommand> [flags]",
			LongHelp:   "Manipulates the initial genesis.json account balances (pre-mines)",
			ShortHelp:  "Manages genesis.json account balances",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newBalancesAddCmd(cfg, io),
		// newBalancesRemoveCmd(cfg, io)
	)

	return cmd
}

func (c *balancesCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.genesisPath,
		"genesis-path",
		"./genesis.json",
		"the path to the genesis.json",
	)
}
