package cmd

import (
	"context"
	"loop/cmd/cfg"
	"loop/cmd/portalloop"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"go.uber.org/zap"
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
	return ExecAll(
		ctx,
		cfg_,
		func(ctx context.Context, cfg *cfg.CmdCfg, portalLoopHandler *portalloop.PortalLoopHandler) error {
			var wg sync.WaitGroup
			logger, _ := zap.NewProduction()

			for {
				var err_ error
				wg.Add(1)
				go func() {
					defer wg.Done()
					err_ = portalloop.RunPortalLoop(ctx, *portalLoopHandler, false)
					if err_ != nil {
						logger.Error("Portal Loop Run ended with error", zap.Error(err_))
						return
					}
					// Wait for a new round
					logger.Info("Waiting 3 min before new loop attempt")
					time.Sleep(3 * time.Minute)
				}()
				wg.Wait()
				if err_ != nil {
					return err_
				}
			}
		},
	)
}
