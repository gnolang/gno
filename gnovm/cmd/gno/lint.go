package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/tests"
	"github.com/gnolang/gno/tm2/pkg/commands"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

var reParseRecover = regexp.MustCompile(`^(.+):(\d+): ?(.*)$`)

type lintCfg struct {
	verbose       bool
	rootDir       string
	setExitStatus int
	// min_confidence: minimum confidence of a problem to print it (default 0.8)
	// auto-fix: apply suggested fixes automatically.
}

func newLintCmd(io *commands.IO) *commands.Command {
	cfg := &lintCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "lint",
			ShortUsage: "lint [flags] <package> [<package>...]",
			ShortHelp:  "Runs the linter for the specified packages",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execLint(cfg, args, io)
		},
	)
}

func (c *lintCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&c.verbose, "verbose", false, "verbose output when lintning")
	fs.StringVar(&c.rootDir, "root-dir", "", "clone location of github.com/gnolang/gno (gno tries to guess it)")
	fs.IntVar(&c.setExitStatus, "set_exit_status", 1, "set exit status to 1 if any issues are found")
}

func execLint(cfg *lintCfg, args []string, io *commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	var (
		verbose = cfg.verbose
		rootDir = cfg.rootDir
	)
	if rootDir == "" {
		rootDir = guessRootDir()
	}

	pkgPaths, err := gnoPackagesFromArgs(args)
	if err != nil {
		return fmt.Errorf("list packages from args: %w", err)
	}

	hasError := false
	addIssue := func(issue lintIssue) {
		hasError = true
		fmt.Fprint(io.Err, issue.String()+"\n")
	}

	for _, pkgPath := range pkgPaths {
		if verbose {
			fmt.Fprintf(io.Err, "Linting %q...\n", pkgPath)
		}

		// 'gno.mod' exists?
		{
			gnoModPath := filepath.Join(pkgPath, "gno.mod")
			if !osm.FileExists(gnoModPath) {
				addIssue(lintIssue{
					Code:       lintNoGnoMod,
					Confidence: 1,
					Location:   pkgPath,
					Msg:        "missing 'gno.mod' file",
				})
			}
		}

		// run gno machine to detect basic package errors
		{
			var (
				stdout = io.Out
				stdin  = io.In
				stderr = io.Err

				testStore = tests.TestStore(
					rootDir, "",
					stdin, stdout, stderr,
					tests.ImportModeStdlibsOnly,
				)
			)

			catchError := func() {
				r := recover()
				if r == nil {
					return
				}

				if err, ok := r.(error); ok {
					loc := strings.TrimSpace(err.Error())
					// XXX: this should not happen, loc should not contain package path
					loc = strings.TrimPrefix(loc, pkgPath+"/")
					subm := reParseRecover.FindStringSubmatch(loc)
					if len(subm) > 0 {
						// subm[1]: file
						// subm[2]: line
						// subm[3]: error
						addIssue(lintIssue{
							Code:       lintGnoError,
							Confidence: 1,
							Location:   fmt.Sprintf("%s:%s", subm[1], subm[2]),
							Msg:        strings.TrimSpace(subm[3]),
						})
					}
				}
			}

			// run the machine on the target package
			func() {
				defer catchError()

				// reuse test run files
				memPkg := gno.ReadMemPackage(filepath.Dir(pkgPath), pkgPath)
				m := tests.TestMachine(testStore, stdout, memPkg.Name)

				// check package
				m.RunMemPackage(memPkg, true)

				// check test files
				{
					testfiles := &gno.FileSet{}
					for _, mfile := range memPkg.Files {
						if !strings.HasSuffix(mfile.Name, ".gno") {
							continue // skip this file.
						}

						n, _ := gno.ParseFile(mfile.Name, mfile.Body)
						if n == nil {
							continue
						}

						if strings.HasSuffix(mfile.Name, "_test.gno") {
							testfiles.AddFiles(n)
						}
					}

					m.RunFiles(testfiles.Files...)
				}
			}()
		}

		// XXX: add more checkers
	}

	if hasError && cfg.setExitStatus != 0 {
		os.Exit(cfg.setExitStatus)
	}

	if verbose {
		fmt.Println("no lint errors")
	}

	return nil
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
