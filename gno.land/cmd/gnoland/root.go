package main

import (
	"context"
	"flag"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftoml"
)

func main() {
	cmd := newRootCmd(commands.NewDefaultIO())

	cmd.Execute(context.Background(), os.Args[1:])
}

type rootCfg struct {
	homeDir homeDirectory
}

func (c *rootCfg) RegisterFlags(fs *flag.FlagSet) {
	if home := fs.Lookup("home"); home == nil {
		fs.StringVar(
			&c.homeDir.homeDir,
			"home",
			"./gnoland-data",
			"Directory for config, secrets and data",
		)
	} else {
		c.homeDir.homeDir = home.Value.(flag.Getter).Get().(string)
	}
}

func newRootCmd(io commands.IO) *commands.Command {
	cfg := &rootCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "starts the gnoland blockchain node",
			Options: []ff.Option{
				ff.WithConfigFileParser(fftoml.Parser),
			},
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newStartCmd(io),
		newGenesisCmd(io),
		newSecretsCmd(io),
		newConfigCmd(io),
	)

	return cmd
}
