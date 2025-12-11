package main

import (
	"context"
	"flag"

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
	loadMode:            LoadModeFull, // Pre-load all packages
	emptyBlocks:         false,
	emptyBlocksInterval: 1,
	// As we have no reason to configure this yet, set this to random port
	// to avoid potential conflict with other app
	nodeP2PListenerAddr:      "tcp://127.0.0.1:0",
	nodeProxyAppListenerAddr: "tcp://127.0.0.1:0",
}

func NewStagingCmd(io commands.IO) *commands.Command {
	var cfg StagingAppConfig

	return commands.NewCommand(
		commands.Metadata{
			Name:       "staging",
			ShortUsage: "gnodev staging [flags] [package_dir...]",
			ShortHelp:  "Start gnodev in staging mode",
			LongHelp: `STAGING: Staging mode configures the node for server usage.
This mode is designed for stability and security, suitable for pre-deployment testing.
Interactive mode and unsafe API access are disabled to ensure a secure environment.
The log format is set to JSON, facilitating integration with logging systems.

PACKAGE LOADING:
This mode uses -load=full by default, which pre-loads all discovered packages.
The lazy loading proxy is disabled in this mode.

Additional package directories can be passed as arguments.
Use -load=auto or -load=lazy to change the loading behavior.
`,
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
	return runApp(&cfg.AppConfig, io, args...)
}
