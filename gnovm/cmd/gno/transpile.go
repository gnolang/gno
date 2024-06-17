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

	"github.com/gnolang/gno/gnovm/pkg/transpiler"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type importPath string

type transpileCfg struct {
	verbose     bool
	skipFmt     bool
	skipImports bool
	gobuild     bool
	goBinary    string
	gofmtBinary string
	output      string
}

type transpileOptions struct {
	cfg *transpileCfg
	// transpiled is the set of packages already
	// transpiled from .gno to .go.
	transpiled map[importPath]struct{}
}

func newTranspileOptions(cfg *transpileCfg) *transpileOptions {
	return &transpileOptions{cfg, map[importPath]struct{}{}}
}

func (p *transpileOptions) getFlags() *transpileCfg {
	return p.cfg
}

func (p *transpileOptions) isTranspiled(pkg importPath) bool {
	_, transpiled := p.transpiled[pkg]
	return transpiled
}

func (p *transpileOptions) markAsTranspiled(pkg importPath) {
	p.transpiled[pkg] = struct{}{}
}

func newTranspileCmd(io commands.IO) *commands.Command {
	cfg := &transpileCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "transpile",
			ShortUsage: "transpile [flags] <package> [<package>...]",
			ShortHelp:  "transpiles .gno files to .go",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execTranspile(cfg, args, io)
		},
	)
}

func (c *transpileCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.verbose,
		"v",
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
		"do not transpile imports recursively",
	)

	fs.BoolVar(
		&c.gobuild,
		"gobuild",
		false,
		"run go build on generated go files, ignoring test files",
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

func execTranspile(cfg *transpileCfg, args []string, io commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	// transpile .gno files.
	paths, err := gnoFilesFromArgsRecursively(args)
	if err != nil {
		return fmt.Errorf("list paths: %w", err)
	}

	opts := newTranspileOptions(cfg)
	var errlist scanner.ErrorList
	for _, filepath := range paths {
		if err := transpileFile(filepath, opts); err != nil {
			var fileErrlist scanner.ErrorList
			if !errors.As(err, &fileErrlist) {
				// Not an scanner.ErrorList: return immediately.
				return fmt.Errorf("%s: transpile: %w", filepath, err)
			}
			errlist = append(errlist, fileErrlist...)
		}
	}

	if errlist.Len() == 0 && cfg.gobuild {
		paths, err := gnoPackagesFromArgsRecursively(args)
		if err != nil {
			return fmt.Errorf("list packages: %w", err)
		}

		for _, pkgPath := range paths {
			err := goBuildFileOrPkg(pkgPath, cfg)
			if err != nil {
				var fileErrlist scanner.ErrorList
				if !errors.As(err, &fileErrlist) {
					// Not an scanner.ErrorList: return immediately.
					return fmt.Errorf("%s: build: %w", pkgPath, err)
				}
				errlist = append(errlist, fileErrlist...)
			}
		}
	}

	if errlist.Len() > 0 {
		for _, err := range errlist {
			io.ErrPrintfln(err.Error())
		}
		return fmt.Errorf("%d transpile error(s)", errlist.Len())
	}
	return nil
}

func transpilePkg(pkgPath importPath, opts *transpileOptions) error {
	if opts.isTranspiled(pkgPath) {
		return nil
	}
	opts.markAsTranspiled(pkgPath)

	files, err := filepath.Glob(filepath.Join(string(pkgPath), "*.gno"))
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if err = transpileFile(file, opts); err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
	}

	return nil
}

func transpileFile(srcPath string, opts *transpileOptions) error {
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
	targetFilename, tags := transpiler.GetTranspileFilenameAndTags(srcPath)

	// preprocess.
	transpileRes, err := transpiler.Transpile(string(source), tags, srcPath)
	if err != nil {
		return fmt.Errorf("transpile: %w", err)
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
	err = WriteDirFile(targetPath, []byte(transpileRes.Translated))
	if err != nil {
		return fmt.Errorf("write .go file: %w", err)
	}

	// check .go fmt, if `SkipFmt` sets to false.
	if !flags.skipFmt {
		err = transpiler.TranspileVerifyFile(targetPath, gofmt)
		if err != nil {
			return fmt.Errorf("check .go file: %w", err)
		}
	}

	// transpile imported packages, if `SkipImports` sets to false
	if !flags.skipImports {
		importPaths := getPathsFromImportSpec(transpileRes.Imports)
		for _, path := range importPaths {
			if err := transpilePkg(path, opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func goBuildFileOrPkg(fileOrPkg string, cfg *transpileCfg) error {
	verbose := cfg.verbose
	goBinary := cfg.goBinary

	if verbose {
		fmt.Fprintf(os.Stderr, "%s\n", fileOrPkg)
	}

	return transpiler.TranspileBuildPackage(fileOrPkg, goBinary)
}
