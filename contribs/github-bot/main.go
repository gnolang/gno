package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "github-bot <subcommand> [flags]",
			LongHelp:   "Bot that allows for advanced management of GitHub pull requests.",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newCheckCmd(),
		newMatrixCmd(),
	)

	cmd.Execute(context.Background(), os.Args[1:])
}
