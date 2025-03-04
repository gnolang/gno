package cmd

import (
	"context"
	"loop/cmd/cfg"
	"loop/cmd/portalloop"

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
	return ExecAll(
		ctx,
		cfg_,
		func(ctx context.Context, cfg *cfg.CmdCfg, portalLoopHandler *portalloop.PortalLoopHandler) error {
			err := portalloop.StartPortalLoop(ctx, *portalLoopHandler, false)
			if err != nil {
				return err
			}

			return portalLoopHandler.BackupTXs(ctx)
		},
	)
}
