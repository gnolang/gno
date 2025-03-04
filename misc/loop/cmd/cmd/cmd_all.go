package cmd

import (
	"context"
	"loop/cmd/cfg"
	"loop/cmd/portalloop"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

type ExecFn func(ctx context.Context, cfg *cfg.CmdCfg, portalLoopHandler *portalloop.PortalLoopHandler) error

func ExecAll(ctx context.Context, cfg *cfg.CmdCfg, execFn ExecFn) error {
	logger, _ := zap.NewProduction()
	portalLoopHandler, err := portalloop.NewPortalLoopHandler(cfg, logger)
	if err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	//  Wait for termination signal
	go func() {
		sig := <-sigs // Wait for SIGTERM or SIGINT
		logger.Info("Received termination signal, gracefully stopping...")
		portalLoopHandler.ProxyRemoveContainers(ctx)
		logger.Info("Killed all the existing portal loop instances.")

		exitCode := 1
		if s, ok := sig.(syscall.Signal); ok {
			exitCode = int(s)
		}
		os.Exit(exitCode)
	}()

	defer func() {
		sigs <- syscall.SIGTERM
	}()
	return execFn(ctx, cfg, portalLoopHandler)
}
