package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/node"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

func newStartCmd(io *commands.IO) *commands.Command {
	cfg := &startCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "start",
			ShortUsage: "start [flags]",
			ShortHelp:  "Run the full node",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execStart(cfg, args, io)
		},
	)
}

type startCfg struct {
	init initCfg
}

func (s startCfg)RegisterFlags(fs *flag.FlagSet) {
	s.init.RegisterFlags(fs)
}

func execStart(c *startCfg, args []string, io *commands.IO) error {
	rootDir := c.init.rootDir
	logger := log.NewTMLogger(log.NewSyncWriter(io.Out))

	// lazy init
	if !config.Exists(rootDir) {
		err := execInit(&c.init, []string{}, io) // XXX: create an helper instead of calling the cmd?
		if err != nil {
			return fmt.Errorf("lazy init: %w", err)
		}
	}

	// load (existing) config
	cfg := config.LoadOrMakeDefaultConfig(rootDir)
	gnoApp, err := gnoland.NewApp(rootDir, c.init.skipFailingGenesisTxs, logger, c.init.genesisMaxVMCycles)
	if err != nil {
		return fmt.Errorf("error in creating new app: %w", err)
	}
	cfg.LocalApp = gnoApp

	// create node
	gnoNode, err := node.DefaultNewNode(cfg, logger)
	if err != nil {
		return fmt.Errorf("error in creating node: %w", err)
	}

	// start node
	fmt.Fprintln(io.Err, "Starting Node...")
	if err := gnoNode.Start(); err != nil {
		return fmt.Errorf("error in start node: %w", err)
	}

	// run forever
	osm.TrapSignal(func() {
		if gnoNode.IsRunning() {
			_ = gnoNode.Stop()
		}
	})
	select {} // run forever
}
