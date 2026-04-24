package keyscli

import (
	"context"
	"flag"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
)

func NewMakeTxCmd(rootCfg *client.BaseCfg, io commands.IO) *commands.Command {
	cfg := &client.MakeTxCfg{
		RootCfg: rootCfg,
	}

	maketxExec := func(_ context.Context, args []string) error {
		if commands.IsIOInteractive(io) && !cfg.NoInteractive {
			return execMakeTxInteractive(cfg, args, io)
		}
		commands.HelpExec(context.Background(), args)
		return flag.ErrHelp
	}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "maketx",
			ShortUsage: "<subcommand> [flags] [<arg>...]",
			ShortHelp:  "composes a tx document to sign",
		},
		cfg,
		maketxExec,
	)

	cmd.AddSubCommands(
		newMakeSendCmd(cfg, io),

		NewMakeAddPkgCmd(cfg, io),
		NewMakeCallCmd(cfg, io),
		NewMakeRunCmd(cfg, io),
	)

	return cmd
}
