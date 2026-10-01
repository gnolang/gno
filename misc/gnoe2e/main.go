package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	cmd := newRootCmd(commands.NewDefaultIO())
	cmd.Execute(context.Background(), os.Args[1:])
}

func newRootCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "gnoe2e <subcommand> [flags] [<arg>...]",
			ShortHelp:  "E2E multi-node testing tool",
			LongHelp:   "Runs txtar scenarios against local multi-validator gnoland clusters built from the enclosing checkout",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newRunCmd(io),
		newServeCmd(io),
		newDefaultsCmd(io),
	)

	return cmd
}
