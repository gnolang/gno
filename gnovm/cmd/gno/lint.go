package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	goio "io"
	"io/fs"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/cmd/gno/internal/cmdutil"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/packages"
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
	fs.BoolVar(&c.autoGnomod, "auto-gnomod", true, "auto-generate gnomod.toml file if not already present")
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

	loadCfg := packages.LoadConfig{
		Fetcher:    testPackageFetcher,
		Deps:       true,
		Test:       true,
		Out:        io.Err(),
		AllowEmpty: true,
		GnoRoot:    cmd.rootDir,
	}
	pkgs, err := packages.Load(loadCfg, args...)
	if err != nil {
		return err
	}

	hasError := false

	prodbs, prodgs := test.StoreWithOptions(
		cmd.rootDir, goio.Discard,
		test.StoreOptions{PreprocessOnly: true, WithExtern: false, WithExamples: true, Testing: false, Packages: pkgs},
	)
	testbs, testgs := test.StoreWithOptions(
		cmd.rootDir, goio.Discard,
		test.StoreOptions{
			PreprocessOnly: true,
			WithExtern:     false,
			WithExamples:   true,
			Testing:        true,
			SourceStore:    prodgs,
			Packages:       pkgs,
		},
	)
	ppkgs := map[string]cmdutil.ProcessedPackage{}
	cache := make(gno.TypeCheckCache)

	if cmd.verbose {
		targetsNames := []string{}
		for _, pkg := range pkgs {
			if len(pkg.Match) == 0 {
				continue
			}
			targetsNames = append(targetsNames, lintTargetName(pkg))
		}
		io.ErrPrintfln("linting packages: %v", targetsNames)
	}
	//----------------------------------------
	// LINT STAGE 1: Typecheck and lint.
	for _, pkg := range pkgs {
		// ignore dependencies
		if len(pkg.Match) == 0 {
			continue
		}

		if cmd.verbose {
			io.ErrPrintfln("linting %q", lintTargetName(pkg))
		}

		// XXX Currently the linter only supports linting directories.
		// In order to support linting individual files, we need to
		// refactor this code to work with mempackages, not dirs, and
		// cmd/gno/util.go needs to be refactored to return mempackages
		// rather than dirs. Commands like `gno lint a.gno b.gno`
		// should create a temporary package from just those files. We
		// could also load mempackages lazily for memory efficiency.
		// Alternative: support `command-line-arguments` in packages.Load
		dir := pkg.Dir

		// Read and parse gnomod.toml directly.
		fpath := filepath.Join(dir, "gnomod.toml")
		mod, err := gnomod.ParseFilepath(fpath)
		if errors.Is(err, fs.ErrNotExist) {
			// TODO: gno.mod is deprecated, but we still support it for now.
			// if gno.mod exists -> port
			if cmd.autoGnomod {
				modulePath, _ := determinePkgPath(nil, dir, cmd.rootDir)
				modstr := gno.GenGnoModLatest(modulePath)
				mod, err = gnomod.ParseBytes("gnomod.toml", []byte(modstr))
				if err != nil {
					panic(fmt.Errorf("unexpected panic parsing default gnomod.toml bytes: %w", err))
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
		mpkg, err := gno.ReadMemPackage(dir, pkgPath, gno.MPAnyAll)
		if err != nil {
			printError(io.Err(), dir, pkgPath, err)
			hasError = true
			continue
		}

		// Skip processing for ignored modules
		if mod.Ignore {
			if cmd.verbose {
				io.ErrPrintfln("%s: module is ignored, skipping", dir)
			}
			continue
		}

		// Perform imports using the parent store.
		abortOnError := true
		if err := test.LoadImports(testgs, mpkg, abortOnError); err != nil {
			printError(io.Err(), dir, pkgPath, err)
			hasError = true
			continue
		}

		// Wrap in cache wrap so execution of the linter
		// doesn't impact other packages.
		newProdGnoStore := func() gno.Store {
			pcw := prodbs.CacheWrap()
			pgs := prodgs.BeginTransaction(pcw, pcw, nil)
			return pgs
		}
		injectTmpkg := func(tgs gno.Store) {
			// NOTE: if we don't do it lazily like this, otherwise there
			// needs to be a hook from original store creation
			// (complicated), or, if not done lazily we won't get the Go
			// typecheck error we prefer.
			tgetter := tgs.GetPackageGetter()
			tgs.SetPackageGetter(func(pkgPath string, store gno.Store) (
				*gno.PackageNode, *gno.PackageValue,
			) {
				if pkgPath == mpkg.Path {
					tmpkg := gno.MPFTest.FilterMemPackage(mpkg)
					m2 := gno.NewMachineWithOptions(gno.MachineOptions{
						PkgPath:     pkgPath,
						Output:      goio.Discard,
						Store:       tgs,
						SkipPackage: true,
					})
					// Use the actual type of the filtered package
					tmpkgType := tmpkg.Type.(gno.MemPackageType)
					m2.Store.AddMemPackage(tmpkg, tmpkgType)
					return m2.PreprocessFiles(tmpkg.Name, tmpkg.Path,
						m2.ParseMemPackageAsType(tmpkg, tmpkgType), true, true, "")
				} else {
					return tgetter(pkgPath, store)
				}
			})
		}
		newTestGnoStore := func(withTmpkg bool) gno.Store {
			tcw := testbs.CacheWrap()
			tgs := testgs.BeginTransaction(tcw, tcw, nil)
			if withTmpkg {
				injectTmpkg(tgs)
			}
			return tgs
		}

		// Handle runtime errors
		didPanic := catchPanic(dir, pkgPath, io.Err(), func() {
			// Memo process results here.
			ppkg := cmdutil.ProcessedPackage{MPkg: mpkg, Dir: dir}

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

			tcmode := gno.TCLatestStrict
			if cmd.autoGnomod {
				tcmode = gno.TCLatestRelaxed
			}
			errs := lintTypeCheck(io, dir, mpkg, gno.TypeCheckOptions{
				Getter:     newProdGnoStore(),
				TestGetter: newTestGnoStore(true),
				Mode:       tcmode,
				Cache:      cache,
			})
			if errs != nil {
				// io.ErrPrintln(errs) printed above.
				hasError = true
				return
			}

			// Construct machine for testing.
			tm := test.Machine(newProdGnoStore(), goio.Discard, pkgPath, false, nil)
			defer tm.Release()

			// LINT STEP 4: re-parse for preprocessor.
			// While lintTypeCheck > TypeCheckMemPackage will find
			// most issues, the preprocessor may have additional
			// checks.
			// Gno parse source fileset and test filesets.
			_, fset, tfset, _tests, ftests := sourceAndTestFileset(mpkg, false)

			{
				// LINT STEP 5: PreprocessFiles()
				// Preprocess fset files (no test files)
				tm.Store = newProdGnoStore()
				pn, _ := tm.PreprocessFiles(
					mpkg.Name, mpkg.Path, fset, false, false, "")
				ppkg.AddNormal(pn, fset)
			}
			{
				// LINT STEP 5: PreprocessFiles()
				// Preprocess fset files (w/ some *_test.gno).
				tm.Store = newTestGnoStore(false)
				pn, _ := tm.PreprocessFiles(
					mpkg.Name, mpkg.Path, tfset, false, false, "")
				ppkg.AddTest(pn, fset)
			}
			{
				// LINT STEP 5: PreprocessFiles()
				// Preprocess _test files (all xxx_test *_test.gno).
				tm.Store = newTestGnoStore(true)
				pn, _ := tm.PreprocessFiles(
					mpkg.Name+"_test", mpkg.Path+"_test", _tests, false, false, "")
				ppkg.AddUnderscoreTests(pn, _tests)
			}
			{
				// LINT STEP 5: PreprocessFiles()
				// Preprocess _filetest.gno files.
				for i, fset := range ftests {
					tm.Store = newTestGnoStore(true)
					fname := fset.Files[0].FileName
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
	for _, pkg := range pkgs {
		// ignore dependencies
		if len(pkg.Match) == 0 {
			continue
		}

		ppkg, ok := ppkgs[pkg.Dir]
		if !ok {
			// Skip directories that were not processed (e.g., ignored modules)
			continue
		}

		// LINT STEP 6: mpkg.WriteTo():
		err := ppkg.MPkg.WriteTo(pkg.Dir)
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
	opts gno.TypeCheckOptions) (
	// Results:
	lerr error,
) {
	// gno.TypeCheckMemPackage(mpkg, testStore).
	_, tcErrs := gno.TypeCheckMemPackage(mpkg, opts)

	// Print errors, and return the first unexpected error.
	errors := multierr.Errors(tcErrs)
	for _, err := range errors {
		printError(io.Err(), dir, mpkg.Path, err)
	}

	lerr = tcErrs
	return
}

func lintTargetName(pkg *packages.Package) string {
	if pkg.ImportPath != "" {
		return pkg.ImportPath
	}

	return tryRelativizePath(pkg.Dir)
}
