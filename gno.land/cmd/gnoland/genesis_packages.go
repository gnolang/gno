package main

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	test1      = crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	defaultFee = std.NewFee(50000, std.MustParseCoin("1000000ugnot"))
)

const msgAddPkg = "add_package"

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
		newPackagesClearCmd(cfg, io),
		newPackagesListCmd(cfg, io),
		newPackagesGetCmd(cfg, io),
		newPackagesDelCmd(cfg, io),
	)

	return cmd
}

func (c *packagesCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonCfg.RegisterFlags(fs)
}
