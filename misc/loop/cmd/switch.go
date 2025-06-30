package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gnolang/gno/misc/loop/cmd/cfg"
	"github.com/gnolang/gno/misc/loop/cmd/portalloop"

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
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	portalLoopHandler, err := portalloop.NewPortalLoopHandler(cfg_, logger)
	if err != nil {
		return err
	}
	return portalloop.RunPortalLoop(ctx, *portalLoopHandler, true)
}
