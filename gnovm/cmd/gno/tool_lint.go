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
	"os"
	"path"
	"path/filepath"
	"regexp"
	rdebug "runtime/debug"
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
	mpkg   *std.MemPackage
	fset   *gno.FileSet
	pn     *gno.PackageNode
	_tests []*gno.FileSet
	ftests []*gno.FileSet
}

type lintCfg struct {
	verbose bool
	rootDir string
	// min_confidence: minimum confidence of a problem to print it
	// (default 0.8) auto-fix: apply suggested fixes automatically.
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

type lintCode string

const (
	lintUnknown        lintCode = "lintUnknown"
	lintGnoMod                  = "lintGnoMod"
	lintGnoError                = "lintGnoError"
	lintParserError             = "lintParserError"
	lintTypeCheckError          = "lintTypeCheckError"

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
	// STAGE 1:
	for _, dir := range dirs {
		if verbose {
			io.ErrPrintln(dir)
		}

		info, err := os.Stat(dir)
		if err == nil && !info.IsDir() {
			dir = filepath.Dir(dir)
		}

		// TODO: This should be handled by ReadMemPackage.
		fpath := path.Join(dir, "gno.mod")
		mod, err := gnomod.ParseFilepath(fpath)
		if err != nil {
			issue := lintIssue{
				Code:       lintGnoMod,
				Confidence: 1, // ??
				Location:   fpath,
				Msg:        err.Error(),
			}
			io.ErrPrintln(issue)
			hasError = true
			continue
		}

		// STEP 1: ReadMemPackage()
		// Read MemPackage with pkgPath.
		pkgPath, _ := determinePkgPath(mod, dir, cfg.rootDir)
		mpkg, err := gno.ReadMemPackage(dir, pkgPath)
		if err != nil {
			io.ErrPrintln(issueFromError(
				dir, pkgPath, err, "ReadMemPackge").String())
			hasError = true
			continue
		}

		// Perform imports using the parent store.
		if err := test.LoadImports(ts, mpkg); err != nil {
			io.ErrPrintln(issueFromError(
				dir, pkgPath, err, "LoadImports").String())
			hasError = true
			continue
		}

		// Handle runtime errors
		hasRuntimeErr := catchRuntimeError(dir, pkgPath, io.Err(), func() {
			// Wrap in cache wrap so execution of the linter
			// doesn't impact other packages.
			cw := bs.CacheWrap()
			gs := ts.BeginTransaction(cw, cw, nil)

			// These are Go types.
			var pkg *token.Package
			var fset *token.FileSet
			var astfs []*ast.File
			var errs error

			// Run type checking
			if mod == nil || !mod.Draft {
				// STEP 2: ParseGnoMod()
				// STEP 3: GoParse*()
				//
				// typeCheckAndPrintErrors(mpkg) -->
				//   TypeCheckMemPackage(mpkg) -->
				//     imp.typeCheckMemPackage(mpkg)
				//       ParseGnoMod(mpkg);
				//       GoParseMemPackage(mpkg);
				//       g.cfg.Check();
				//
				// NOTE: Prepare*() is not called.
				pkg, fset, astfs, errs = typeCheckAndPrintErrors(io, dir, mpkg, gs)
				if errs != nil {
					io.ErrPrintln(errs)
					hasError = true
				} else if foundErr {
					hasError = true
				}
			} else if verbose {
				io.ErrPrintfln("%s: module is draft, skipping type check", dir)
			}

			// STEP 4 & 5: Prepare*() and Preprocess*().
			{
				// Construct machine for testing.
				tm := test.Machine(gs, goio.Discard, pkgPath, false)
				defer tm.Release()

				// Gno parse source fileset and test filesets.
				all, fset2, _tests, ftests := sourceAndTestFileset(mpkg)

				// Prepare Go AST for preprocessing.
				errs := PrepareGno0p9(fset, astfs)
				if errs != nil {
					io.ErrPrintln(errs)
					hasError = true
				}

				// Preprocess fset files (w/ some _test.gno).
				pn, _ := tm.PreprocessFiles(
					mpkg.Name, mpkg.Path, fset2, false, false)
				// Preprocess _test files (all _test.gno).
				for _, fset3 := range _tests {
					tm.PreprocessFiles(
						mpkg.Name, mpkg.Path, fset3, false, false)
				}
				// Preprocess _filetest.gno files.
				for _, fset3 := range ftests {
					tm.PreprocessFiles(
						mpkg.Name, mpkg.Path, fset3, false, false)
				}
			}

			// Record results.
			ppkgs[dir] = processedPackage{
				mpkg, fset, pn, _tests, ftests}

			// STEP 6: FindXItems():
			// FindXItems for all files if outdated.
			if mod.Gno.Version == "0.0" {
				// Use the preprocessor to collect the
				// transformations needed to be done.
				// They are collected in
				// pn.GetAttribute("XREALMITEM")
				for _, fn := range all.Files {
					gno.FindXItemsGno0p9(gs, pn, fn)
				}
			}
		})
		if hasRuntimeErr {
			hasError = true
		}
	}

	//----------------------------------------
	// STAGE 2: Transpile to Gno 0.9
	// Must be a separate stage because dirs depend on each other.
	for _, dir := range dirs {
		ppkg, ok := ppkgs[dir]
		if !ok {
			panic("where did it go")
		}
		mpkg, pn := ppkg.mpkg, ppkg.pn
		xform, _ := pn.GetAttribute(gno.ATTR_GNO0P9_XITEMS).(map[string]string)

		// STEP 7 & 8: gno.TranspileGno0p9() Part 1 & 2
		err := gno.TranspileGno0p9(mpkg, dir, xform)
		if err != nil {
			return err
		}
	}
	if hasError {
		return commands.ExitCodeError(1)
	}

	//----------------------------------------
	// STAGE 3: Write.
	// Must be a separate stage to prevent partial writes.
	for _, dir := range dirs {
		ppkg, ok := ppkgs[dir]
		if !ok {
			panic("where did it go")
		}

		// STEP 9: mpkg.WriteTo():
		err := ppkg.mpkg.WriteTo(dir)
		if err != nil {
			return err
		}
	}

	return nil
}

// Wrapper around TypeCheckMemPackage() to io.ErrPrintln(lintIssue{}).
func typeCheckAndPrintErrors(
	io commands.IO,
	dir string,
	mpkg *std.MemPackage,
	testStore gno.Store,
) (
	pkg *token.Package,
	fset *token.FileSet,
	astfs []*ast.File,
	errs error) {

	pkg, fset, astfs, errs = gno.TypeCheckMemPackage(mpkg, testStore)
	if errs != nil {
		return
	}

	errors := multierr.Errors(errs)
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
			errs = fmt.Errorf("unexpected error type; %T", err)
			return
		}
	}
	return
}

// Gno parses and sorts mpkg files into the following filesets:
//   - all: all files
//   - fset: all normal and _test.go files in package excluding `package xxx_test`
//     integration *_test.gno files.
//   - _tests: `package xxx_test` integration *_test.gno files, each in their
//     own file set.
//   - ftests: *_filetest.gno file tests, each in their own file set.
func sourceAndTestFileset(mpkg *std.MemPackage) (
	all, fset *gno.FileSet, _tests, ftests []*gno.FileSet) {

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
// It extracts the file location, line number, and error message from a
// formatted error string.
// XXX: Ideally, error handling should encapsulate location details within a
// dedicated error type.
var reParseRecover = regexp.MustCompile(`^([^:]+)((?::(?:\d+)){1,2}):? *(.*)$`)

func catchRuntimeError(dir, pkgPath string, stderr goio.WriteCloser, action func()) (hasError bool) {
	defer func() {
		// Errors catched here mostly come from:
		// gnovm/pkg/gnolang/preprocess.go
		r := recover()
		if r == nil {
			return
		}
		rdebug.PrintStack()
		hasError = true
		switch verr := r.(type) {
		case *gno.PreprocessError:
			err := verr.Unwrap()
			fmt.Fprintln(stderr, issueFromError(
				dir, pkgPath, err, "panic=PreprocessError").String())
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
							"panic=+go/scanner.ErrorList",
						).String())
					}
				} else {
					fmt.Fprintln(stderr, issueFromError(
						dir,
						pkgPath,
						err,
						"panic=error",
					).String())
				}
			}
		case string:
			fmt.Fprintln(stderr, issueFromError(
				dir,
				pkgPath,
				errors.New(verr),
				"panic=string",
			).String())
		default:
			panic(r)
		}
	}()

	action()
	return
}

func issueFromError(dir, pkgPath string, err error, why string) lintIssue {
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
		issue.Msg = strings.TrimSpace(matches[3]) + " (" + why + ")"
	} else {
		issue.Location = fmt.Sprintf("%s:0", filepath.Clean(pkgPath))
		issue.Msg = err.Error() + " (" + why + ")"
	}
	return issue
}

func replaceWithDirPath(s, pkgPath, dir string) string {
	if strings.HasPrefix(s, pkgPath) {
		return filepath.Clean(dir + s[len(pkgPath):])
	}
	return s
}
