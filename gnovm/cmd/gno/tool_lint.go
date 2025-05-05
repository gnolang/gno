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

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"go.uber.org/multierr"
)

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

	dirPaths, err := gnoPackagesFromArgsRecursively(args)
	if err != nil {
		return fmt.Errorf("list packages from args: %w", err)
	}

	hasError := false

	bs, ts := test.StoreWithOptions(
		rootDir, goio.Discard,
		test.StoreOptions{PreprocessOnly: true},
	)

	for _, dirPath := range dirPaths {
		if verbose {
			io.ErrPrintln(dirPath)
		}

		info, err := os.Stat(dirPath)
		if err == nil && !info.IsDir() {
			dirPath = filepath.Dir(dirPath)
		}

		// Check if 'gno.mod' exists
		gmFile, err := gnomod.ParseAt(dirPath)
		if err != nil {
			issue := lintIssue{
				Code:       lintGnoMod,
				Confidence: 1,
				Location:   dirPath,
				Msg:        err.Error(),
			}
			io.ErrPrintln(issue)
			hasError = true
		}

		pkgPath, _ := determinePkgPath(gmFile, dirPath, cfg.rootDir)
		memPkg, err := gno.ReadMemPackage(dirPath, pkgPath)
		if err != nil {
			io.ErrPrintln(issueFromError(dirPath, err).String())
			hasError = true
			continue
		}

		// Perform imports using the parent store.
		if err := test.LoadImports(ts, memPkg); err != nil {
			io.ErrPrintln(issueFromError(dirPath, err).String())
			hasError = true
			continue
		}

		// Handle runtime errors
		hasRuntimeErr := catchRuntimeError(dirPath, io.Err(), func() {
			// Wrap in cache wrap so execution of the linter doesn't impact
			// other packages.
			cw := bs.CacheWrap()
			gs := ts.BeginTransaction(cw, cw, nil)

			// Run type checking
			if gmFile == nil || !gmFile.Draft {
				foundErr, err := lintTypeCheck(io, memPkg, gs)
				if err != nil {
					io.ErrPrintln(err)
					hasError = true
				} else if foundErr {
					hasError = true
				}
			} else if verbose {
				io.ErrPrintfln("%s: module is draft, skipping type check", dirPath)
			}

			tm := test.Machine(gs, goio.Discard, memPkg.Path, false)

			defer tm.Release()

			// Check test files
			packageFiles := sourceAndTestFileset(memPkg)

			tm.PreprocessFiles(memPkg.Name, memPkg.Path, packageFiles, false, false)
		})
		if hasRuntimeErr {
			hasError = true
		}
	}

	if hasError {
		return commands.ExitCodeError(1)
	}

	return nil
}

func lintTypeCheck(io commands.IO, memPkg *gnovm.MemPackage, testStore gno.Store) (errorsFound bool, err error) {
	tcErr := gno.TypeCheckMemPackageTest(memPkg, testStore)
	if tcErr == nil {
		return false, nil
	}

	errs := multierr.Errors(tcErr)
	for _, err := range errs {
		switch err := err.(type) {
		case types.Error:
			io.ErrPrintln(lintIssue{
				Code:       lintTypeCheckError,
				Msg:        err.Msg,
				Confidence: 1,
				Location:   err.Fset.Position(err.Pos).String(),
			})
		case scanner.ErrorList:
			for _, scErr := range err {
				io.ErrPrintln(lintIssue{
					Code:       lintParserError,
					Msg:        scErr.Msg,
					Confidence: 1,
					Location:   scErr.Pos.String(),
				})
			}
		case scanner.Error:
			io.ErrPrintln(lintIssue{
				Code:       lintParserError,
				Msg:        err.Msg,
				Confidence: 1,
				Location:   err.Pos.String(),
			})
		default:
			return false, fmt.Errorf("unexpected error type: %T", err)
		}
	}
	return true, nil
}

func sourceAndTestFileset(memPkg *gnovm.MemPackage) *gno.FileSet {
	testfiles := &gno.FileSet{}
	for _, mfile := range memPkg.Files {
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

func catchRuntimeError(pkgPath string, stderr goio.WriteCloser, action func()) (hasError bool) {
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
			fmt.Fprintln(stderr, issueFromError(pkgPath, err).String())
		case error:
			errors := multierr.Errors(verr)
			for _, err := range errors {
				errList, ok := err.(scanner.ErrorList)
				if ok {
					for _, errorInList := range errList {
						fmt.Fprintln(stderr, issueFromError(pkgPath, errorInList).String())
					}
				} else {
					fmt.Fprintln(stderr, issueFromError(pkgPath, err).String())
				}
			}
		case string:
			fmt.Fprintln(stderr, issueFromError(pkgPath, errors.New(verr)).String())
		default:
			panic(r)
		}
	}()

	action()
	return
}

func issueFromError(pkgPath string, err error) lintIssue {
	var issue lintIssue
	issue.Confidence = 1
	issue.Code = lintGnoError

	parsedError := strings.TrimSpace(err.Error())
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
