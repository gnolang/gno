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

// stagingNoWorkspaceHint is staging mode's line in the no-workspace banner.
const stagingNoWorkspaceHint = "staging eager-loads $GNOROOT/examples (unless -no-examples) and every -extra-root up front."

var defaultStagingOptions = AppConfig{
	chainId:     "dev",
	chainDomain: DefaultDomain,
	logFormat:   "json",
	maxGas:      10_000_000_000,
	webHome:     ":none:",
	// Staging serves the network on one port, unlike local mode. The web
	// listener binds every interface: bound to loopback inside a container,
	// `docker run -p` maps a port only the container itself can reach, so
	// the node reports ready and nothing outside can connect.
	//
	// The node's RPC stays on the loopback. The web listener serves it at
	// /rpc, and to JSON-RPC POSTs on /, through an allowlist of public
	// routes, so a deployment is one hostname and one certificate, and
	// gnokey -remote https://<host> works as is.
	webListenerAddr:            "0.0.0.0:8888",
	nodeRPCListenerAddr:        "127.0.0.1:26657",
	faucetAmount:               "10000000ugnot",
	deployKey:                  defaultDeployerAddress.String(),
	home:                       gnoenv.HomeDir(),
	root:                       gnoenv.RootDir(),
	interactive:                false,
	unsafeAPI:                  false,
	staging:                    true,
	noWorkspaceHint:            stagingNoWorkspaceHint,
	withoutQuarantinedExamples: true,
	emptyBlocks:                false,
	emptyBlocksInterval:        1,

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
Staging eager-loads the workspace, every -extra-root, and $GNOROOT/examples by default (use -no-examples to skip).

The web listener is the one port meant for the network. Next to gnoweb it serves
the node's RPC under /rpc, and JSON-RPC posted to /, limited to public routes, so
'gnokey -remote https://<host>' needs nothing else; and a faucet at /faucet. The
node's own RPC listener stays on the loopback.

Additionally, you can specify an additional package directory to load.
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
	// Staging eager-loads the workspace, every -extra-root, and examples
	// (unless -no-examples is set) via Loader.LoadAll. staging=true
	// in defaultStagingOptions triggers the eager path in app.Setup.
	return runApp(&cfg.AppConfig, io, args...)
}
