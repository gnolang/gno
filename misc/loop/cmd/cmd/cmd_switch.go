package cmd

import (
	"context"
	"loop/cmd/cfg"
	"loop/cmd/portalloop"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func NewSwitchCmd(_ commands.IO) *commands.Command {
	cfg := &cfg.CmdCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "switch",
			ShortUsage: "switch [flags]",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execSwitch(ctx, cfg)
		},
	)
}

func execSwitch(ctx context.Context, cfg_ *cfg.CmdCfg) error {
	return ExecAll(
		ctx,
		cfg_,
		func(ctx context.Context, cfg *cfg.CmdCfg, portalLoopHandler *portalloop.PortalLoopHandler) error {
			return portalloop.StartPortalLoop(ctx, *portalLoopHandler, true)

		},
	)
}
