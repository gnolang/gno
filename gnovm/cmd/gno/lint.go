package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/tests"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
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
	addIssue := func(issue lintIssue) {
		hasError = true
		fmt.Fprint(io.Err(), issue.String()+"\n")
	}

	for _, pkgPath := range pkgPaths {
		if verbose {
			fmt.Fprintf(io.Err(), "Linting %q...\n", pkgPath)
		}

		// Check if 'gno.mod' exists
		gnoModPath := filepath.Join(pkgPath, "gno.mod")
		if !osm.FileExists(gnoModPath) {
			addIssue(lintIssue{
				Code:       lintNoGnoMod,
				Confidence: 1,
				Location:   pkgPath,
				Msg:        "missing 'gno.mod' file",
			})
		}

		// Handle runtime errors
		catchRuntimeError(pkgPath, addIssue, func() {
			stdout, stdin, stderr := io.Out(), io.In(), io.Err()
			testStore := tests.TestStore(
				rootDir, "",
				stdin, stdout, stderr,
				tests.ImportModeStdlibsOnly,
			)

			targetPath := pkgPath
			info, err := os.Stat(pkgPath)
			if err == nil && !info.IsDir() {
				targetPath = filepath.Dir(pkgPath)
			}

			memPkg := gno.ReadMemPackage(targetPath, targetPath)
			tm := tests.TestMachine(testStore, stdout, memPkg.Name)

			// Check package
			tm.RunMemPackage(memPkg, true)

			// Check test files
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

			tm.RunFiles(testfiles.Files...)
		})

		// TODO: Add more checkers
	}

	if hasError && cfg.setExitStatus != 0 {
		os.Exit(cfg.setExitStatus)
	}

	return nil
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

func catchRuntimeError(pkgPath string, addIssue func(issue lintIssue), action func()) {
	defer func() {
		// Errors catched here mostly come from: gnovm/pkg/gnolang/preprocess.go
		r := recover()
		if r == nil {
			return
		}

		var err error
		switch verr := r.(type) {
		case *gno.PreprocessError:
			err = verr.Unwrap()
		case error:
			err = verr
		case string:
			err = errors.New(verr)
		default:
			panic(r)
		}

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

		addIssue(issue)
	}()

	action()
}

type lintCode int

const (
	lintUnknown  lintCode = 0
	lintNoGnoMod lintCode = iota
	lintGnoError

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
