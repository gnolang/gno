package main

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type txsCfg struct {
	genesisPath string
}

// newTxsCmd creates the genesis txs subcommand
func newTxsCmd(io *commands.IO) *commands.Command {
	cfg := &txsCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "txs",
			ShortUsage: "txs <subcommand> [flags]",
			LongHelp:   "Manipulates the genesis.json validator set",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newTxsAddCmd(cfg, io),
		// newTxsRemoveCmd(cfg, io),
	)

	return cmd
}

func (c *txsCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.genesisPath,
		"genesis-path",
		"./genesis.json",
		"the path to the genesis.json",
	)
}
