package main

import (
	"context"
	"flag"
	"os"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	io := commands.NewDefaultIO()
	cmd := newRootCmd(io)

	cmd.Execute(context.Background(), os.Args[1:])
}

func newRootCmd(io commands.IO) *commands.Command {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			LongHelp:   "Gno Genesis manipulation suite",
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
