package main

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

// newConfigCmd creates the new config root command
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
		newConfigBaseCmd(io),
		newConfigRPCCmd(io),
		newConfigP2PCmd(io),
		newConfigConsensusCmd(io),
		// newConfigEventStoreCmd(io),
	)

	return cmd
}

type commonEditCfg struct {
	configPath string
}

func (c *commonEditCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.configPath,
		"config-path",
		"./config.toml",
		"the path to the node's config.toml",
	)
}
