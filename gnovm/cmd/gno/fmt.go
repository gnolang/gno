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
	"github.com/gnolang/gno/gnovm/pkg/gnoimports"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type fmtCfg struct {
	write   bool
	verbose bool
	imports bool
	include fmtIncludes
}

var defaultFmtOptions = &fmtCfg{
	imports: true,
}

func newFmtCmd(io commands.IO) *commands.Command {
	cfg := &fmtCfg{}
	return commands.NewCommand(
		commands.Metadata{
			Name:       "fmt",
			ShortUsage: "gno fmt [flags] [path ...]",
			ShortHelp:  "Run gno file formater",
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
		defaultFmtOptions.write,
		"write result to (source) file instead of stdout",
	)

	fs.BoolVar(
		&c.verbose,
		"v",
		defaultFmtOptions.verbose,
		"verbose mode",
	)

	fs.BoolVar(
		&c.imports,
		"imports",
		defaultFmtOptions.imports,
		"attempt to format, resolve and sort file imports",
	)

	fs.Var(
		&c.include,
		"i",
		"specify additional directories containing packages to resolve",
	)
}

type fmtProcessFile func(file string, io commands.IO) []byte

func execFmt(cfg *fmtCfg, args []string, io commands.IO) (err error) {
	paths, err := targetsFromPatterns(args)
	if err != nil {
		return fmt.Errorf("unable to get targets paths from paterns: %w", err)
	}

	files, err := gnoFilesFromArgs(paths)
	if err != nil {
		return fmt.Errorf("unable to gather gno files: %w ", err)
	}

	var processFile fmtProcessFile = fmtFormatFile
	if cfg.imports {
		if processFile, err = fmtFormatFileImports(cfg); err != nil {
			return err
		}
	}

	// Process files sequentially
	errCount := 0
	for _, file := range files {
		if cfg.verbose {
			io.Printfln("processing %q", file)
		}

		var perms os.FileMode
		fi, err := os.Stat(file)
		if err != nil {
			errCount++
			io.ErrPrintfln("unable to stats %q: %w", file, err.Error())
			continue
		}

		data := processFile(file, io)
		if data == nil {
			errCount++
			continue
		}

		if !cfg.write {
			io.Println(string(data))
			continue
		}

		perms = fi.Mode() & os.ModePerm // copy permission
		if err = os.WriteFile(file, data, perms); err != nil {
			errCount++
			io.ErrPrintfln("unable to write %q: %w", file, err.Error())
		}
	}

	if errCount > 0 {
		if !cfg.verbose {
			os.Exit(1)
		}

		return fmt.Errorf("failed to format %d files", errCount)
	}

	return nil
}

func fmtFormatFileImports(cfg *fmtCfg) (fmtProcessFile, error) {
	gnoroot := gnoenv.RootDir()

	p := gnoimports.NewProcessor()

	// Load stdlibs
	stdlibs := filepath.Join(gnoroot, "gnovm", "stdlibs")
	if err := p.LoadStdPackages(stdlibs); err != nil {
		return nil, fmt.Errorf("unable to load %q: %w", stdlibs, err)
	}

	// Load examples directory
	examples := filepath.Join(gnoroot, "examples")
	if err := p.LoadPackages(examples); err != nil {
		return nil, fmt.Errorf("unable to load %q: %w", examples, err)
	}

	// Ultimatly load any additional packages supply by the user
	for _, include := range cfg.include {
		absp, err := filepath.Abs(include)
		if err != nil {
			return nil, fmt.Errorf("unable to determine absolute path of %q: %w", include, err)
		}

		if err := p.LoadPackages(absp); err != nil {
			return nil, fmt.Errorf("unable to load %q: %w", absp, err)
		}
	}

	return func(file string, io commands.IO) []byte {
		data, err := p.FormatImports(file)
		if err == nil {
			return data
		}

		// Print parsing errors
		for uerr := err; uerr != nil; uerr = errors.Unwrap(err) {
			perr, ok := uerr.(gnoimports.ParseError)
			if !ok {
				continue
			}

			err = errors.Unwrap(errors.Unwrap(perr)) // get parser error
			fmtPrintScannerError(err, io)
			return nil
		}

		io.ErrPrintfln("format error: %s", err.Error())
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

func fmtPrintScannerError(err error, io commands.IO) {
	// get underlayin parse error
	if scanErrors, ok := err.(scanner.ErrorList); ok {
		for _, e := range scanErrors {
			io.ErrPrintln(e)
		}

		return
	}

	io.ErrPrintln(err)
}

type fmtIncludes []string

func (i fmtIncludes) String() string {
	return strings.Join(i, ",")
}

func (i *fmtIncludes) Set(path string) error {
	*i = append(*i, path)
	return nil
}
