package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"go/scanner"
	"go/types"
	goio "io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
	"go.uber.org/multierr"
)

type processedPackage struct {
	mpkg *std.MemPackage
	fset *gno.FileSet
	pn   *gno.PackageNode
}

type lintCfg struct {
	verbose bool
	rootDir string
	// min_confidence: minimum confidence of a problem to print it (default 0.8)
	// auto-fix: apply suggested fixes automatically.
}

func newLintCmd(io commands.IO) *commands.Command {
	cfg := &lintCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "lint",
			ShortUsage: "lint [flags] <package> [<package>...]",
			ShortHelp:  "runs the linter for the specified packages",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execLint(cfg, args, io)
		},
	)
}

func (c *lintCfg) RegisterFlags(fs *flag.FlagSet) {
	rootdir := gnoenv.RootDir()

	fs.BoolVar(&c.verbose, "v", false, "verbose output when lintning")
	fs.StringVar(&c.rootDir, "root-dir", rootdir, "clone location of github.com/gnolang/gno (gno tries to guess it)")
}

type lintCode int

const (
	lintUnknown lintCode = iota
	lintGnoMod
	lintGnoError
	lintParserError
	lintTypeCheckError

	// TODO: add new linter codes here.
)

type lintIssue struct {
	Code       lintCode
	Msg        string
	Confidence float64 // 1 is 100%
	Location   string  // file:line, or equivalent
	// TODO: consider writing fix suggestions
}

func (i lintIssue) String() string {
	// TODO: consider crafting a doc URL based on Code.
	return fmt.Sprintf("%s: %s (code=%d)", i.Location, i.Msg, i.Code)
}

func execLint(cfg *lintCfg, args []string, io commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	var (
		verbose = cfg.verbose
		rootDir = cfg.rootDir
	)
	if rootDir == "" {
		rootDir = gnoenv.RootDir()
	}

	dirs, err := gnoPackagesFromArgsRecursively(args)
	if err != nil {
		return fmt.Errorf("list packages from args: %w", err)
	}

	hasError := false

	bs, ts := test.StoreWithOptions(
		rootDir, goio.Discard,
		test.StoreOptions{PreprocessOnly: true},
	)

	ppkgs := map[string]processedPackage{}

	//----------------------------------------
	// STEP 1: PREPROCESS ALL FILES
	for _, dir := range dirs {
		if verbose {
			io.ErrPrintln(dir)
		}

		info, err := os.Stat(dir)
		if err == nil && !info.IsDir() {
			dir = filepath.Dir(dir)
		}

		// Check if 'gno.mod' exists
		gmFile, err := gnomod.ParseAt(dir)
		if err != nil {
			issue := lintIssue{
				Code:       lintGnoMod,
				Confidence: 1,
				Location:   dir,
				Msg:        err.Error(),
			}
			io.ErrPrintln(issue)
			hasError = true
		}

		pkgPath, _ := determinePkgPath(gmFile, dir, cfg.rootDir)
		mpkg, err := gno.ReadMemPackage(dir, pkgPath)
		if err != nil {
			io.ErrPrintln(issueFromError(dir, pkgPath, err).String())
			hasError = true
			continue
		}

		// Perform imports using the parent store.
		if err := test.LoadImports(ts, mpkg); err != nil {
			io.ErrPrintln(issueFromError(dir, pkgPath, err).String())
			hasError = true
			continue
		}

		// Handle runtime errors
		hasRuntimeErr := catchRuntimeError(dir, pkgPath, io.Err(), func() {
			// Wrap in cache wrap so execution of the linter doesn't impact
			// other packages.
			cw := bs.CacheWrap()
			gs := ts.BeginTransaction(cw, cw, nil)

			// Run type checking
			if gmFile == nil || !gmFile.Draft {
				foundErr, err := lintTypeCheck(io, dir, mpkg, gs)
				if err != nil {
					io.ErrPrintln(err)
					hasError = true
				} else if foundErr {
					hasError = true
				}
			} else if verbose {
				io.ErrPrintfln("%s: module is draft, skipping type check", dir)
			}

			tm := test.Machine(gs, goio.Discard, pkgPath, false)

			defer tm.Release()

			// Check test files
			packageFiles := sourceAndTestFileset(mpkg)

			pn, _ := tm.PreprocessFiles(mpkg.Name, mpkg.Path, packageFiles, false, false)
			ppkgs[mpkg.Path] = processedPackage{mpkg, packageFiles, pn}

			// Use the preprocessor to collect the transformations needed to be done.
			// They are collected in pn.GetAttribute("XREALMITEM")
			for _, fn := range packageFiles.Files {
				gno.FindGno0p9XItems(gs, pn, fn)
			}
		})
		if hasRuntimeErr {
			hasError = true
		}
	}

	//----------------------------------------
	// STEP 2: TRANSFORM FOR Gno 0.9
	for _, dir := range dirs {
		ppkg, ok := ppkgs[dir]
		if !ok {
			panic("where did it go")
		}
		mpkg, pn := ppkg.mpkg, ppkg.pn
		xform := pn.GetAttribute(gno.ATTR_GNO0P9_XITEMS).(map[string]string)
		err := gno.TranspileToGno0p9(mpkg, dir, true, true, xform)
		if err != nil {
			panic(err)
		}
	}

	if hasError {
		return commands.ExitCodeError(1)
	}

	return nil
}

func lintTypeCheck(io commands.IO, dir string, mpkg *std.MemPackage, testStore gno.Store) (errorsFound bool, err error) {
	tcErr := gno.TypeCheckMemPackageTest(mpkg, testStore)
	if tcErr == nil {
		return false, nil
	}

	errs := multierr.Errors(tcErr)
	for _, err := range errs {
		switch err := err.(type) {
		case types.Error:
			loc := err.Fset.Position(err.Pos).String()
			loc = replaceWithDirPath(loc, mpkg.Path, dir)
			io.ErrPrintln(lintIssue{
				Code:       lintTypeCheckError,
				Msg:        err.Msg,
				Confidence: 1,
				Location:   loc,
			})
		case scanner.ErrorList:
			for _, scErr := range err {
				loc := scErr.Pos.String()
				loc = replaceWithDirPath(loc, mpkg.Path, dir)
				io.ErrPrintln(lintIssue{
					Code:       lintParserError,
					Msg:        scErr.Msg,
					Confidence: 1,
					Location:   loc,
				})
			}
		case scanner.Error:
			loc := err.Pos.String()
			loc = replaceWithDirPath(loc, mpkg.Path, dir)
			io.ErrPrintln(lintIssue{
				Code:       lintParserError,
				Msg:        err.Msg,
				Confidence: 1,
				Location:   loc,
			})
		default:
			return false, fmt.Errorf("unexpected error type: %T", err)
		}
	}
	return true, nil
}

func sourceAndTestFileset(mpkg *std.MemPackage) *gno.FileSet {
	testfiles := &gno.FileSet{}
	for _, mfile := range mpkg.Files {
		if !strings.HasSuffix(mfile.Name, ".gno") {
			continue // Skip non-GNO files
		}

		n := gno.MustParseFile(mfile.Name, mfile.Body)
		if n == nil {
			continue // Skip empty files
		}

		// XXX: package ending with `_test` is not supported yet
		if !strings.HasSuffix(mfile.Name, "_filetest.gno") &&
			!strings.HasSuffix(string(n.PkgName), "_test") {
			// Keep only test files
			testfiles.AddFiles(n)
		}
	}
	return testfiles
}

func guessSourcePath(pkg, source string) string {
	if info, err := os.Stat(pkg); !os.IsNotExist(err) && !info.IsDir() {
		pkg = filepath.Dir(pkg)
	}

	sourceJoin := filepath.Join(pkg, source)
	if _, err := os.Stat(sourceJoin); !os.IsNotExist(err) {
		return filepath.Clean(sourceJoin)
	}

	if _, err := os.Stat(source); !os.IsNotExist(err) {
		return filepath.Clean(source)
	}

	return filepath.Clean(pkg)
}

// reParseRecover is a regex designed to parse error details from a string.
// It extracts the file location, line number, and error message from a formatted error string.
// XXX: Ideally, error handling should encapsulate location details within a dedicated error type.
var reParseRecover = regexp.MustCompile(`^([^:]+)((?::(?:\d+)){1,2}):? *(.*)$`)

func catchRuntimeError(dir, pkgPath string, stderr goio.WriteCloser, action func()) (hasError bool) {
	defer func() {
		// Errors catched here mostly come from: gnovm/pkg/gnolang/preprocess.go
		r := recover()
		if r == nil {
			return
		}
		hasError = true
		switch verr := r.(type) {
		case *gno.PreprocessError:
			err := verr.Unwrap()
			fmt.Fprintln(stderr, issueFromError(dir, pkgPath, err).String())
		case error:
			errors := multierr.Errors(verr)
			for _, err := range errors {
				errList, ok := err.(scanner.ErrorList)
				if ok {
					for _, errorInList := range errList {
						fmt.Fprintln(stderr, issueFromError(dir, pkgPath, errorInList).String())
					}
				} else {
					fmt.Fprintln(stderr, issueFromError(dir, pkgPath, err).String())
				}
			}
		case string:
			fmt.Fprintln(stderr, issueFromError(dir, pkgPath, errors.New(verr)).String())
		default:
			panic(r)
		}
	}()

	action()
	return
}

func issueFromError(dir, pkgPath string, err error) lintIssue {
	var issue lintIssue
	issue.Confidence = 1
	issue.Code = lintGnoError

	parsedError := strings.TrimSpace(err.Error())
	parsedError = replaceWithDirPath(parsedError, pkgPath, dir)
	parsedError = strings.TrimPrefix(parsedError, pkgPath+"/")

	matches := reParseRecover.FindStringSubmatch(parsedError)
	if len(matches) > 0 {
		sourcepath := guessSourcePath(pkgPath, matches[1])
		issue.Location = sourcepath + matches[2]
		issue.Msg = strings.TrimSpace(matches[3])
	} else {
		issue.Location = fmt.Sprintf("%s:0", filepath.Clean(pkgPath))
		issue.Msg = err.Error()
	}
	return issue
}

func replaceWithDirPath(s, pkgPath, dir string) string {
	if strings.HasPrefix(s, pkgPath) {
		return filepath.Clean(dir + s[len(pkgPath):])
	}
	return s
}
