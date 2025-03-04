package cmd

import (
	"context"
	"loop/cmd/cfg"
	"loop/cmd/portalloop"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func NewServeCmd(_ commands.IO) *commands.Command {
	cfg := &cfg.CmdCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "serve",
			ShortUsage: "serve [flags]",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execServe(ctx, cfg)
		},
	)
}

func execServe(ctx context.Context, cfg *cfg.CmdCfg) error {
	portalLoopHandler, err := portalloop.NewPortalLoopHandler(cfg)
	if err != nil {
		return err
	}

	return portalloop.StartPortalLoop(ctx, *portalLoopHandler, false)
}
