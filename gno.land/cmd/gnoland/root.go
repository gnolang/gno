package main

import (
	"context"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftoml"
)

const flagConfigFlag = "flag-config-path"

func main() {
	cmd := newRootCmd(commands.NewDefaultIO())

	cmd.Execute(context.Background(), os.Args[1:])
}

func newRootCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "starts the gnoland blockchain node",
			Options: []ff.Option{
				ff.WithConfigFileFlag(flagConfigFlag),
				ff.WithConfigFileParser(fftoml.Parser),
			},
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newStartCmd(io),
		newSecretsCmd(io),
		newConfigCmd(io),
	)

	return cmd
}
