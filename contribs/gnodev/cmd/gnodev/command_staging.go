package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type stagingCfg struct {
	dev devCfg
}

var defaultStagingOptions = devCfg{
	chainId:             "staging",
	maxGas:              10_000_000_000,
	webListenerAddr:     "127.0.0.1:8888",
	nodeRPCListenerAddr: "127.0.0.1:26657",
	deployKey:           DefaultDeployerAddress.String(),
	home:                gnoenv.HomeDir(),
	root:                gnoenv.RootDir(),
	interactive:         false,
	unsafeAPI:           false,

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
			ShortUsage:    "gnodev staging [flags] <key-name>",
			ShortHelp:     "start gnodev in staging mode",
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
	if len(args) == 0 {
		return fmt.Errorf("no argument given")
	}

	mathches, err := filepath.Glob(args[0])
	if err != nil {
		return fmt.Errorf("invalid glob: %w", err)
	}

	io.Println(strings.Join(mathches, "\n"))
	return nil
}
