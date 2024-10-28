package main

import (
	"github.com/gnolang/contribs/gnogenesis/internal/balances"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func newGenesisCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "gno genesis manipulation suite",
			LongHelp:   "Gno genesis.json manipulation suite, for managing genesis parameters",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		NewGenerateCmd(io),
		NewValidatorCmd(io),
		NewVerifyCmd(io),
		balances.NewBalancesCmd(io),
		NewTxsCmd(io),
	)

	return cmd
}
