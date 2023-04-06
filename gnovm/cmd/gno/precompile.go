package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type importPath string

type precompileCfg struct {
	verbose     bool
	skipFmt     bool
	skipImports bool
	goBinary    string
	gofmtBinary string
	output      string
}

type precompileOptions struct {
	cfg *precompileCfg
	// precompiled is the set of packages already
	// precompiled from .gno to .go.
	precompiled map[importPath]struct{}
}

func newPrecompileOptions(cfg *precompileCfg) *precompileOptions {
	return &precompileOptions{cfg, map[importPath]struct{}{}}
}

func (p *precompileOptions) getFlags() *precompileCfg {
	return p.cfg
}

func (p *precompileOptions) isPrecompiled(pkg importPath) bool {
	_, precompiled := p.precompiled[pkg]
	return precompiled
}

func (p *precompileOptions) markAsPrecompiled(pkg importPath) {
	p.precompiled[pkg] = struct{}{}
}

func newPrecompileCmd(io *commands.IO) *commands.Command {
	cfg := &precompileCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "precompile",
			ShortUsage: "precompile [flags] <package> [<package>...]",
			ShortHelp:  "Precompiles .gno files to .go",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execPrecompile(cfg, args, io)
		},
	)
}

func (c *precompileCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.verbose,
		"verbose",
		false,
		"verbose output when running",
	)

	fs.BoolVar(
		&c.skipFmt,
		"skip-fmt",
		false,
		"do not check syntax of generated .go files",
	)

	fs.BoolVar(
		&c.skipImports,
		"skip-imports",
		false,
		"do not precompile imports recursively",
	)

	fs.StringVar(
		&c.goBinary,
		"go-binary",
		"go",
		"go binary to use for building",
	)

	fs.StringVar(
		&c.gofmtBinary,
		"go-fmt-binary",
		"gofmt",
		"gofmt binary to use for syntax checking",
	)

	fs.StringVar(
		&c.output,
		"output",
		".",
		"output directory",
	)
}

func execPrecompile(cfg *precompileCfg, args []string, io *commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	// precompile .gno files.
	paths, err := gnoFilesFromArgs(args)
	if err != nil {
		return fmt.Errorf("list paths: %w", err)
	}

	opts := newPrecompileOptions(cfg)
	errCount := 0
	for _, filepath := range paths {
		err = precompileFile(filepath, opts)
		if err != nil {
			err = fmt.Errorf("%s: precompile: %w", filepath, err)
			io.ErrPrintfln("%s", err.Error())

			errCount++
		}
	}

	if errCount > 0 {
		return fmt.Errorf("%d precompile errors", errCount)
	}

	return nil
}

func precompilePkg(pkgPath importPath, opts *precompileOptions) error {
	if opts.isPrecompiled(pkgPath) {
		return nil
	}
	opts.markAsPrecompiled(pkgPath)

	files, err := filepath.Glob(filepath.Join(string(pkgPath), "*.gno"))
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if err = precompileFile(file, opts); err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
	}

	return nil
}

func precompileFile(srcPath string, opts *precompileOptions) error {
	flags := opts.getFlags()
	gofmt := flags.gofmtBinary
	if gofmt == "" {
		gofmt = "gofmt"
	}

	if flags.verbose {
		fmt.Fprintf(os.Stderr, "%s\n", srcPath)
	}

	// parse .gno.
	source, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	// compute attributes based on filename.
	targetFilename, tags := gno.GetPrecompileFilenameAndTags(srcPath)

	// preprocess.
	precompileRes, err := gno.Precompile(string(source), tags, srcPath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// resolve target path
	var targetPath string
	if flags.output != "." {
		path, err := ResolvePath(flags.output, importPath(filepath.Dir(srcPath)))
		if err != nil {
			return fmt.Errorf("resolve output path: %w", err)
		}
		targetPath = filepath.Join(path, targetFilename)
	} else {
		targetPath = filepath.Join(filepath.Dir(srcPath), targetFilename)
	}

	// write .go file.
	err = WriteDirFile(targetPath, []byte(precompileRes.Translated))
	if err != nil {
		return fmt.Errorf("write .go file: %w", err)
	}

	// check .go fmt, if `SkipFmt` sets to false.
	if !flags.skipFmt {
		err = gno.PrecompileVerifyFile(targetPath, gofmt)
		if err != nil {
			return fmt.Errorf("check .go file: %w", err)
		}
	}

	// precompile imported packages, if `SkipImports` sets to false
	if !flags.skipImports {
		importPaths := getPathsFromImportSpec(precompileRes.Imports)
		for _, path := range importPaths {
			precompilePkg(path, opts)
		}
	}

	return nil
}
