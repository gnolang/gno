package keyscli

import (
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
)

type MakeTxCfg struct {
	RootCfg *client.BaseCfg

	GasWanted int64
	GasFee    string
	Memo      string

	Broadcast bool
	ChainID   string
}

func NewMakeTxCmd(rootCfg *client.BaseCfg, io commands.IO) *commands.Command {
	cfg := &client.MakeTxCfg{
		RootCfg: rootCfg,
	}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "maketx",
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "composes a tx document to sign",
		},
		cfg,
		commands.HelpExec,
	)

	cmd.AddSubCommands(
		client.NewMakeSendCmd(cfg, io),

		// custom commands
		NewMakeAddPkgCmd(cfg, io),
		NewMakeCallCmd(cfg, io),
		NewMakeRunCmd(cfg, io),
	)

	return cmd
}

func (c *MakeTxCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.Int64Var(
		&c.GasWanted,
		"gas-wanted",
		0,
		"gas requested for tx",
	)

	fs.StringVar(
		&c.GasFee,
		"gas-fee",
		"",
		"gas payment fee",
	)

	fs.StringVar(
		&c.Memo,
		"memo",
		"",
		"any descriptive text",
	)

	fs.BoolVar(
		&c.Broadcast,
		"broadcast",
		true,
		"sign and broadcast",
	)

	fs.StringVar(
		&c.ChainID,
		"chainid",
		"dev",
		"chainid to sign for (only useful if --broadcast)",
	)
}
