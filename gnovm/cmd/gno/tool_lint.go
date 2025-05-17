package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/scanner"
	"go/token"
	"go/types"
	goio "io"
	"io/fs"
	"os"
	"path"
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

/*
	Linting.
	Refer to the [Lint and Transpile ADR](./adr/pr4264_lint_transpile.md).
*/

type processedPackage struct {
	mpkg   *std.MemPackage
	fset   *gno.FileSet
	pn     *gno.PackageNode
	_tests []*gno.FileSet
	ftests []*gno.FileSet
}

type lintCmd struct {
	verbose    bool
	rootDir    string
	autoGnomod bool
	// min_confidence: minimum confidence of a problem to print it
	// (default 0.8) auto-fix: apply suggested fixes automatically.
}

func newLintCmd(io commands.IO) *commands.Command {
	cmd := &lintCmd{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "lint",
			ShortUsage: "lint [flags] <package> [<package>...]",
			ShortHelp:  "runs the linter for the specified packages",
		},
		cmd,
		func(_ context.Context, args []string) error {
			return execLint(cmd, args, io)
		},
	)
}

func (c *lintCmd) RegisterFlags(fs *flag.FlagSet) {
	rootdir := gnoenv.RootDir()

	fs.BoolVar(&c.verbose, "v", false, "verbose output when lintning")
	fs.StringVar(&c.rootDir, "root-dir", rootdir, "clone location of github.com/gnolang/gno (gno tries to guess it)")
	fs.BoolVar(&c.autoGnomod, "auto-gnomod", true, "auto-generate gno.mod file if not already present.")
}

type lintCode string

const (
	lintUnknownError    lintCode = "lintUnknownError"
	lintReadError                = "lintReadError"
	lintImportError              = "lintImportError" // XXX break this out
	lintGnoModError              = "lintGnoModError"
	lintPreprocessError          = "lintPreprocessError"
	lintParserError              = "lintParserError"
	lintTypeCheckError           = "lintTypeCheckError"

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
	return fmt.Sprintf("%s: %s (code=%s)", i.Location, i.Msg, i.Code)
}

func execLint(cmd *lintCmd, args []string, io commands.IO) error {
	// Show a help message by default.
	if len(args) == 0 {
		return flag.ErrHelp
	}

	// Guess opts.RootDir.
	if cmd.rootDir == "" {
		cmd.rootDir = gnoenv.RootDir()
	}

	dirs, err := gnoPackagesFromArgsRecursively(args)
	if err != nil {
		return fmt.Errorf("list packages from args: %w", err)
	}

	hasError := false

	bs, ts := test.StoreWithOptions(
		cmd.rootDir, goio.Discard,
		test.StoreOptions{PreprocessOnly: true},
	)
	ppkgs := map[string]processedPackage{}

	if cmd.verbose {
		io.ErrPrintfln("flinting directories: %v", dirs)
	}
	//----------------------------------------
	// STAGE 1: Lint.
	for _, dir := range dirs {
		if cmd.verbose {
			io.ErrPrintfln("linting %q", dir)
		}

		// XXX Currently the linter only supports linting directories.
		// In order to support linting individual files, we need to
		// refactor this code to work with mempackages, not dirs, and
		// cmd/gno/util.go needs to be refactored to return mempackages
		// rather than dirs. Commands like `gno lint a.gno b.gno`
		// should create a temporary package from just those files. We
		// could also load mempackages lazily for memory efficiency.
		info, err := os.Stat(dir)
		if err == nil && !info.IsDir() {
			dir = filepath.Dir(dir)
		}

		// Read and parse gno.mod directly.
		fpath := path.Join(dir, "gno.mod")
		mod, err := gnomod.ParseFilepath(fpath)
		if errors.Is(err, fs.ErrNotExist) {
			if cmd.autoGnomod {
				modstr := gno.GenGnoModDefault("gno.land/r/xxx_myrealm_xxx/xxx_fixme_xxx")
				mod, err = gnomod.ParseBytes("gno.mod", []byte(modstr))
				if err != nil {
					panic(fmt.Errorf("unexpected panic parsing default gno.mod bytes: %w", err))
				}
				io.ErrPrintfln("auto-generated %q", fpath)
				err = mod.WriteFile(fpath)
				if err != nil {
					panic(fmt.Errorf("unexpected panic writing to %q: %w", fpath, err))
				}
				// err == nil.
			}
		}
		if err != nil {
			issue := lintIssue{
				Code:       lintGnoModError,
				Confidence: 1, // ??
				Location:   fpath,
				Msg:        err.Error(),
			}
			io.ErrPrintln(issue)
			hasError = true
			return commands.ExitCodeError(1)
		}

		// See adr/pr4264_lint_transpile.md
		// STEP 1: ReadMemPackage()
		// Read MemPackage with pkgPath.
		pkgPath, _ := determinePkgPath(mod, dir, cmd.rootDir)
		mpkg, err := gno.ReadMemPackage(dir, pkgPath)
		if err != nil {
			io.ErrPrintln(issueFromError(
				dir, pkgPath, err, lintReadError).String())
			hasError = true
			continue
		}

		// Perform imports using the parent store.
		if err := test.LoadImports(ts, mpkg); err != nil {
			io.ErrPrintln(issueFromError(
				dir, pkgPath, err, lintImportError).String())
			hasError = true
			continue
		}

		// Handle runtime errors
		didPanic := catchPanic(dir, pkgPath, io.Err(), func() {
			// Wrap in cache wrap so execution of the linter
			// doesn't impact other packages.
			cw := bs.CacheWrap()
			gs := ts.BeginTransaction(cw, cw, nil)

			// These are Go types.
			var pn *gno.PackageNode
			var errs error

			// Run type checking
			// STEP 2: ParseGnoMod()
			// STEP 3: GoParse*()
			//
			// lintTypeCheck(mpkg) -->
			//   TypeCheckMemPackage(mpkg) -->
			//     imp.typeCheckMemPackage(mpkg)
			//       ParseGnoMod(mpkg);
			//       GoParseMemPackage(mpkg);
			//       g.cmd.Check();
			if !mod.Draft {
				_, _, _, _, _, errs = lintTypeCheck(io, dir, mpkg, gs)
				if errs != nil {
					// io.ErrPrintln(errs) printed above.
					hasError = true
					return
				}
			} else if cmd.verbose {
				io.ErrPrintfln("%s: module is draft, skipping type check", dir)
			}

			// STEP 4: Prepare*()
			// Construct machine for testing.
			tm := test.Machine(gs, goio.Discard, pkgPath, false)
			defer tm.Release()

			// STEP 5: re-parse
			// Gno parse source fileset and test filesets.
			_, fset, _tests, ftests := sourceAndTestFileset(mpkg)

			// STEP 6: PreprocessFiles()
			// Preprocess fset files (w/ some _test.gno).
			pn, _ = tm.PreprocessFiles(
				mpkg.Name, mpkg.Path, fset, false, false)
			// STEP 6: PreprocessFiles()
			// Preprocess _test files (all _test.gno).
			for _, fset := range _tests {
				tm.PreprocessFiles(
					mpkg.Name, mpkg.Path, fset, false, false)
			}
			// STEP 6: PreprocessFiles()
			// Preprocess _filetest.gno files.
			for _, fset := range ftests {
				tm.PreprocessFiles(
					mpkg.Name, mpkg.Path, fset, false, false)
			}

			// Record results.
			ppkgs[dir] = processedPackage{
				mpkg, fset, pn, _tests, ftests,
			}
		})
		if didPanic {
			hasError = true
		}
	}
	if hasError {
		return commands.ExitCodeError(1)
	}

	//----------------------------------------
	// STAGE 2: Write.
	// Must be a separate stage to prevent partial writes.
	for _, dir := range dirs {
		ppkg, ok := ppkgs[dir]
		if !ok {
			panic("where did it go")
		}

		// STEP 10: mpkg.WriteTo():
		err := ppkg.mpkg.WriteTo(dir)
		if err != nil {
			return err
		}
	}

	return nil
}

// Wrapper around TypeCheckMemPackage() to io.ErrPrintln(lintIssue{}).
// Prints and returns errors. Panics upon an unexpected error.
func lintTypeCheck(
	// Args:
	io commands.IO,
	dir string,
	mpkg *std.MemPackage,
	testStore gno.Store) (
	// Results:
	gopkg *types.Package,
	gofset *token.FileSet,
	gofs, _gofs, tgofs []*ast.File,
	lerr error,
) {
	//----------------------------------------

	// gno.TypeCheckMemPackage(mpkg, testStore)
	var tcErrs error
	gopkg, gofset, gofs, _gofs, tgofs, tcErrs = gno.TypeCheckMemPackage(mpkg, testStore)

	// Print errors, and return the first unexpected error.
	errors := multierr.Errors(tcErrs)
	for _, err := range errors {
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
			panic(err) // unexpected.
		}
	}

	lerr = tcErrs
	return
}

// Gno parses and sorts mpkg files into the following filesets:
//   - fset: all normal and _test.go files in package excluding `package xxx_test`
//     integration *_test.gno files.
//   - _tests: `package xxx_test` integration *_test.gno files, each in their
//     own file set.
//   - ftests: *_filetest.gno file tests, each in their own file set.
func sourceAndTestFileset(mpkg *std.MemPackage) (
	all, fset *gno.FileSet, _tests, ftests []*gno.FileSet,
) {
	all = &gno.FileSet{}
	fset = &gno.FileSet{}
	for _, mfile := range mpkg.Files {
		if !strings.HasSuffix(mfile.Name, ".gno") {
			continue // Skip non-GNO files
		}

		n := gno.MustParseFile(mfile.Name, mfile.Body)
		if n == nil {
			continue // Skip empty files
		}
		all.AddFiles(n)
		if string(n.PkgName) == string(mpkg.Name)+"_test" {
			// A xxx_file integration test is a package of its own.
			fset := &gno.FileSet{}
			fset.AddFiles(n)
			_tests = append(_tests, fset)
		} else if strings.HasSuffix(mfile.Name, "_filetest.gno") {
			// A _filetest.gno is a package of its own.
			fset := &gno.FileSet{}
			fset.AddFiles(n)
			ftests = append(ftests, fset)
		} else {
			// All normal package files and,
			// _test.gno files that aren't xxx_test.
			fset.AddFiles(n)
		}
	}
	return
}

func guessSourcePath(pkgPath, fname string) string {
	if info, err := os.Stat(pkgPath); !os.IsNotExist(err) && !info.IsDir() {
		pkgPath = filepath.Dir(pkgPath)
	}

	fnameJoin := filepath.Join(pkgPath, fname)
	if _, err := os.Stat(fnameJoin); !os.IsNotExist(err) {
		return filepath.Clean(fnameJoin)
	}

	if _, err := os.Stat(fname); !os.IsNotExist(err) {
		return filepath.Clean(fname)
	}

	return filepath.Clean(pkgPath)
}

// reParseRecover is a regex designed to parse error details from a string.
// It extracts the file location, line number, and error message from a
// formatted error string.
// XXX: Ideally, error handling should encapsulate location details within a
// dedicated error type.
var reParseRecover = regexp.MustCompile(`^([^:]+)((?::(?:\d+)){1,2}):? *(.*)$`)

func catchPanic(dir, pkgPath string, stderr goio.WriteCloser, action func()) (didPanic bool) {
	defer func() {
		// Errors catched here mostly come from:
		// gnovm/pkg/gnolang/preprocess.go
		r := recover()
		if r == nil {
			return
		}
		didPanic = true
		switch verr := r.(type) {
		case *gno.PreprocessError:
			err := verr.Unwrap()
			fmt.Fprintln(stderr, issueFromError(
				dir, pkgPath, err, lintPreprocessError).String())
		case error:
			errors := multierr.Errors(verr)
			for _, err := range errors {
				errList, ok := err.(scanner.ErrorList)
				if ok {
					for _, errorInList := range errList {
						fmt.Fprintln(stderr, issueFromError(
							dir,
							pkgPath,
							errorInList,
							lintParserError,
						).String())
					}
				} else {
					fmt.Fprintln(stderr, issueFromError(
						dir,
						pkgPath,
						err,
						lintUnknownError,
					).String())
				}
			}
		case string:
			fmt.Fprintln(stderr, issueFromError(
				dir,
				pkgPath,
				errors.New(verr),
				lintUnknownError,
			).String())
		default:
			panic(r)
		}
	}()

	action()
	return
}

func issueFromError(dir, pkgPath string, err error, code lintCode) lintIssue {
	var issue lintIssue
	issue.Confidence = 1
	issue.Code = code

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
