package main

import (
	"context"
	"flag"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type stagingCfg struct {
	devCfg
}

var defaultStagingOptions = devCfg{
	chainId:             "staging",
	maxGas:              10_000_000_000,
	webListenerAddr:     "127.0.0.1:8888",
	nodeRPCListenerAddr: "127.0.0.1:26657",
	deployKey:           DefaultDeployerAddress.String(),
	home:                gnoenv.HomeDir(),
	root:                gnoenv.RootDir(),
	serverMode:          true,
	unsafeAPI:           false,

	// As we have no reason to configure this yet, set this to random port
	// to avoid potential conflict with other app
	nodeP2PListenerAddr:      "tcp://127.0.0.1:0",
	nodeProxyAppListenerAddr: "tcp://127.0.0.1:0",
}

func NewStagingCmd(io commands.IO) *commands.Command {
	cfg := &stagingCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:          "staging",
			ShortUsage:    "gnodev staging [flags] <key-name>",
			ShortHelp:     "start gnodev in staging mode",
			NoParentFlags: true,
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execStagingCmd(cfg, args, io)
		},
	)
}

func (c *stagingCfg) RegisterFlags(fs *flag.FlagSet) {
	c.devCfg.registerFlagsWithDefault(defaultStagingOptions, fs)
}

func execStagingCmd(cg *stagingCfg, args []string, io commands.IO) error {
	return nil
}
