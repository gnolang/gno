package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
	gno "github.com/gnolang/gno/pkgs/gnolang"
)

type importPath string

type precompileFlags struct {
	Verbose     bool   `flag:"verbose" help:"verbose"`
	SkipFmt     bool   `flag:"skip-fmt" help:"do not check syntax of generated .go files"`
	SkipImports bool   `flag:"skip-imports" help:"do not precompile imports recursively"`
	GoBinary    string `flag:"go-binary" help:"go binary to use for building"`
	GofmtBinary string `flag:"go-binary" help:"gofmt binary to use for syntax checking"`
	Output      string `flag:"output" help:"output directory"`
}

var defaultPrecompileFlags = precompileFlags{
	Verbose:     false,
	SkipFmt:     false,
	SkipImports: false,
	GoBinary:    "go",
	GofmtBinary: "gofmt",
	Output:      ".",
}

type precompileOptions struct {
	flags precompileFlags
	// precompiled is the set of packages already
	// precompiled from .gno to .go.
	precompiled map[importPath]struct{}
}

func newPrecompileOptions(flags precompileFlags) *precompileOptions {
	return &precompileOptions{flags, map[importPath]struct{}{}}
}

func (p *precompileOptions) getFlags() precompileFlags {
	return p.flags
}

func (p *precompileOptions) isPrecompiled(pkg importPath) bool {
	_, precompiled := p.precompiled[pkg]

	return precompiled
}

func (p *precompileOptions) markAsPrecompiled(pkg importPath) {
	p.precompiled[pkg] = struct{}{}
}

func precompileApp(cmd *command.Command, args []string, f interface{}) error {
	flags := f.(precompileFlags)

	if len(args) < 1 {
		cmd.ErrPrintfln("Usage: precompile [precompile flags] [packages]")

		return errors.New("invalid args")
	}

	// precompile .gno files.
	paths, err := gnoFilesFromArgs(args)
	if err != nil {
		return fmt.Errorf("list paths: %w", err)
	}

	opts := newPrecompileOptions(flags)
	errCount := 0

	for _, filepath := range paths {
		err = precompileFile(filepath, opts)
		if err != nil {
			err = fmt.Errorf("%s: precompile: %w", filepath, err)
			cmd.ErrPrintfln("%s", err.Error())
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
	gofmt := flags.GofmtBinary

	if gofmt == "" {
		gofmt = defaultPrecompileFlags.GofmtBinary
	}

	if flags.Verbose {
		fmt.Fprintf(os.Stderr, "%s\n", srcPath)
	}

	// parse .gno.
	source, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	// compute attributes based on filename.
	var (
		targetFilename string
		tags           string
	)

	nameNoExtension := strings.TrimSuffix(filepath.Base(srcPath), ".gno")

	switch {
	case strings.HasSuffix(srcPath, "_filetest.gno"):
		tags = "gno,filetest"
		targetFilename = "." + nameNoExtension + ".gno.gen.go"
	case strings.HasSuffix(srcPath, "_test.gno"):
		tags = "gno,test"
		targetFilename = "." + nameNoExtension + ".gno.gen_test.go"
	default:
		tags = "gno"
		targetFilename = nameNoExtension + ".gno.gen.go"
	}

	// preprocess.
	precompileRes, err := gno.Precompile(string(source), tags, srcPath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// resolve target path
	var targetPath string

	if flags.Output != defaultPrecompileFlags.Output {
		path, err := ResolvePath(flags.Output, importPath(filepath.Dir(srcPath)))
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
	if !flags.SkipFmt {
		err = gno.PrecompileVerifyFile(targetPath, gofmt)
		if err != nil {
			return fmt.Errorf("check .go file: %w", err)
		}
	}

	// precompile imported packages, if `SkipImports` sets to false
	if !flags.SkipImports {
		importPaths := getPathsFromImportSpec(precompileRes.Imports)
		for _, path := range importPaths {
			precompilePkg(path, opts)
		}
	}

	return nil
}
