package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gnolang/gno"
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
)

type buildOptions struct {
	Verbose  bool   `flag:"verbose" help:"verbose"`
	GoBinary string `flag:"go-binary" help:"go binary to use for building"`
}

var DefaultBuildOptions = buildOptions{
	Verbose:  false,
	GoBinary: "go",
}

func buildApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(buildOptions)
	if len(args) < 1 {
		cmd.ErrPrintfln("Usage: build [build flags] [packages]")
		return errors.New("invalid args")
	}

	errCount := 0

	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return fmt.Errorf("invalid file or package path: %w", err)
		}
		if !info.IsDir() {
			file := arg
			err = goBuildFileOrPkg(file, opts)
			if err != nil {
				return fmt.Errorf("%s: build file: %w", file, err)
			}
		} else {
			// if the passed arg is a dir, then we'll recursively walk the dir
			// and look for directories containing at least one .gno file.

			visited := map[string]bool{} // used to run the builder only once per folder.
			err = filepath.WalkDir(arg, func(curpath string, f fs.DirEntry, err error) error {
				if err != nil {
					return fmt.Errorf("%s: walk dir: %w", arg, err)
				}
				if f.IsDir() {
					return nil // skip
				}
				if !isGnoFile(f) {
					return nil // skip
				}

				parentDir := filepath.Dir(curpath)
				if _, found := visited[parentDir]; found {
					return nil
				}
				visited[parentDir] = true

				// cannot use path.Join or filepath.Join, because we need
				// to ensure that ./ is the prefix to pass to go build.
				pkg := "./" + parentDir
				err = goBuildFileOrPkg(pkg, opts)
				if err != nil {
					err = fmt.Errorf("%s: build pkg: %w", pkg, err)
					cmd.ErrPrintfln("%s", err.Error())
					errCount++
					return nil
				}
				return nil
			})
			if err != nil {
				return err
			}
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

	return gno.PrecompileCheckPackage(fileOrPkg, goBinary)
}
