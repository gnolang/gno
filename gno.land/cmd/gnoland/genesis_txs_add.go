package main

import (
	"github.com/gnolang/gno/tm2/pkg/commands"
)

// newTxsAddCmd creates the genesis txs add subcommand
func newTxsAddCmd(txsCfg *txsCfg, io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "add",
			ShortUsage: "txs add <subcommand> [flags] [<arg>...]",
			ShortHelp:  "adds transactions into the genesis.json",
			LongHelp:   "Adds initial transactions to the genesis.json",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newTxsAddSheetCmd(txsCfg, io),
		newTxsAddPackagesCmd(txsCfg, io),
	)

	return cmd
}
