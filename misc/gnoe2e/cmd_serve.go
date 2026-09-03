package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/gnolang/gno/misc/gnoe2e/internal/builder"
	"github.com/gnolang/gno/misc/gnoe2e/internal/cluster"
	"github.com/gnolang/gno/misc/gnoe2e/internal/termlog"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type serveCfg struct {
	cluster.ClusterConfig
	verbose bool
}

func (c *serveCfg) RegisterFlags(fs *flag.FlagSet) {
	c.ClusterConfig.RegisterFlags(fs)
	fs.BoolVar(&c.verbose, "verbose", false, "verbose output")
}

func newServeCmd(_ commands.IO) *commands.Command {
	clusterCfg := cluster.DefaultClusterConfig()
	clusterCfg.Genesis.LoadExamples = false // opt-in via --load-examples
	cfg := &serveCfg{ClusterConfig: clusterCfg}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "serve",
			ShortUsage: "serve [flags]",
			ShortHelp:  "start a local validator cluster and keep it running",
			LongHelp:   "Starts a local gnoland cluster with configurable validators and genesis, prints the RPC address, and blocks until interrupted",
		},
		cfg,
		func(ctx context.Context, _ []string) error {
			ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
			defer cancel()

			return execServe(ctx, cfg)
		},
	)
}

func execServe(ctx context.Context, cfg *serveCfg) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	logger := slog.New(termlog.NewHandler(os.Stderr, cfg.verbose))
	cfg.ClusterConfig.Logger = logger.With("component", "cluster")

	bldr := builder.NewLocalBuilder()
	gnolandBin, err := bldr.Build(ctx, builder.BuildOpts{})
	if err != nil {
		return fmt.Errorf("build gnoland: %w", err)
	}
	cl, err := cluster.StartCluster(ctx, cfg.ClusterConfig, gnolandBin)
	if err != nil {
		return fmt.Errorf("start cluster: %w", err)
	}
	defer cl.Cleanup()

	logger.Info("cluster ready",
		"rpc_addr", cl.RPCAddr,
		"chain_id", cfg.Genesis.ChainID,
		"validators", cfg.NumValidators,
	)
	fmt.Fprintf(os.Stderr, "\nCluster running. RPC address: %s\nPress Ctrl+C to stop.\n\n", cl.RPCAddr)

	// Block until interrupted
	<-ctx.Done()

	logger.Info("shutting down...")
	return nil
}
