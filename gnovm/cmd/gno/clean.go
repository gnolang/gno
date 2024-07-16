package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type cleanCfg struct {
	dryRun   bool // clean -n flag
	verbose  bool // clean -x flag
	modCache bool // clean -modcache flag
}

func newCleanCmd(io commands.IO) *commands.Command {
	cfg := &cleanCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "clean",
			ShortUsage: "clean [flags]",
			ShortHelp:  "removes generated files and cached data",
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execClean(cfg, args, io)
		},
	)
}

func (c *cleanCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.dryRun,
		"n",
		false,
		"print remove commands it would execute, but not run them",
	)

	fs.BoolVar(
		&c.verbose,
		"x",
		false,
		"print remove commands as it executes them",
	)

	fs.BoolVar(
		&c.modCache,
		"modcache",
		false,
		"remove the entire module download cache",
	)
}

func execClean(cfg *cleanCfg, args []string, io commands.IO) error {
	if len(args) > 0 {
		return flag.ErrHelp
	}

	path, err := os.Getwd()
	if err != nil {
		return err
	}
	modDir, err := gnomod.FindRootDir(path)
	if err != nil {
		return fmt.Errorf("not a gno module: %w", err)
	}

	if path != modDir && (cfg.dryRun || cfg.verbose) {
		io.Println("cd", modDir)
	}
	err = clean(modDir, cfg, io)
	if err != nil {
		return err
	}

	if cfg.modCache {
		modCacheDir := filepath.Join(gnoenv.HomeDir(), "pkg", "mod")
		if !cfg.dryRun {
			if err := os.RemoveAll(modCacheDir); err != nil {
				return err
			}
		}
		if cfg.dryRun || cfg.verbose {
			io.Println("rm -rf", modCacheDir)
		}
	}
	return nil
}

// clean removes generated files from a directory.
func clean(dir string, cfg *cleanCfg, io commands.IO) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Ignore if not a generated file
		if !strings.HasSuffix(path, ".gno.gen.go") && !strings.HasSuffix(path, ".gno.gen_test.go") {
			return nil
		}
		if !cfg.dryRun {
			if err := os.Remove(path); err != nil {
				return err
			}
		}
		if cfg.dryRun || cfg.verbose {
			io.Println("rm", strings.TrimPrefix(path, dir+"/"))
		}

		return nil
	})
}
