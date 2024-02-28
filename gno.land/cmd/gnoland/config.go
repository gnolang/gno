package main

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type configCfg struct {
	configPath string
}

// newConfigCmd creates the config root command
func newConfigCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "config",
			ShortUsage: "config <subcommand> [flags]",
			ShortHelp:  "Gno config manipulation suite",
			LongHelp:   "Gno config manipulation suite, for editing base and module configurations",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newConfigInitCmd(io),
		newConfigEditCmd(io),
	)

	return cmd
}

func (c *configCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.configPath,
		"config-path",
		"./config.toml",
		"the path for the config.toml",
	)
}
