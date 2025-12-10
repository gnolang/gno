package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/contribs/gnodev/pkg/packages"
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
	lazyLoader:          true,
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

	// If no resolvers is defined, use gno example as root resolver
	var baseResolvers []packages.Resolver

	if len(cfg.resolvers) == 0 {
		// Check if we are not in gnoroot
		if !strings.HasPrefix(dir, filepath.Clean(cfg.root)+"/") {
			// Add current dir as root resolvers
			baseResolvers = append(baseResolvers, packages.NewRootResolver(dir))
		}

		// Add examples as root resolver
		gnoroot, err := gnoenv.GuessRootDir()
		if err != nil {
			return err
		}
		exampleRoot := filepath.Join(gnoroot, "examples")
		baseResolvers = append(baseResolvers, packages.NewRootResolver(exampleRoot))
	}

	// Check if current directory is a valid gno package
	path := guessPath(&cfg.AppConfig, dir)
	resolver := packages.NewLocalResolver(path, dir)
	if resolver.IsValid() {
		// Add current directory as local resolver
		baseResolvers = append([]packages.Resolver{resolver}, baseResolvers...)
		if len(cfg.paths) > 0 {
			cfg.paths += ","
		}
		cfg.paths += resolver.Path
	}
	cfg.resolvers = append(baseResolvers, cfg.resolvers...)

	return runApp(&cfg.AppConfig, cio) // else run app without any dir
}
