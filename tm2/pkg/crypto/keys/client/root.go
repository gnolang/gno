// Dedicated to my love, Lexi.
package client

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"

	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftoml"
)

const (
	mnemonicEntropySize = 256
)

type BaseCfg struct {
	BaseOptions
}

func NewRootCmd(io commands.IO) *commands.Command {
	return NewRootCmdWithBaseConfig(io, DefaultBaseOptions)
}

func NewRootCmdWithBaseConfig(io commands.IO, base BaseOptions) *commands.Command {
	cfg := &BaseCfg{
		BaseOptions: base,
	}

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "Manages private keys for the node",
			Options: []ff.Option{
				ff.WithConfigFileFlag("config"),
				ff.WithConfigFileParser(fftoml.Parser),
			},
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		NewAddCmd(cfg, io),
		NewDeleteCmd(cfg, io),
		NewGenerateCmd(cfg, io),
		NewExportCmd(cfg, io),
		NewImportCmd(cfg, io),
		NewListCmd(cfg, io),
		NewSignCmd(cfg, io),
		NewVerifyCmd(cfg, io),
		NewQueryCmd(cfg, io),
		NewBroadcastCmd(cfg, io),
		NewMakeTxCmd(cfg, io),
	)

	return cmd
}

func (c *BaseCfg) RegisterFlags(fs *flag.FlagSet) {
	// Base options
	fs.StringVar(
		&c.Home,
		"home",
		c.Home,
		"home directory",
	)

	fs.StringVar(
		&c.Remote,
		"remote",
		c.Remote,
		"remote node URL",
	)

	fs.BoolVar(
		&c.Quiet,
		"quiet",
		c.Quiet,
		"suppress output during execution",
	)

	fs.BoolVar(
		&c.InsecurePasswordStdin,
		"insecure-password-stdin",
		c.Quiet,
		"WARNING! take password from stdin",
	)

	fs.StringVar(
		&c.Config,
		"config",
		c.Config,
		"config file (optional)",
	)
}
