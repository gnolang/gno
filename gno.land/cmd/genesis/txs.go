package main

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type txsCfg struct {
	commonCfg
}

// newTxsCmd creates the genesis txs subcommand
func newTxsCmd(io commands.IO) *commands.Command {
	cfg := &txsCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "txs",
			ShortUsage: "txs <subcommand> [flags]",
			ShortHelp:  "manages the initial genesis transactions",
			LongHelp:   "Manages genesis transactions through input files",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newTxsAddCmd(cfg, io),
		newTxsRemoveCmd(cfg, io),
		newTxsExportCmd(cfg, io),
	)

	return cmd
}

func (c *txsCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonCfg.RegisterFlags(fs)
}
