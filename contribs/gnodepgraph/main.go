package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	/*cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "depgraph [flags] [<arg>...]",
			LongHelp:   "Generates the dependency graph for gno.land",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newDepGraphCmd(commands.NewDefaultIO()),
	)*/

	cmd := newDepGraphCmd(commands.NewDefaultIO())

	cmd.Execute(context.Background(), os.Args[1:])
}
