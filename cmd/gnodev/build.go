package main

import (
	"fmt"
	"os"

	"github.com/gnolang/gno"
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
)

type buildOptions struct {
	Verbose  bool   `flag:"verbose" help:"verbose"`
	GoBinary string `flag:"go-binary" help:"go binary to use for building"`
}

var defaultBuildOptions = buildOptions{
	Verbose:  false,
	GoBinary: "go",
}

func buildApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(buildOptions)
	if len(args) < 1 {
		cmd.ErrPrintfln("Usage: build [build flags] [packages]")
		return errors.New("invalid args")
	}

	paths, err := gnoPackagesFromArgs(args)
	if err != nil {
		return fmt.Errorf("list packages: %w", err)
	}

	errCount := 0
	for _, pkgPath := range paths {
		err = goBuildFileOrPkg(pkgPath, opts)
		if err != nil {
			err = fmt.Errorf("%s: build pkg: %w", pkgPath, err)
			cmd.ErrPrintfln("%s", err.Error())
			errCount++
		}
	}
	if errCount > 0 {
		return fmt.Errorf("%d go build errors", errCount)
	}

	return nil
}

func goBuildFileOrPkg(fileOrPkg string, opts buildOptions) error {
	verbose := opts.Verbose
	goBinary := opts.GoBinary

	if verbose {
		fmt.Fprintf(os.Stderr, "%s\n", fileOrPkg)
	}

	return gno.PrecompileBuildPackage(fileOrPkg, goBinary)
}
