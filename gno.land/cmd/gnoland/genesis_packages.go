package main

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type packagesCfg struct {
	commonCfg
}

// newPackagesCmd creates the genesis packages subcommand
func newPackagesCmd(io commands.IO) *commands.Command {
	cfg := &packagesCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "packages",
			ShortUsage: "packages <subcommand> [flags]",
			ShortHelp:  "manages genesis.json packages",
			LongHelp:   "Manipulates the initial genesis.json packages",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newPackagesAddCmd(cfg, io),
		newPackagesClearCmd(cfg, io),
		newPackagesListCmd(cfg, io),
	)

	return cmd
}

func (c *packagesCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonCfg.RegisterFlags(fs)
}
