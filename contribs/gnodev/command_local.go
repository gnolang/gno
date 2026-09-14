package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/mattn/go-isatty"
)

const DefaultDomain = "gno.land"

// localNoWorkspaceHint is local mode's line in the no-workspace banner.
const localNoWorkspaceHint = "running in discovery mode: packages resolve on-demand from examples, and from a chain RPC for domains passed via -remote."

var ErrConflictingFileArgs = errors.New("cannot specify `balances-file` or `txs-file` along with `genesis-file`")

type LocalAppConfig struct {
	AppConfig

	chdir string // directory context
}

var defaultLocalAppConfig = AppConfig{
	chainId:             "dev",
	logFormat:           "console",
	chainDomain:         DefaultDomain,
	maxGas:              10_000_000_000,
	webListenerAddr:     "127.0.0.1:8888",
	nodeRPCListenerAddr: "127.0.0.1:26657",
	deployKey:           defaultDeployerAddress.String(),
	home:                gnoenv.HomeDir(),
	root:                gnoenv.RootDir(),
	interactive:         isatty.IsTerminal(os.Stdout.Fd()),
	unsafeAPI:           true,
	noWorkspaceHint:     localNoWorkspaceHint,
	emptyBlocks:         false,
	emptyBlocksInterval: 1,

	// As we have no reason to configure this yet, set this to random port
	// to avoid potential conflict with other app
	nodeP2PListenerAddr:      "tcp://127.0.0.1:0",
	nodeProxyAppListenerAddr: "tcp://127.0.0.1:0",
}

func NewLocalCmd(io commands.IO) *commands.Command {
	var cfg LocalAppConfig

	return commands.NewCommand(
		commands.Metadata{
			Name:       "local",
			ShortUsage: "gnodev local [flags] [package_dir...]",
			ShortHelp:  "Start gnodev in local development mode (default)",
			LongHelp: `LOCAL: Local mode configures the node for local development usage.
This mode is optimized for realm development, providing an interactive and flexible environment.
It enables features such as interactive mode, unsafe API access for testing, and lazy loading to improve performance.
The log format is set to console for easier readability, and the web interface is accessible locally, making it ideal for iterative development and testing.

By default, the workspace containing the current directory is loaded at startup;
packages from "$GNOROOT/examples" are resolved on demand upon a query or transaction.
`,
			NoParentFlags: true,
		},
		&cfg,
		func(_ context.Context, args []string) error {
			return execLocalApp(&cfg, args, io)
		},
	)
}

func (c *LocalAppConfig) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.chdir,
		"C",
		c.chdir,
		"change directory context before running gnodev",
	)

	c.AppConfig.RegisterFlagsWith(fs, defaultLocalAppConfig)
}

func execLocalApp(cfg *LocalAppConfig, args []string, cio commands.IO) error {
	if cfg.chdir != "" {
		if err := os.Chdir(cfg.chdir); err != nil {
			return fmt.Errorf("unable to change directory: %w", err)
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to guess current dir: %w", err)
	}

	// Explicit [package_dir...] args first, then the CWD; Setup absolutizes
	// and deduplicates them, applies the same handling to each (gnomod.toml
	// module path, or a generated one with a warning), and quietly skips
	// the CWD when it holds no gno package.
	dirs := make([]string, 0, len(args)+1)
	dirs = append(dirs, args...)
	dirs = append(dirs, cwd)
	return runApp(&cfg.AppConfig, cio, dirs...)
}
