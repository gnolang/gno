package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"go/scanner"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/tests"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
	"go.uber.org/multierr"
)

type lintCfg struct {
	verbose       bool
	rootDir       string
	setExitStatus int
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
	fs.IntVar(&c.setExitStatus, "set-exit-status", 1, "set exit status to 1 if any issues are found")
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
	return fmt.Sprintf("%s: %s (code=%d).", i.Location, i.Msg, i.Code)
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

	pkgPaths, err := gnoPackagesFromArgs(args)
	if err != nil {
		return fmt.Errorf("list packages from args: %w", err)
	}

	hasError := false

	for _, pkgPath := range pkgPaths {
		if verbose {
			io.ErrPrintln(pkgPath)
		}

		info, err := os.Stat(pkgPath)
		if err == nil && !info.IsDir() {
			pkgPath = filepath.Dir(pkgPath)
		}

		// Check if 'gno.mod' exists
		gmFile, err := gnomod.ParseAt(pkgPath)
		if err != nil {
			issue := lintIssue{
				Code:       lintGnoMod,
				Confidence: 1,
				Location:   pkgPath,
				Msg:        err.Error(),
			}
			io.ErrPrintln(issue)
			hasError = true
		}

		// Handle runtime errors
		hasError = catchRuntimeError(pkgPath, io.Err(), func() {
			stdout, stdin, stderr := io.Out(), io.In(), io.Err()

			testStore := tests.TestStore(
				rootDir, "",
				stdin, stdout, stderr,
				tests.ImportModeStdlibsOnly,
			)

			memPkg := gno.ReadMemPackage(pkgPath, pkgPath)

			// Run type checking
			if gmFile == nil || !gmFile.Draft {
				foundErr, err := lintTypeCheck(io, memPkg, testStore)
				if err != nil {
					io.ErrPrintln(err)
					hasError = true
				} else {
					hasError = foundErr || hasError
				}
			} else if verbose {
				io.ErrPrintfln("%s: module is draft, skipping type check\n", pkgPath)
			}

			tm := tests.TestMachine(testStore, stdout, memPkg.Name)

			// Check package
			tm.RunMemPackage(memPkg, true)

			// Check test files
			testFiles := lintTestFiles(memPkg)

			tm.RunFiles(testFiles.Files...)
		}) || hasError
	}

	if hasError && cfg.setExitStatus != 0 {
		os.Exit(cfg.setExitStatus)
	}

	return nil
}

func lintTypeCheck(io commands.IO, memPkg *std.MemPackage, testStore gno.Store) (errorsFound bool, err error) {
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
			//nolint
			return false, fmt.Errorf("unexpected error type: %T", err)
		}
	}
	return true, nil
}

func lintTestFiles(memPkg *std.MemPackage) *gno.FileSet {
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
var reParseRecover = regexp.MustCompile(`^([^:]+):(\d+)(?::\d+)?:? *(.*)$`)

func catchRuntimeError(pkgPath string, stderr io.WriteCloser, action func()) (hasError bool) {
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
			fmt.Fprint(stderr, issueFromError(pkgPath, err).String()+"\n")
		case scanner.ErrorList:
			for _, err := range verr {
				fmt.Fprint(stderr, issueFromError(pkgPath, err).String()+"\n")
			}
		case error:
			fmt.Fprint(stderr, issueFromError(pkgPath, verr).String()+"\n")
		case string:
			fmt.Fprint(stderr, issueFromError(pkgPath, errors.New(verr)).String()+"\n")
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
	if len(matches) == 4 {
		sourcepath := guessSourcePath(pkgPath, matches[1])
		issue.Location = fmt.Sprintf("%s:%s", sourcepath, matches[2])
		issue.Msg = strings.TrimSpace(matches[3])
	} else {
		issue.Location = fmt.Sprintf("%s:0", filepath.Clean(pkgPath))
		issue.Msg = err.Error()
	}
	return issue
}
