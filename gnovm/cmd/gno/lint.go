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
	"github.com/gnolang/gno/gnovm/pkg/packages"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"go.uber.org/multierr"
)

type lintCfg struct {
	verbose      bool
	rootDir      string
	rootExamples bool
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
	fs.BoolVar(&c.rootExamples, "root-examples", false, "use the examples present in GNOROOT rather than downloading them")
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
	location := i.Location
	wd, err := os.Getwd()
	if err == nil {
		location, err = filepath.Rel(wd, i.Location)
		if err != nil {
			location = i.Location
		} else {
			location = fmt.Sprintf(".%c%s", filepath.Separator, filepath.Join(location))
		}
	}
	// TODO: consider crafting a doc URL based on Code.
	return fmt.Sprintf("%s: %s (code=%d)", location, i.Msg, i.Code)
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

	loadCfg := &packages.LoadConfig{IO: io, Fetcher: testPackageFetcher}
	if cfg.rootExamples {
		loadCfg.GnorootExamples = true
	}

	pkgs, err := packages.Load(loadCfg, args...)
	if err != nil {
		return fmt.Errorf("list packages from args: %w", err)
	}

	pkgsMap := map[string]*packages.Package{}
	packages.Inject(pkgsMap, pkgs)

	hasError := false

	bs, ts := test.Store(
		rootDir, pkgsMap, false,
		nopReader{}, goio.Discard, goio.Discard,
	)

	for _, pkg := range pkgs {
		logName := pkg.ImportPath
		if logName == "" {
			logName = pkg.Dir
		}

		if verbose {
			io.ErrPrintln(logName)
		}

		// Check if 'gno.mod' exists
		if pkg.Root == "" {
			issue := lintIssue{
				Code:       lintGnoMod,
				Confidence: 1,
				Location:   pkg.Dir,
				Msg:        "gno.mod file not found in current or any parent directory",
			}
			io.ErrPrintln(issue)
			hasError = true
		}

		// load deps
		loadDepsCfg := *loadCfg
		loadDepsCfg.Deps = true
		loadDepsCfg.Cache = pkgsMap
		deps, loadDepsErr := packages.Load(&loadDepsCfg, pkg.Dir)
		if loadDepsErr != nil {
			io.ErrPrintln(issueFromError(pkg.Dir, err).String())
			hasError = true
			continue
		}
		packages.Inject(pkgsMap, deps)

		// read mempkg
		memPkgPath := pkg.ImportPath
		if memPkgPath == "" {
			memPkgPath = pkg.Dir
		}
		memPkg, err := gno.ReadMemPackage(pkg.Dir, memPkgPath)
		if err != nil {
			io.ErrPrintln(issueFromError(pkg.Dir, err).String())
			hasError = true
			continue
		}

		// Perform imports using the parent store.
		if err := test.LoadImports(ts, memPkg); err != nil {
			io.ErrPrintln(issueFromError(pkg.Dir, err).String())
			hasError = true
			continue
		}

		// Handle runtime errors
		hasRuntimeErr := catchRuntimeError(pkg.Dir, io.Err(), func() {
			// Wrap in cache wrap so execution of the linter doesn't impact
			// other packages.
			cw := bs.CacheWrap()
			gs := ts.BeginTransaction(cw, cw, nil)

			// Run type checking
			if !pkg.Draft {
				foundErr, err := lintTypeCheck(io, pkg.Dir, memPkg, gs)
				if err != nil {
					io.ErrPrintln(err)
					hasError = true
				} else if foundErr {
					hasError = true
				}
			} else if verbose {
				io.ErrPrintfln("%s: module is draft, skipping type check", logName)
			}

			tm := test.Machine(gs, goio.Discard, memPkg.Path)
			defer tm.Release()

			// Check package
			tm.RunMemPackage(memPkg, true)

			// Check test files
			testFiles := lintTestFiles(memPkg)

			tm.RunFiles(testFiles.Files...)
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

func lintTypeCheck(io commands.IO, pkgDir string, memPkg *gnovm.MemPackage, testStore gno.Store) (errorsFound bool, err error) {
	tcErr := gno.TypeCheckMemPackageTest(pkgDir, memPkg, testStore)
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

func lintTestFiles(memPkg *gnovm.MemPackage) *gno.FileSet {
	testfiles := &gno.FileSet{}
	for _, mfile := range memPkg.Files {
		if !strings.HasSuffix(mfile.Name, ".gno") {
			continue // Skip non-GNO files
		}

		n, _ := gno.ParseFile(mfile.Name, mfile.Body)
		if n == nil {
			continue // Skip empty files
		}

		// XXX: package ending with `_test` is not supported yet
		if strings.HasSuffix(mfile.Name, "_test.gno") && !strings.HasSuffix(string(n.PkgName), "_test") {
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

type nopReader struct{}

func (nopReader) Read(p []byte) (int, error) { return 0, goio.EOF }
