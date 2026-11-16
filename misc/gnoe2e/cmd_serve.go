package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/gnolang/gno/misc/gnoe2e/internal/builder"
	"github.com/gnolang/gno/misc/gnoe2e/internal/cluster"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type serveCfg struct {
	cluster.ClusterConfig
	verbose bool
}

// Validate refuses a genesis the node would refuse, before the command builds
// the binary that would boot it.
//
// serve has no scenario to fill anything in, so the flags describe the chain
// completely: unlike the run template, whose approver set is filled per
// scenario, what is here is what starts.
func (c *serveCfg) Validate() error {
	if err := c.ClusterConfig.Validate(); err != nil {
		return err
	}
	state, err := cluster.ResolveGenesisState(c.Genesis)
	if err != nil {
		return err
	}
	return cluster.ValidateGenesisState(state)
}

func (c *serveCfg) RegisterFlags(fs *flag.FlagSet) {
	c.ClusterConfig.RegisterFlags(fs)
	fs.BoolVar(&c.verbose, "verbose", false, "verbose output")
}

func newServeCmd(io commands.IO) *commands.Command {
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
			ctx, cancel := signal.NotifyContext(ctx, runSignals...)
			defer cancel()

			return execServe(ctx, io, cfg)
		},
	)
}

func execServe(ctx context.Context, io commands.IO, cfg *serveCfg) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	logger := newRunLogger(os.Stderr, cfg.verbose)
	cfg.ClusterConfig.Logger = logger.With("component", "cluster")

	// The binary's directory is this command's, so the ~100 MB gnoland it
	// builds goes away with the cluster it was built for.
	binDir, err := os.MkdirTemp("", "gnoe2e-bin-*")
	if err != nil {
		return fmt.Errorf("create binary dir: %w", err)
	}
	defer os.RemoveAll(binDir)

	bldr := builder.NewLocalBuilder()
	gnolandBin, err := bldr.Build(ctx, builder.BuildOpts{OutDir: binDir})
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
	io.Printfln("\nCluster running. RPC address: %s\nPress Ctrl+C to stop.\n", cl.RPCAddr)

	<-ctx.Done()

	logger.Info("shutting down...")
	return nil
}
