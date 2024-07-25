package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnofmt"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/rogpeppe/go-internal/diff"
)

type fmtCfg struct {
	write   bool
	quiet   bool
	diff    bool
	verbose bool
	imports bool
	include fmtIncludes
}

func newFmtCmd(io commands.IO) *commands.Command {
	cfg := &fmtCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "fmt",
			ShortUsage: "gno fmt [flags] [path ...]",
			ShortHelp:  "Run gno file formatter.",
			LongHelp:   "The `gno fmt` tool processes, formats, and cleans up `gno` source files.",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execFmt(cfg, args, io)
		})
}

func (c *fmtCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.write,
		"w",
		false,
		"write result to (source) file instead of stdout",
	)

	fs.BoolVar(
		&c.verbose,
		"v",
		false,
		"verbose mode",
	)

	fs.BoolVar(
		&c.quiet,
		"q",
		false,
		"quiet mode",
	)

	fs.Var(
		&c.include,
		"include",
		"specify additional directories containing packages to resolve",
	)

	fs.BoolVar(
		&c.imports,
		"imports",
		true,
		"attempt to format, resolve and sort file imports",
	)

	fs.BoolVar(
		&c.diff,
		"diff",
		false,
		"print and make the command fail if any diff is found",
	)
}

type fmtProcessFileFunc func(file string, io commands.IO) []byte

func execFmt(cfg *fmtCfg, args []string, io commands.IO) error {
	if len(args) == 0 {
		return flag.ErrHelp
	}

	paths, err := targetsFromPatterns(args)
	if err != nil {
		return fmt.Errorf("unable to get targets paths from patterns: %w", err)
	}

	files, err := gnoFilesFromArgs(paths)
	if err != nil {
		return fmt.Errorf("unable to gather gno files: %w", err)
	}

	processFileFunc, err := fmtGetProcessFileFunc(cfg, io)
	if err != nil {
		return err
	}

	errCount := fmtProcessFiles(cfg, files, processFileFunc, io)
	if errCount > 0 {
		if !cfg.verbose {
			return commands.ExitCodeError(1)
		}

		return fmt.Errorf("failed to format %d files", errCount)
	}

	return nil
}

func fmtGetProcessFileFunc(cfg *fmtCfg, io commands.IO) (fmtProcessFileFunc, error) {
	if cfg.imports {
		return fmtFormatFileImports(cfg, io)
	}
	return fmtFormatFile, nil
}

func fmtProcessFiles(cfg *fmtCfg, files []string, processFile fmtProcessFileFunc, io commands.IO) int {
	errCount := 0
	for _, file := range files {
		if fmtProcessSingleFile(cfg, file, processFile, io) {
			continue // ok
		}

		errCount++
	}
	return errCount
}

// fmtProcessSingleFile process a single file and return false if any error occurred
func fmtProcessSingleFile(cfg *fmtCfg, file string, processFile fmtProcessFileFunc, io commands.IO) bool {
	if cfg.verbose {
		io.Printfln("processing %q", file)
	}

	fi, err := os.Stat(file)
	if err != nil {
		io.ErrPrintfln("unable to stat %q: %v", file, err)
		return false
	}

	out := processFile(file, io)
	if out == nil {
		return false
	}

	if cfg.diff && fmtProcessDiff(file, out, io) {
		return false
	}
	if !cfg.write {
		if !cfg.diff && !cfg.quiet {
			io.Println(string(out))
		}
		return true
	}

	perms := fi.Mode() & os.ModePerm
	if err = os.WriteFile(file, out, perms); err != nil {
		io.ErrPrintfln("unable to write %q: %v", file, err)
		return false
	}

	return true
}

func fmtProcessDiff(file string, data []byte, io commands.IO) bool {
	oldFile, err := os.ReadFile(file)
	if err != nil {
		io.ErrPrintfln("unable to read %q for diffing: %v", file, err)
		return true
	}

	if d := diff.Diff(file, oldFile, file+".formatted", data); d != nil {
		io.ErrPrintln(string(d))
		return true
	}

	return false
}

func fmtFormatFileImports(cfg *fmtCfg, io commands.IO) (fmtProcessFileFunc, error) {
	r := gnofmt.NewFSResolver()

	gnoroot := gnoenv.RootDir()

	pkgHandler := func(path string, err error) error {
		if err == nil {
			return nil
		}

		if !fmtPrintScannerError(err, io) {
			io.ErrPrintfln("unable to load %q: %w", err.Error())
		}

		return nil
	}

	// Load any additional packages supplied by the user
	for _, include := range cfg.include {
		absp, err := filepath.Abs(include)
		if err != nil {
			return nil, fmt.Errorf("unable to determine absolute path of %q: %w", include, err)
		}

		if err := r.LoadPackages(absp, pkgHandler); err != nil {
			return nil, fmt.Errorf("unable to load %q: %w", absp, err)
		}
	}

	// Load stdlibs
	stdlibs := filepath.Join(gnoroot, "gnovm", "stdlibs")
	if err := r.LoadPackages(stdlibs, pkgHandler); err != nil {
		return nil, fmt.Errorf("unable to load %q: %w", stdlibs, err)
	}

	// Load examples directory
	examples := filepath.Join(gnoroot, "examples")
	if err := r.LoadPackages(examples, pkgHandler); err != nil {
		return nil, fmt.Errorf("unable to load %q: %w", examples, err)
	}

	p := gnofmt.NewProcessor(r)
	return func(file string, io commands.IO) []byte {
		data, err := p.FormatFile(file)
		if err == nil {
			return data
		}

		if !fmtPrintScannerError(err, io) {
			io.ErrPrintfln("format error: %s", err.Error())
		}

		return nil
	}, nil
}

func fmtFormatFile(file string, io commands.IO) []byte {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, file, nil, parser.AllErrors|parser.ParseComments)
	if err != nil {
		fmtPrintScannerError(err, io)
		return nil
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		io.ErrPrintfln("format error: %s", err.Error())
		return nil
	}

	return buf.Bytes()
}

func fmtPrintScannerError(err error, io commands.IO) bool {
	// Get underlying parse error
	for ; err != nil; err = errors.Unwrap(err) {
		if scanErrors, ok := err.(scanner.ErrorList); ok {
			for _, e := range scanErrors {
				io.ErrPrintln(e)
			}

			return true
		}
	}

	return false
}

type fmtIncludes []string

func (i fmtIncludes) String() string {
	return strings.Join(i, ",")
}

func (i *fmtIncludes) Set(path string) error {
	*i = append(*i, path)
	return nil
}
