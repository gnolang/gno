package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno"
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
)

func main() {
	cmd := command.NewStdCommand()
	exec := os.Args[0]
	args := os.Args[1:]
	err := runMain(cmd, exec, args)
	if err != nil {
		cmd.ErrPrintfln("%s", err.Error())
		cmd.ErrPrintfln("%#v", err)
		return // exit
	}
}

type AppItem = command.AppItem
type AppList = command.AppList

var mainApps AppList = []AppItem{
	{precompileApp, "precompile", "precompile .gno to .go", DefaultPrecompileOptions},
}

func runMain(cmd *command.Command, exec string, args []string) error {

	// show help message.
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" {
		cmd.Println("available subcommands:")
		for _, appItem := range mainApps {
			cmd.Printf("  %s - %s\n", appItem.Name, appItem.Desc)
		}
		return nil
	}

	// switch on first argument.
	for _, appItem := range mainApps {
		if appItem.Name == args[0] {
			err := cmd.Run(appItem.App, args[1:], appItem.Defaults)
			return err // done
		}
	}

	// unknown app command!
	return errors.New("unknown command " + args[0])

}

//----------------------------------------
// precompileApp

type precompileOptions struct {
	Verbose   bool   `flag:"verbose" help:"verbose"`
	SkipBuild bool   `flag:"skip-build" help:"convert to .go without building with go"`
	GoBinary  string `flag:"go-binary" help:"go binary to use for building"`
}

var DefaultPrecompileOptions = precompileOptions{
	Verbose:   false,
	SkipBuild: false,
	GoBinary:  "go",
}

func precompileApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(precompileOptions)
	if len(args) < 1 {
		cmd.ErrPrintfln("Usage: precompile [precompile flags] [packages]")
		return errors.New("invalid args")
	}

	// precompile files.
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return fmt.Errorf("invalid file or package path: %w", err)
		}
		if !info.IsDir() {
			curpath := arg
			err = precompileFile(curpath, opts)
			if err != nil {
				return fmt.Errorf("%s: failed to precompile: %w", curpath, err)
			}
		} else {
			err = filepath.WalkDir(arg, func(curpath string, f fs.DirEntry, err error) error {
				if err != nil {
					return fmt.Errorf("%s: failed to walk dir: %w", arg, err)
				}

				if !isGnoFile(f) {
					return nil // skip
				}
				err = precompileFile(curpath, opts)
				if err != nil {
					return fmt.Errorf("%s: failed to precompile: %w", curpath, err)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
	}

	// go build the generated packages.
	shouldBuild := !opts.SkipBuild
	if shouldBuild {
		for _, arg := range args {
			info, err := os.Stat(arg)
			if err != nil {
				return fmt.Errorf("invalid file or package path: %w", err)
			}
			if !info.IsDir() {
				file := arg
				err = goBuildFileOrPkg(file, opts)
				if err != nil {
					return fmt.Errorf("%s: failed to build file: %w", file, err)
				}
			} else {
				// if the passed arg is a dir, then we'll recursively walk the dir
				// and look for directories containing at least one .gno file.

				visited := map[string]bool{} // used to run the builder only once per folder.
				err = filepath.WalkDir(arg, func(curpath string, f fs.DirEntry, err error) error {
					if err != nil {
						return fmt.Errorf("%s: failed to walk dir: %w", arg, err)
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

					pkg := "./" + parentDir
					err = goBuildFileOrPkg(pkg, opts)
					if err != nil {
						return fmt.Errorf("%s: failed to build pkg: %w", pkg, err)
					}
					return nil
				})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func goBuildFileOrPkg(fileOrPkg string, opts precompileOptions) error {
	verbose := opts.Verbose
	goBinary := opts.GoBinary

	// TODO: should we call cmd/compile instead of exec?
	// TODO: guess the nearest go.mod file, chdir in the same folder, adapt trim prefix from fileOrPkg?
	args := []string{"build", "-v", "-tags=gno", fileOrPkg}
	cmd := exec.Command(goBinary, args...)
	if verbose {
		fmt.Fprintln(os.Stderr, cmd.String())
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, string(out))
		return fmt.Errorf("failed to build .go file: %w", err)
	}

	return nil
}

func precompileFile(srcPath string, opts precompileOptions) error {
	verbose := opts.Verbose
	if verbose {
		fmt.Fprintln(os.Stderr, srcPath)
	}

	// parse .gno.
	source, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read: %w", err)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, srcPath, source, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	// preprocess.
	transformed, err := gno.Precompile(fset, f)
	if err != nil {
		return fmt.Errorf("failed to precompile: %w", err)
	}

	// write .go file.
	targetPath := strings.TrimSuffix(srcPath, ".gno") + ".gno.gen.go"
	err = ioutil.WriteFile(targetPath, []byte(transformed), 0644)
	if err != nil {
		return fmt.Errorf("failed to write .go file: %w", err)
	}

	return nil
}

func isGnoFile(f fs.DirEntry) bool {
	name := f.Name()
	return !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".gno") && !f.IsDir()
}
