package cmd

import (
	"context"
	"fmt"
	"loop/cmd/cfg"
	"loop/cmd/portalloop"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/log"
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
	logger := log.NewNoopLogger()
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
