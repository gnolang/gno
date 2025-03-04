package cmd

import (
	"context"
	"loop/cmd/cfg"
	"loop/cmd/portalloop"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

type ExecFn func(ctx context.Context, cfg *cfg.CmdCfg, portalLoopHandler *portalloop.PortalLoopHandler) error

func ExecAll(ctx context.Context, cfg *cfg.CmdCfg, execFn ExecFn) error {
	logger := logrus.New()
	portalLoopHandler, err := portalloop.NewPortalLoopHandler(cfg)
	if err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	//  Wait for termination signal
	go func() {
		sig := <-sigs // Wait for SIGTERM or SIGINT
		logger.Info("Received termination signal, gracefully killing all the existing portal loop instances")
		portalLoopHandler.ProxyRemoveContainers(ctx)

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
