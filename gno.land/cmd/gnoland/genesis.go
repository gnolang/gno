package main

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func newGenesisCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "genesis",
			ShortUsage: "genesis <subcommand> [flags] [<arg>...]",
			ShortHelp:  "gno genesis manipulation suite",
			LongHelp:   "Gno genesis.json manipulation suite, for managing genesis parameters",
		},
		commands.NewEmptyConfig(),
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newGenerateCmd(io),
		newValidatorCmd(io),
		newVerifyCmd(io),
		newBalancesCmd(io),
		newTxsCmd(io),
	)

	return cmd
}

// commonCfg is the common
// configuration for genesis commands
// that require a genesis.json
type commonCfg struct {
	genesisPath string
}

func (c *commonCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.genesisPath,
		"genesis-path",
		"./genesis.json",
		"the path to the genesis.json",
	)
}
