package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gnolang/gno/pkgs/commands"
)

func main() {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "Starts the fund faucet that can be used by users",
		},
		nil,
		func(_ context.Context, _ []string) error {
			return commands.HelpExec()
		},
	)

	cmd.AddSubCommands(
		newServeCmd(),
	)

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v", err)

		os.Exit(1)
	}
}
