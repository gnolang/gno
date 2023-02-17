// Dedicated to my love, Lexi.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/cmd/common"
	"github.com/gnolang/gno/pkgs/commands"
)

const (
	mnemonicEntropySize = 256
)

type baseCfg struct {
	common.BaseOptions
}

func main() {
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

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v", err)

		os.Exit(1)
	}
}

func (c *baseCfg) RegisterFlags(fs *flag.FlagSet) {
	// Base options
	fs.StringVar(
		&c.Home,
		"home",
		common.DefaultBaseOptions.Home,
		"home directory",
	)

	fs.StringVar(
		&c.Remote,
		"remote",
		common.DefaultBaseOptions.Remote,
		"remote node URL",
	)

	fs.BoolVar(
		&c.Quiet,
		"quiet",
		common.DefaultBaseOptions.Quiet,
		"suppress output during execution",
	)

	fs.BoolVar(
		&c.InsecurePasswordStdin,
		"insecure-password-stdin",
		common.DefaultBaseOptions.Quiet,
		"WARNING! take password from stdin",
	)
}
