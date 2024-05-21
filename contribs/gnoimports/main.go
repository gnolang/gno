package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"go/scanner"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

const gnoExt = ".gno"

type importsCfg struct {
	write   bool
	verbose bool
}

var defaultDevOptions = &importsCfg{}

func main() {
	cfg := &importsCfg{}

	stdio := commands.NewDefaultIO()
	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "gnoimports",
			ShortUsage: "gnoimports [flags] [path ...]",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execGnoImports(cfg, args, stdio)
		})

	cmd.Execute(context.Background(), os.Args[1:])
}

func (c *importsCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.write,
		"w",
		defaultDevOptions.write,
		"write result to (source) file instead of stdout",
	)

	fs.BoolVar(
		&c.verbose,
		"v",
		defaultDevOptions.verbose,
		"verbose mode",
	)
}

func execGnoImports(cfg *importsCfg, args []string, io commands.IO) (err error) {
	for _, arg := range args {
		files, err := expandsGnoFiles(arg)
		if err != nil {
			return fmt.Errorf("unable to expands gno files: %w", err)
		}

		for _, file := range files {
			if cfg.verbose {
				io.Printfln("processing %q", file)
			}

			var perms os.FileMode
			fi, err := os.Stat(file)
			if err != nil {
				io.ErrPrintfln("unable to stats %q: %s", file, err.Error())
				os.Exit(1)
			}

			data, err := processGnoFile(file)
			if err != nil {
				printScannerError(err, io)
				// io.ErrPrintfln("unable to process %q: %s", file, err.Error())
				os.Exit(1)
			}

			if !cfg.write {
				io.Println(string(data))
				continue
			}

			perms = fi.Mode() & os.ModePerm
			if err = os.WriteFile(file, data, perms); err != nil {
				return err
			}
		}
	}

	return err
}

func expandsGnoFiles(path string) ([]string, error) {
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("unable to get abs path of %q: %w", path, err)
	}

	cleanp, _ := strings.CutSuffix(abs, "/...")
	file, err := os.Stat(cleanp)
	if err != nil {
		return nil, fmt.Errorf("unable to stat %q: %w", path, err)
	}

	if !file.IsDir() {
		return []string{abs}, nil
	}

	directories, err := expandWildcard(path)
	if err != nil {
		return nil, fmt.Errorf("error expanding pattern %q: %w", abs, err)
	}

	// Collect .xft files
	var files []string
	for _, dir := range directories {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && filepath.Ext(path) == gnoExt {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			log.Fatalf("error walking directory %q: %v\n", dir, err)
		}
	}

	return files, nil
}

// Helper function to handle '...' wildcard
func expandWildcard(path string) ([]string, error) {
	var directories []string

	if strings.HasSuffix(path, "/...") {
		basePath := strings.TrimSuffix(path, "/...")
		basePath = filepath.Clean(basePath)

		err := filepath.Walk(basePath, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				directories = append(directories, p)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		directories = []string{path}
	}

	return directories, nil
}

func printScannerError(err error, io commands.IO) {
	for ; err != nil; err = errors.Unwrap(err) {
		perr, ok := err.(ParseError)
		if !ok {
			continue
		}

		// get underlayin parse error
		err = errors.Unwrap(errors.Unwrap(perr))
		if scanErrors, ok := err.(scanner.ErrorList); ok {
			for _, e := range scanErrors {
				io.ErrPrintln(e)
			}

			return
		}

		io.ErrPrintln(err)
		return
	}

	io.ErrPrintln(err)
}
