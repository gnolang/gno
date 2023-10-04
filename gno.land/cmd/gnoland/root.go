package main

import (
	//	"context"

	"flag"
	"os"

	tmcfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/log"
)

var logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))

type baseCfg struct {
	rootDir  string
	tmConfig tmcfg.Config
}

func (bc *baseCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&bc.rootDir,
		"root-dir",
		"testdir",
		"directory for config and data",
	)
}

func newRootCmd() *commands.Command {
	bc := baseCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "The gnoland blockchain node",
		},
		&bc,
		commands.HelpExec,
	)

	initTmConfig(&bc)
	cmd.AddSubCommands(
		newStartCmd(bc),
		newResetAllCmd(bc),
	)

	return cmd
}

// we relies on the flag option to pass in the root directory before we can identify where
func initTmConfig(bc *baseCfg) error {
	bc.tmConfig = *tmcfg.LoadOrMakeConfigWithOptions(bc.rootDir, func(cfg *tmcfg.Config) {
	})

	return nil
}
