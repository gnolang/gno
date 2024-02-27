package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/scanner"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/transpiler"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type transpileCfg struct {
	verbose     bool
	rootDir     string
	skipImports bool
	gobuild     bool
	goBinary    string
	output      string
}

type transpileOptions struct {
	cfg *transpileCfg
	// transpiled is the set of packages already
	// transpiled from .gno to .go.
	transpiled map[string]struct{}
	// skipped packages (gno mod marks them as draft)
	skipped []string
}

var defaultTranspileCfg = &transpileCfg{
	verbose:  false,
	goBinary: "go",
}

func newTranspileOptions(cfg *transpileCfg) *transpileOptions {
	return &transpileOptions{
		cfg:        cfg,
		transpiled: map[string]struct{}{},
	}
}

func (p *transpileOptions) getFlags() *transpileCfg {
	return p.cfg
}

func (p *transpileOptions) isTranspiled(pkg string) bool {
	_, transpiled := p.transpiled[pkg]
	return transpiled
}

func (p *transpileOptions) markAsTranspiled(pkg string) {
	p.transpiled[pkg] = struct{}{}
}

func newTranspileCmd(io commands.IO) *commands.Command {
	cfg := &transpileCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "transpile",
			ShortUsage: "transpile [flags] <package> [<package>...]",
			ShortHelp:  "Transpiles .gno files to .go",
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
		"verbose",
		false,
		"verbose output when running",
	)

	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"",
		"clone location of github.com/gnolang/gno (gno tries to guess it)",
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

	// guess cfg.RootDir
	if cfg.rootDir == "" {
		cfg.rootDir = gnoenv.RootDir()
	}

	// transpile .gno packages and files.
	paths, err := gnoPackagesFromArgs(args)
	if err != nil {
		return fmt.Errorf("list paths: %w", err)
	}

	opts := newTranspileOptions(cfg)
	var errlist scanner.ErrorList
	for _, path := range paths {
		st, err := os.Stat(path)
		if err != nil {
			return err
		}
		if st.IsDir() {
			err = transpilePkg(path, opts)
		} else {
			if opts.cfg.verbose {
				fmt.Fprintf(os.Stderr, "%s\n", filepath.Clean(path))
			}

			err = transpileFile(path, opts)
		}
		if err != nil {
			var fileErrlist scanner.ErrorList
			if !errors.As(err, &fileErrlist) {
				// Not an scanner.ErrorList: return immediately.
				return fmt.Errorf("%s: transpile: %w", path, err)
			}
			errlist = append(errlist, fileErrlist...)
		}
	}

	if errlist.Len() == 0 && cfg.gobuild {
		for _, pkgPath := range paths {
			if slices.Contains(opts.skipped, pkgPath) {
				continue
			}
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

// transpilePkg transpiles all non-test files at the given location.
// Additionally, it checks the gno.mod in said location, and skips it if it is
// a draft module
func transpilePkg(dirPath string, opts *transpileOptions) error {
	if opts.isTranspiled(dirPath) {
		return nil
	}
	opts.markAsTranspiled(dirPath)

	gmod, err := gnomod.ParseAt(dirPath)
	if err != nil && !errors.Is(err, gnomod.ErrGnoModNotFound) {
		return err
	}
	if err == nil && gmod.Draft {
		if opts.cfg.verbose {
			fmt.Fprintf(os.Stderr, "%s (skipped, gno.mod marks module as draft)\n", filepath.Clean(dirPath))
		}
		opts.skipped = append(opts.skipped, dirPath)
		return nil
	}

	// XXX(morgan): Currently avoiding test files as they contain imports like "fmt".
	// The transpiler doesn't currently support "test stdlibs", and even if it
	// did all packages like "fmt" would have to exist as standard libraries to work.
	// Easier to skip for now.
	files, err := listNonTestFiles(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	if opts.cfg.verbose {
		fmt.Fprintf(os.Stderr, "%s\n", filepath.Clean(dirPath))
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

	// parse .gno.
	source, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	// compute attributes based on filename.
	targetFilename, tags := transpiler.TranspiledFilenameAndTags(srcPath)

	// preprocess.
	transpileRes, err := transpiler.Transpile(string(source), tags, srcPath)
	if err != nil {
		return fmt.Errorf("transpile: %w", err)
	}

	// resolve target path
	var targetPath string
	if flags.output != "." {
		path, err := ResolvePath(flags.output, filepath.Dir(srcPath))
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

	// transpile imported packages, if `SkipImports` sets to false
	if !flags.skipImports &&
		!strings.HasSuffix(srcPath, "_filetest.gno") && !strings.HasSuffix(srcPath, "_test.gno") {
		dirPaths := getPathsFromImportSpec(opts.cfg.rootDir, transpileRes.Imports)
		for _, path := range dirPaths {
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
		fmt.Fprintf(os.Stderr, "%s [build]\n", fileOrPkg)
	}

	return transpiler.TranspileBuildPackage(fileOrPkg, goBinary)
}

// getPathsFromImportSpec returns the directory paths where the code for each
// importSpec is stored (assuming they start with [transpiler.ImportPrefix]).
func getPathsFromImportSpec(rootDir string, importSpec []*ast.ImportSpec) (dirs []string) {
	for _, i := range importSpec {
		path, err := strconv.Unquote(i.Path.Value)
		if err != nil {
			continue
		}
		if strings.HasPrefix(path, transpiler.ImportPrefix) {
			res := strings.TrimPrefix(path, transpiler.ImportPrefix)

			dirs = append(dirs, rootDir+filepath.FromSlash(res))
		}
	}
	return
}
