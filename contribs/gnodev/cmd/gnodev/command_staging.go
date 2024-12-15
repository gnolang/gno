package main

import (
	"context"
	"flag"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type stagingCfg struct {
	dev devCfg
}

var defaultStagingOptions = devCfg{
	chainId:             "staging",
	chainDomain:         DefaultDomain,
	logFormat:           "json",
	maxGas:              10_000_000_000,
	webHome:             "/",
	webListenerAddr:     "127.0.0.1:8888",
	nodeRPCListenerAddr: "127.0.0.1:26657",
	deployKey:           DefaultDeployerAddress.String(),
	home:                gnoenv.HomeDir(),
	root:                gnoenv.RootDir(),
	interactive:         false,
	unsafeAPI:           false,
	paths:               filepath.Join(DefaultDomain, "/**"), // Load every package under the main domain},

	// As we have no reason to configure this yet, set this to random port
	// to avoid potential conflict with other app
	nodeP2PListenerAddr:      "tcp://127.0.0.1:0",
	nodeProxyAppListenerAddr: "tcp://127.0.0.1:0",
}

func NewStagingCmd(io commands.IO) *commands.Command {
	var cfg stagingCfg

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

func (c *stagingCfg) RegisterFlags(fs *flag.FlagSet) {
	c.dev.registerFlagsWithDefault(defaultStagingOptions, fs)
}

func execStagingCmd(cfg *stagingCfg, args []string, io commands.IO) error {
	return runApp(&cfg.dev, io, args...)
}
