package main

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type balancesCfg struct {
	commonCfg
}

// newBalancesCmd creates the genesis balances subcommand
func newBalancesCmd(io commands.IO) *commands.Command {
	cfg := &balancesCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "balances",
			ShortUsage: "balances <subcommand> [flags]",
			ShortHelp:  "manages genesis.json account balances",
			LongHelp:   "Manipulates the initial genesis.json account balances (pre-mines)",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newBalancesAddCmd(cfg, io),
		newBalancesRemoveCmd(cfg, io),
		newBalancesExportCmd(cfg, io),
	)

	return cmd
}

func (c *balancesCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonCfg.RegisterFlags(fs)
}
