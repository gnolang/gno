package main

import (
	"github.com/gnolang/contribs/gnogenesis/internal/balances"
	"github.com/gnolang/contribs/gnogenesis/internal/generate"
	"github.com/gnolang/contribs/gnogenesis/internal/params"
	"github.com/gnolang/contribs/gnogenesis/internal/txs"
	"github.com/gnolang/contribs/gnogenesis/internal/validator"
	"github.com/gnolang/contribs/gnogenesis/internal/verify"
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
		generate.NewGenerateCmd(io),
		validator.NewValidatorCmd(io),
		verify.NewVerifyCmd(io),
		balances.NewBalancesCmd(io),
		txs.NewTxsCmd(io),
		params.NewParamsCmd(io),
	)

	return cmd
}
