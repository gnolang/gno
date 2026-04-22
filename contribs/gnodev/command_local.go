package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
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

By default, the current directory and the "example" folder from "gnoroot" will be used as the root resolver.
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

	// Append the current directory's package path to cfg.paths so the node
	// preloads/watches it. Workspace discovery + -extra-root now drive the
	// loader; no resolver wiring is needed here.
	path := guessPath(&cfg.AppConfig, dir)
	if _, err := gnolang.ReadMemPackage(dir, path, gnolang.MPAnyAll); err == nil {
		if len(cfg.paths) > 0 {
			cfg.paths += ","
		}
		cfg.paths += path

		// Register CWD as an extra root when no workspace was detected.
		// Without this, a loose realm dir (no gnomod.toml/gnowork.toml ancestor)
		// would be referenced in cfg.paths but unresolvable by the loader.
		if packages.FindWorkspace(dir) == "" {
			cfg.extraRoots = append(cfg.extraRoots, dir)
		}
	}

	return runApp(&cfg.AppConfig, cio) // else run app without any dir
}
