package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/contribs/gnokms/internal/auth"
	"github.com/gnolang/gno/contribs/gnokms/internal/gnokey"
	"github.com/gnolang/gno/contribs/gnokms/internal/ledger"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	io := commands.NewDefaultIO()

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "Gno Key Management System used to securely manage and control node validator signing keys.",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		auth.NewAuthCmd(io),
		gnokey.NewGnokeyCmd(io),
		ledger.NewLedgerCmd(io),
	)

	cmd.Execute(context.Background(), os.Args[1:])
}
