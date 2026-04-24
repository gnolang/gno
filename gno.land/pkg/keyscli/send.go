package keyscli

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
)

func newMakeSendCmd(rootCfg *client.MakeTxCfg, io commands.IO) *commands.Command {
	cfg := &client.MakeSendCfg{
		RootCfg: rootCfg,
	}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "send",
			ShortUsage: "send [flags] <key-name or address>",
			ShortHelp:  "sends native currency",
		},
		cfg,
		func(_ context.Context, args []string) error {
			if canPrompt(cfg.RootCfg, io) {
				return execMakeSendInteractive(cfg, args, io, false)
			}
			return client.ExecMakeSend(cfg, args, io)
		},
	)
}
