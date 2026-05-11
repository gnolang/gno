package keyscli

import (
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
)

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
