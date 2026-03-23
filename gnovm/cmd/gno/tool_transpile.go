package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/scanner"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	// CLI output
	io commands.IO
	// transpiled is the set of packages already
	// transpiled from .gno to .go.
	transpiled map[string]struct{}
	// skipped packages (gno mod marks them as ignore)
	skipped []string
}

func newTranspileOptions(cfg *transpileCfg, io commands.IO) *transpileOptions {
	return &transpileOptions{
		cfg:        cfg,
		io:         io,
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
	paths, err := gnoFilesFromArgsRecursively(args)
	if err != nil {
		return fmt.Errorf("list paths: %w", err)
	}

	opts := newTranspileOptions(cfg, io)
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
				io.ErrPrintln(filepath.Clean(path))
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
			if cfg.output != "." {
				if pkgPath, err = ResolvePath(cfg.output, pkgPath); err != nil {
					return fmt.Errorf("resolve output path: %w", err)
				}
			}
			err := goBuildFileOrPkg(io, pkgPath, cfg)
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
// Additionally, it checks the gnomod.toml in said location, and skips it if it is
// a ignore module
func transpilePkg(dirPath string, opts *transpileOptions) error {
	if opts.isTranspiled(dirPath) {
		return nil
	}
	opts.markAsTranspiled(dirPath)

	gmod, err := gnomod.ParseDir(dirPath)
	if err != nil && !errors.Is(err, gnomod.ErrNoModFile) {
		return err
	}
	if err == nil && gmod.Ignore {
		if opts.cfg.verbose {
			opts.io.ErrPrintfln("%s (skipped, gnomod.toml marks module as ignored)", filepath.Clean(dirPath))
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
		return err
	}

	if opts.cfg.verbose {
		opts.io.ErrPrintln(filepath.Clean(dirPath))
	}
	for _, file := range files {
		if err := transpileFile(file, opts); err != nil {
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
		dirPaths, err := getPathsFromImportSpec(opts.cfg.rootDir, transpileRes.Imports)
		if err != nil {
			return err
		}
		for _, path := range dirPaths {
			if err := transpilePkg(path, opts); err != nil {
				return err
			}
		}
	}

	return nil
}

func goBuildFileOrPkg(io commands.IO, fileOrPkg string, cfg *transpileCfg) error {
	verbose := cfg.verbose
	goBinary := cfg.goBinary

	if verbose {
		io.ErrPrintfln("%s [build]", filepath.Clean(fileOrPkg))
	}

	return buildTranspiledPackage(fileOrPkg, goBinary)
}

// getPathsFromImportSpec returns the directory paths where the code for each
// importSpec is stored (assuming they start with [transpiler.ImportPrefix]).
func getPathsFromImportSpec(rootDir string, importSpec []*ast.ImportSpec) (dirs []string, err error) {
	for _, i := range importSpec {
		path, err := strconv.Unquote(i.Path.Value)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(path, transpiler.ImportPrefix) {
			res := strings.TrimPrefix(path, transpiler.ImportPrefix)

			dirs = append(dirs, rootDir+filepath.FromSlash(res))
		}
	}
	return
}

// buildTranspiledPackage tries to run `go build` against the transpiled .go files.
//
// This method is the most efficient to detect errors but requires that
// all the import are valid and available.
func buildTranspiledPackage(fileOrPkg, goBinary string) error {
	// TODO: use cmd/compile instead of exec?
	// TODO: find the nearest go.mod file, chdir in the same folder, trim prefix?
	// TODO: temporarily create an in-memory go.mod or disable go modules for gno?
	// TODO: ignore .go files that were not generated from gno?

	info, err := os.Stat(fileOrPkg)
	if err != nil {
		return fmt.Errorf("invalid file or package path %s: %w", fileOrPkg, err)
	}
	var (
		target string
		chdir  string
	)
	if !info.IsDir() {
		dstFilename, _ := transpiler.TranspiledFilenameAndTags(fileOrPkg)
		// Makes clear to go compiler that this is a relative path,
		// rather than a path to a package/module.
		// can't use filepath.Join as it cleans its results.
		target = filepath.Dir(fileOrPkg) + string(filepath.Separator) + dstFilename
	} else {
		if info.Name() == "filetests" {
			// We don't transpile filetest files, so we will get the error "no Go files in dir"
			return nil
		}
		// Go does not allow building packages using absolute paths, and requires
		// relative paths to always be prefixed with `./` (because the argument
		// go expects are import paths, not directories).
		// To circumvent this, we use the -C flag to chdir into the right
		// directory, then run `go build .`
		chdir = fileOrPkg
		target = "."
	}

	// pre-alloc max 5 args
	args := append(make([]string, 0, 5), "build")
	if chdir != "" {
		args = append(args, "-C", chdir)
	}
	args = append(args, "-tags=gno", target)
	cmd := exec.Command(goBinary, args...)
	out, err := cmd.CombinedOutput()
	if errors.As(err, new(*exec.ExitError)) {
		// there was a non-zero exit code; parse the go build errors
		return parseGoBuildErrors(string(out))
	}
	// other kinds of errors; return
	return err
}

var (
	reGoBuildError   = regexp.MustCompile(`(?m)^(\S+):(\d+):(\d+): (.+)$`)
	reGoBuildComment = regexp.MustCompile(`(?m)^#.*$`)
)

// parseGoBuildErrors returns a scanner.ErrorList filled with all errors found
// in out, which is supposed to be the output of the `go build` command.
//
// TODO(tb): update when `go build -json` is released to replace regexp usage.
// See https://github.com/golang/go/issues/62067
func parseGoBuildErrors(out string) error {
	var errList scanner.ErrorList
	matches := reGoBuildError.FindAllStringSubmatch(out, -1)
	for _, match := range matches {
		filename := match[1]
		line, err := strconv.Atoi(match[2])
		if err != nil {
			return fmt.Errorf("parse line go build error %s: %w", match, err)
		}

		column, err := strconv.Atoi(match[3])
		if err != nil {
			return fmt.Errorf("parse column go build error %s: %w", match, err)
		}
		msg := match[4]
		errList.Add(token.Position{
			Filename: filename,
			Line:     line,
			Column:   column,
		}, msg)
	}

	replaced := reGoBuildError.ReplaceAllLiteralString(out, "")
	replaced = reGoBuildComment.ReplaceAllString(replaced, "")
	replaced = strings.TrimSpace(replaced)
	if replaced != "" {
		errList.Add(token.Position{}, "Additional go build errors:\n"+replaced)
	}

	return errList.Err()
}
