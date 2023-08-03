package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	cmd := newGnocliCmd(commands.NewDefaultIO())

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)

		os.Exit(1)
	}
}

func newGnocliCmd(io *commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "Runs the gno development toolkit",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newModCmd(io),
		newTestCmd(io),
		newLintCmd(io),
		newRunCmd(io),
		newBuildCmd(io),
		newPrecompileCmd(io),
		newCleanCmd(io),
		newReplCmd(),
		newDocCmd(io),
		// fmt -- gofmt
		// graph
		// vendor -- download deps from the chain in vendor/
		// list -- list packages
		// render -- call render()?
		// publish/release
		// generate
		// "vm" -- starts an in-memory chain that can be interacted with?
		// bug -- start a bug report
		// version -- show gnodev, golang versions
	)

	return cmd
}
