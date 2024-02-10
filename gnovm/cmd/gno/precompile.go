package main

import (
	"context"
	"flag"
	"fmt"
	precompile "github.com/gnolang/gno/gnovm/pkg/precompile"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

//type importPath string
//
//type precompileCfg struct {
//	verbose     bool
//	skipFmt     bool
//	skipImports bool
//	gobuild     bool
//	goBinary    string
//	gofmtBinary string
//	output      string
//}
//
//func NewPrecompileCfg(goBuild bool, goBinary string) *precompileCfg {
//	return &precompileCfg{gobuild: goBuild, goBinary: goBinary}
//}
//
//type precompileOptions struct {
//	cfg *precompileCfg
//	// precompiled is the set of packages already
//	// precompiled from .gno to .go.
//	precompiled map[importPath]struct{}
//}
//
//var defaultPrecompileCfg = &precompileCfg{
//	verbose:  false,
//	goBinary: "go",
//}
//
//func NewPrecompileOptions(cfg *precompileCfg) *precompileOptions {
//	return &precompileOptions{cfg, map[importPath]struct{}{}}
//}
//
//func (p *precompileOptions) getFlags() *precompileCfg {
//	return p.cfg
//}
//
//func (p *precompileOptions) isPrecompiled(pkg importPath) bool {
//	_, precompiled := p.precompiled[pkg]
//	return precompiled
//}
//
//func (p *precompileOptions) markAsPrecompiled(pkg importPath) {
//	p.precompiled[pkg] = struct{}{}
//}

func newPrecompileCmd(io commands.IO) *commands.Command {
	cfg := &precompile.PrecompileCfg{}

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

//// TODO: add clean
//func (c *precompileCfg) RegisterFlags(fs *flag.FlagSet) {
//	fs.BoolVar(
//		&c.verbose,
//		"verbose",
//		false,
//		"verbose output when running",
//	)
//
//	fs.BoolVar(
//		&c.skipFmt,
//		"skip-fmt",
//		false,
//		"do not check syntax of generated .go files",
//	)
//
//	fs.BoolVar(
//		&c.skipImports,
//		"skip-imports",
//		false,
//		"do not precompile imports recursively",
//	)
//
//	fs.BoolVar(
//		&c.gobuild,
//		"gobuild",
//		false,
//		"run go build on generated go files, ignoring test files",
//	)
//
//	fs.StringVar(
//		&c.goBinary,
//		"go-binary",
//		"go",
//		"go binary to use for building",
//	)
//
//	fs.StringVar(
//		&c.gofmtBinary,
//		"go-fmt-binary",
//		"gofmt",
//		"gofmt binary to use for syntax checking",
//	)
//
//	fs.StringVar(
//		&c.output,
//		"output",
//		".",
//		"output directory",
//	)
//}

func execPrecompile(cfg *precompile.PrecompileCfg, args []string, io commands.IO) error {
	fmt.Println("---execPrecompile")
	if len(args) < 1 {
		return flag.ErrHelp
	}

	var srcPaths []string
	var opts *precompile.PrecompileOptions

	// clear generated files
	defer func() {
		for _, srcPath := range srcPaths {
			fmt.Println("---clean dir:", srcPath)
			err := precompile.CleanGeneratedFiles(srcPath)
			if err != nil {
				panic(err)
			}
		}
		for pkgPath := range opts.Precompiled {
			fmt.Println("precompiled import pkg:", pkgPath)
			fmt.Println("---clean dir:", pkgPath)
			err := precompile.CleanGeneratedFiles(string(pkgPath))
			if err != nil {
				panic(err)
			}
		}
	}()

	// precompile .gno files.
	srcPaths, err := precompile.GnoFilesFromArgs(args)
	if err != nil {
		return fmt.Errorf("list paths: %w", err)
	}

	opts = precompile.NewPrecompileOptions(cfg)
	fmt.Printf("---opts.Precompiled: %v, opts.cfg: %v \n", opts.Precompiled, opts.Cfg)
	errCount := 0
	for _, filepath := range srcPaths {
		err = precompile.PrecompileFile(filepath, opts)
		if err != nil {
			err = fmt.Errorf("%s: precompile: %w", filepath, err)
			io.ErrPrintfln("%s", err.Error())
			errCount++
		}
	}

	if errCount > 0 {
		return fmt.Errorf("%d precompile errors", errCount)
	}

	if cfg.Gobuild {
		paths, err := precompile.GnoPackagesFromArgs(args)
		if err != nil {
			return fmt.Errorf("list packages: %w", err)
		}

		errCount = 0
		for _, pkgPath := range paths {
			_ = pkgPath
			err = precompile.GoBuildFileOrPkg(pkgPath, cfg)
			if err != nil {
				err = fmt.Errorf("%s: build pkg: %w", pkgPath, err)
				io.ErrPrintfln("%s\n", err.Error())
				errCount++
			}
		}
		if errCount > 0 {
			return fmt.Errorf("%d build errors", errCount)
		}
	}

	return nil
}
