package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type buildCfg struct {
	verbose  bool
	goBinary string
}

var defaultBuildOptions = &buildCfg{
	verbose:  false,
	goBinary: "go",
}

func newBuildCmd(io *commands.IO) *commands.Command {
	cfg := &buildCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "build",
			ShortUsage: "build [flags] <package>",
			ShortHelp:  "Builds the specified gno package",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execBuild(cfg, args, io)
		},
	)
}

func (c *buildCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.verbose,
		"verbose",
		defaultBuildOptions.verbose,
		"verbose output when building",
	)

	fs.StringVar(
		&c.goBinary,
		"go-binary",
		defaultBuildOptions.goBinary,
		"go binary to use for building",
	)
}

func execBuild(cfg *buildCfg, args []string, io *commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	paths, err := gnoPackagePathsFromArgs(args)
	if err != nil {
		return fmt.Errorf("list packages: %w", err)
	}

	errCount := 0
	for _, pkgPath := range paths {
		err = goBuildFileOrPkg(pkgPath, cfg)
		if err != nil {
			err = fmt.Errorf("%s: build pkg: %w", pkgPath, err)
			io.ErrPrintfln("%s\n", err.Error())

			errCount++
		}
	}

	if errCount > 0 {
		return fmt.Errorf("%d go build errors", errCount)
	}

	return nil
}

func goBuildFileOrPkg(fileOrPkg string, cfg *buildCfg) error {
	verbose := cfg.verbose
	goBinary := cfg.goBinary

	if verbose {
		fmt.Fprintf(os.Stderr, "%s\n", fileOrPkg)
	}

	return gno.PrecompileBuildPackage(fileOrPkg, goBinary)
}
