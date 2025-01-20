package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/contribs/gnokms/internal/common"
	check "github.com/gnolang/gno/contribs/gnokms/internal/gnokey"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	var (
		io          = commands.NewDefaultIO()
		commonFlags = &common.Flags{}
	)

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "Gno Key Management System used to securely manage and control cryptographic keys.",
		},
		commonFlags,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		check.NewGnokeyCmd(io, commonFlags),
	)

	cmd.Execute(context.Background(), os.Args[1:])
}
