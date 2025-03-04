package cmd

import (
	"context"
	"loop/cmd/cfg"
	"loop/cmd/portalloop"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/sirupsen/logrus"
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
	var wg sync.WaitGroup
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

	for {
		wg.Add(1)
		go func() {
			defer wg.Done()
			portalloop.StartPortalLoop(ctx, *portalLoopHandler, false)
			// Wait for a new round
			logger.Info("Waiting 3 min before new loop attempt")
			time.Sleep(3 * time.Minute)
		}()
		wg.Wait()
	}
}
