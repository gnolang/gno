package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gnolang/gno/misc/loop/cmd/cfg"
	"github.com/gnolang/gno/misc/loop/cmd/portalloop"

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

func execServe(ctx context.Context, cfg_ *cfg.CmdCfg) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	portalLoopHandler, err := portalloop.NewPortalLoopHandler(cfg_, logger)
	if err != nil {
		return err
	}

	for {
		if err := portalloop.RunPortalLoop(ctx, *portalLoopHandler, false); err != nil {
			return fmt.Errorf("unable to run loop: %w", err)
		}

		logger.Info("Waiting 3 min before new loop attempt")

		// Wait for a new round
		select {
		case <-time.After(3 * time.Minute):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
