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
	loadMode:            LoadModeAuto,
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
It enables features such as interactive mode, unsafe API access for testing, and on-demand
package loading. The log format is set to console for easier readability, and the web
interface is accessible locally, making it ideal for iterative development and testing.

LOAD MODES (-load flag):
  auto   Pre-load current workspace/package only. If running from the examples folder,
         uses lazy loading instead. (default)
  lazy   Load packages on-demand as they are accessed via queries or transactions.
  full   Pre-load all discovered packages under the chain domain.

PACKAGE DISCOVERY:
  - If the current directory contains a gnomod.toml file, the package is automatically
    detected and loaded using the module path defined in the file.
  - If the current directory contains a gnowork.toml file, it is treated as a workspace
    and all packages within are discovered.
  - Additional package directories can be passed as arguments.
  - The -paths flag can be used to pre-load additional packages on top of the load mode.

The examples folder from GNOROOT is always included as a workspace for dependency resolution.
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

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to guess current dir: %w", err)
	}

	// Always add current directory as workspace root for discovery
	// (even if it's not itself a gno package, it may contain packages in subdirs)
	// The load mode will determine what gets pre-loaded from these directories
	args = append([]string{dir}, args...)

	// If args are provided, they are directories to add
	return runApp(&cfg.AppConfig, cio, args...)
}
