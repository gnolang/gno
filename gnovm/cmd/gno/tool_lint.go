package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"go/types"
	goio "io"
	"io/fs"
	"os"
	"path/filepath"

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
	// LINT STAGE 1: Typecheck and lint.
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
		fpath := filepath.Join(dir, "gno.mod")
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
			issue := gnoIssue{
				Code:       gnoGnoModError,
				Confidence: 1, // ??
				Location:   fpath,
				Msg:        err.Error(),
			}
			io.ErrPrintln(issue)
			hasError = true
			return commands.ExitCodeError(1)
		}

		// See adr/pr4264_lint_transpile.md
		// LINT STEP 1: ReadMemPackage()
		// Read MemPackage with pkgPath.
		pkgPath, _ := determinePkgPath(mod, dir, cmd.rootDir)
		mpkg, err := gno.ReadMemPackage(dir, pkgPath)
		if err != nil {
			printError(io.Err(), dir, pkgPath, err)
			hasError = true
			continue
		}

		// Perform imports using the parent store.
		if err := test.LoadImports(ts, mpkg); err != nil {
			printError(io.Err(), dir, pkgPath, err)
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
			ppkg := processedPackage{mpkg: mpkg, dir: dir}
			var errs error

			// Run type checking
			// LINT STEP 2: ParseGnoMod()
			// STEP 3: GoParse*()
			//
			// lintTypeCheck(mpkg) -->
			//   TypeCheckMemPackage(mpkg) -->
			//     imp.typeCheckMemPackage(mpkg)
			//       ParseGnoMod(mpkg);
			//       GoParseMemPackage(mpkg);
			//       g.cmd.Check();
			if !mod.Draft {
				_, _, errs = lintTypeCheck(io, dir, mpkg, gs)
				if errs != nil {
					// io.ErrPrintln(errs) printed above.
					hasError = true
					return
				}
			} else if cmd.verbose {
				io.ErrPrintfln("%s: module is draft, skipping type check", dir)
			}

			// Construct machine for testing.
			tm := test.Machine(gs, goio.Discard, pkgPath, false)
			defer tm.Release()

			// LINT STEP 4: re-parse
			// Gno parse source fileset and test filesets.
			_, fset, _tests, ftests := sourceAndTestFileset(mpkg)

			{
				// LINT STEP 5: PreprocessFiles()
				// Preprocess fset files (w/ some _test.gno).
				pn, _ := tm.PreprocessFiles(
					mpkg.Name, mpkg.Path, fset, false, false, "")
				ppkg.AddNormal(pn, fset)
			}
			{
				// LINT STEP 5: PreprocessFiles()
				// Preprocess _test files (all _test.gno).
				cw := bs.CacheWrap()
				gs := ts.BeginTransaction(cw, cw, nil)
				tm.Store = gs
				pn, _ := tm.PreprocessFiles(
					mpkg.Name+"_test", mpkg.Path+"_test", _tests, false, false, "")
				ppkg.AddUnderscoreTests(pn, _tests)
			}
			{
				// LINT STEP 5: PreprocessFiles()
				// Preprocess _filetest.gno files.
				for i, fset := range ftests {
					cw := bs.CacheWrap()
					gs := ts.BeginTransaction(cw, cw, nil)
					tm.Store = gs
					fname := string(fset.Files[0].Name)
					mfile := mpkg.GetFile(fname)
					pkgPath := fmt.Sprintf("%s_filetest%d", mpkg.Path, i)
					pkgPath, err = parsePkgPathDirective(mfile.Body, pkgPath)
					if err != nil {
						io.ErrPrintln(err)
						hasError = true
						continue
					}
					pkgName := string(fset.Files[0].PkgName)
					pn, _ := tm.PreprocessFiles(pkgName, pkgPath, fset, false, false, "")
					ppkg.AddFileTest(pn, fset)
				}
			}

			// Record results.
			ppkgs[dir] = ppkg
		})
		if didPanic {
			hasError = true
		}
	}
	if hasError {
		return commands.ExitCodeError(1)
	}

	//----------------------------------------
	// LINT STAGE 2: Write.
	// Must be a separate stage to prevent partial writes.
	for _, dir := range dirs {
		ppkg, ok := ppkgs[dir]
		if !ok {
			panic("where did it go")
		}

		// LINT STEP 6: mpkg.WriteTo():
		err := ppkg.mpkg.WriteTo(dir)
		if err != nil {
			return err
		}
	}

	return nil
}

// Wrapper around TypeCheckMemPackage() to io.ErrPrintln(gnoIssue{}).
// Prints and returns errors. Panics upon an unexpected error.
func lintTypeCheck(
	// Args:
	io commands.IO,
	dir string,
	mpkg *std.MemPackage,
	testStore gno.Store) (
	// Results:
	gopkg *types.Package,
	tfiles *gno.TypeCheckFilesResult,
	lerr error,
) {
	//----------------------------------------

	// gno.TypeCheckMemPackage(mpkg, testStore)
	var tcErrs error
	gopkg, tfiles, tcErrs = gno.TypeCheckMemPackage(mpkg, testStore, gno.ParseModeAll)

	// Print errors, and return the first unexpected error.
	errors := multierr.Errors(tcErrs)
	for _, err := range errors {
		printError(io.Err(), dir, mpkg.Path, err)
	}

	lerr = tcErrs
	return
}
