package main

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func newGenesisCmd(io commands.IO) *commands.Command {
	cfg := &commonCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "genesis",
			ShortUsage: "genesis <subcommand> [flags] [<arg>...]",
			ShortHelp:  "gno genesis manipulation suite",
			LongHelp:   "Gno genesis.json manipulation suite, for managing genesis parameters",
		},
		cfg,
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
	rootCfg
}

func (c *commonCfg) RegisterFlags(fs *flag.FlagSet) {
	c.rootCfg.RegisterFlags(fs)

	if genesisFile := fs.Lookup("genesis"); genesisFile == nil {
		fs.StringVar(
			&c.homeDir.genesisFile,
			"genesis",
			"",
			"the path to the genesis.json",
		)
	} else {
		c.homeDir.genesisFile = genesisFile.Value.(flag.Getter).Get().(string)
	}
}
