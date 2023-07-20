package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

type lintCfg struct {
	verbose bool
	rootDir string
	// dry-run
	// auto-fix
}

// – setup in ci
// – add comment about what to do with syntax, also
// – gno.mod

func newLintCmd(io *commands.IO) *commands.Command {
	cfg := &lintCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "lint",
			ShortUsage: "lint [flags] <package> [<package>...]",
			ShortHelp:  "Runs the linter for the specified packages",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execLint(cfg, args, io)
		},
	)
}

func (c *lintCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.verbose,
		"verbose",
		false,
		"verbose output when lintning",
	)

	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"",
		"clone location of github.com/gnolang/gno (gnodev tries to guess it)",
	)
}

func execLint(cfg *lintCfg, args []string, io *commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	var (
		// verbose = cfg.verbose
		rootDir = cfg.rootDir
		// stdin   = io.In
		// stdout  = io.Out
		// stderr  = io.Err
	)
	if rootDir == "" {
		rootDir = guessRootDir()
	}

	pkgPaths, err := gnoPackagesFromArgs(args)
	if err != nil {
		return fmt.Errorf("list packages from args: %w", err)
	}

	for _, pkgPath := range pkgPaths {
		println(pkgPath)
		// 1. lint the package (gno.mod, etc)
		// 2. lint the files individually
		// 3. TODO: consider making `gno precompile; go lint *gen.go`
	}
	return nil
}
