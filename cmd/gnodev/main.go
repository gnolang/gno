package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"io/ioutil"
	"os"
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
	// build-dry-run
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
	Verbose bool `flag:"verbose" help:"verbose"`
}

var DefaultPrecompileOptions = precompileOptions{
	Verbose: false,
}

func precompileApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(precompileOptions)
	if len(args) < 1 {
		cmd.ErrPrintfln("Usage: precompile [precompile flags] [packages]")
		return errors.New("invalid args")
	}

	verbose := opts.Verbose
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return fmt.Errorf("invalid package path: %w", err)
		}
		if !info.IsDir() {
			filepath := arg
			err = precompileFile(filepath, verbose)
			if err != nil {
				return fmt.Errorf("%s: failed to precompile: %w", filepath, err)
			}
		} else {
			err = filepath.WalkDir(arg, func(filepath string, f fs.DirEntry, err error) error {
				if err != nil {
					return fmt.Errorf("%s: failed to walk dir: %w", arg, err)
				}

				if !isGnoFile(f) {
					return nil // skip
				}
				err = precompileFile(filepath, verbose)
				if err != nil {
					return fmt.Errorf("%s: failed to precompile: %w", filepath, err)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func precompileFile(srcPath string, verbose bool) error {
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
