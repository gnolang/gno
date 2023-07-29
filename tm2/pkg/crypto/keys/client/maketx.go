package client

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type makeTxCfg struct {
	rootCfg *baseCfg

	gasWanted int64
	gasFee    string
	memo      string

	broadcast bool
	chainID   string
}

func newMakeTxCmd(rootCfg *baseCfg, io *commands.IO) *commands.Command {
	cfg := &makeTxCfg{
		rootCfg: rootCfg,
	}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "maketx",
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "Composes a tx document to sign",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		newAddPkgCmd(cfg, io),
		newSendCmd(cfg, io),
		newCallCmd(cfg, io),
		// newExecCmd(cfg, io),
	)

	return cmd
}

func (c *makeTxCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.Int64Var(
		&c.gasWanted,
		"gas-wanted",
		0,
		"gas requested for tx",
	)

	fs.StringVar(
		&c.gasFee,
		"gas-fee",
		"",
		"gas payment fee",
	)

	fs.StringVar(
		&c.memo,
		"memo",
		"",
		"any descriptive text",
	)

	fs.BoolVar(
		&c.broadcast,
		"broadcast",
		false,
		"sign and broadcast",
	)

	fs.StringVar(
		&c.chainID,
		"chainid",
		"dev",
		"chainid to sign for (only useful if --broadcast)",
	)
}
