package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftoml"
)

func main() {
	io := commands.NewDefaultIO()

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "Starts the gnoland blockchain node",
			Options: []ff.Option{
				ff.WithConfigFileFlag("config"),
				ff.WithConfigFileParser(fftoml.Parser),
			},
		},
		commands.NewEmptyConfig(),

		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newServerCmd(io),
	)

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)

		os.Exit(1)
	}
}
