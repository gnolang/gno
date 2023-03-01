// Dedicated to my love, Lexi.
package client

import (
	"flag"

	"github.com/gnolang/gno/pkgs/commands"
)

const (
	mnemonicEntropySize = 256
)

type baseCfg struct {
	BaseOptions
}

func NewRootCmd() *commands.Command {
	cfg := &baseCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "Manages private keys for the node",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newAddCmd(cfg),
		newDeleteCmd(cfg),
		newGenerateCmd(cfg),
		newExportCmd(cfg),
		newImportCmd(cfg),
		newListCmd(cfg),
		newSignCmd(cfg),
		newVerifyCmd(cfg),
		newQueryCmd(cfg),
		newBroadcastCmd(cfg),
		newMakeTxCmd(cfg),
	)

	return cmd
}

func (c *baseCfg) RegisterFlags(fs *flag.FlagSet) {
	// Base options
	fs.StringVar(
		&c.Home,
		"home",
		DefaultBaseOptions.Home,
		"home directory",
	)

	fs.StringVar(
		&c.Remote,
		"remote",
		DefaultBaseOptions.Remote,
		"remote node URL",
	)

	fs.BoolVar(
		&c.Quiet,
		"quiet",
		DefaultBaseOptions.Quiet,
		"suppress output during execution",
	)

	fs.BoolVar(
		&c.InsecurePasswordStdin,
		"insecure-password-stdin",
		DefaultBaseOptions.Quiet,
		"WARNING! take password from stdin",
	)
}
