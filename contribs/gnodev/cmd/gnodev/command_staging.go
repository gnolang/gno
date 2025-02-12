package main

import (
	"context"
	"flag"
	"path/filepath"

	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type StagingAppConfig struct {
	AppConfig
}

var defaultStagingOptions = AppConfig{
	chainId:             "dev",
	chainDomain:         DefaultDomain,
	logFormat:           "json",
	maxGas:              10_000_000_000,
	webHome:             ":none:",
	webListenerAddr:     "127.0.0.1:8888",
	nodeRPCListenerAddr: "127.0.0.1:26657",
	deployKey:           defaultDeployerAddress.String(),
	home:                gnoenv.HomeDir(),
	root:                gnoenv.RootDir(),
	interactive:         false,
	unsafeAPI:           false,
	lazyLoader:          false,
	paths:               filepath.Join(DefaultDomain, "/**"), // Load every package under the main domain},

	// As we have no reason to configure this yet, set this to random port
	// to avoid potential conflict with other app
	nodeP2PListenerAddr:      "tcp://127.0.0.1:0",
	nodeProxyAppListenerAddr: "tcp://127.0.0.1:0",
}

func NewStagingCmd(io commands.IO) *commands.Command {
	var cfg StagingAppConfig

	return commands.NewCommand(
		commands.Metadata{
			Name:          "staging",
			ShortUsage:    "gnodev staging [flags] [package_dir...]",
			ShortHelp:     "Start gnodev in staging mode",
			LongHelp:      "STAGING: Staging mode configure the node for server usage",
			NoParentFlags: true,
		},
		&cfg,
		func(_ context.Context, args []string) error {
			return execStagingCmd(&cfg, args, io)
		},
	)
}

func (c *StagingAppConfig) RegisterFlags(fs *flag.FlagSet) {
	c.AppConfig.RegisterFlagsWith(fs, defaultStagingOptions)
}

func execStagingCmd(cfg *StagingAppConfig, args []string, io commands.IO) error {
	// If no resolvers is defined, use gno example as root resolver
	if len(cfg.AppConfig.resolvers) == 0 {
		gnoroot, err := gnoenv.GuessRootDir()
		if err != nil {
			return err
		}

		exampleRoot := filepath.Join(gnoroot, "examples")
		cfg.AppConfig.resolvers = append(cfg.AppConfig.resolvers, packages.NewRootResolver(exampleRoot))
	}

	return runApp(&cfg.AppConfig, io, args...)
}
