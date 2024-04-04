package main

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type validatorCfg struct {
	commonCfg

	address string
}

// newValidatorCmd creates the genesis validator subcommand
func newValidatorCmd(io commands.IO) *commands.Command {
	cfg := &validatorCfg{
		commonCfg: commonCfg{},
	}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "validator",
			ShortUsage: "validator <subcommand> [flags]",
			ShortHelp:  "validator set management in genesis.json",
			LongHelp:   "Manipulates the genesis.json validator set",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newValidatorAddCmd(cfg, io),
		newValidatorRemoveCmd(cfg, io),
	)

	return cmd
}

func (c *validatorCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonCfg.RegisterFlags(fs)

	fs.StringVar(
		&c.address,
		"address",
		"",
		"the output path for the genesis.json",
	)
}
