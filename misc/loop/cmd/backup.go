package main

import (
	"context"

	"github.com/gnolang/gno/misc/loop/cmd/cfg"
	"github.com/gnolang/gno/misc/loop/cmd/portalloop"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/log"
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
	logger := log.NewNoopLogger()
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
