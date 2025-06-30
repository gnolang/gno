package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gnolang/gno/misc/loop/cmd/cfg"
	"github.com/gnolang/gno/misc/loop/cmd/portalloop"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func NewBackupCmd(_ commands.IO) *commands.Command {
	cfg := &cfg.CmdCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "backup",
			ShortUsage: "backup [flags]",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			return execBackup(ctx, cfg)
		},
	)
}

func execBackup(ctx context.Context, cfg_ *cfg.CmdCfg) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	portalLoopHandler, err := portalloop.NewPortalLoopHandler(cfg_, logger)
	if err != nil {
		return err
	}
	err = portalloop.RunPortalLoop(ctx, *portalLoopHandler, false)
	if err != nil {
		return err
	}

	return portalLoopHandler.BackupTXs(ctx)
}
