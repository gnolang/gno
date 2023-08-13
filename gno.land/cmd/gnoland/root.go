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
	cmd := newRootCmd(io)

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

func newRootCmd(io *commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "Manages the gnoland blockchain node",
			Options: []ff.Option{
				ff.WithConfigFileFlag("config"),
				ff.WithConfigFileParser(fftoml.Parser),
			},
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newInitCmd(io),
		newStartCmd(io),
	)

	return cmd
}
