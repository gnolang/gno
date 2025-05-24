package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	goio "io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

/*
Translate Interrealm Spec 2 to Interrealm Spec 3 (Gno 0.9)

 - Interrealm Spec 1: Original; every realm function is (automatically)
 a crossing function. This was working for our examples and was
 conceptually simple, but several problems were identified late in
 development;

   1. p package code copied over to r realms would behave differently
   with respect to std.CurrentRealm() and std.PreviousRealm(). It will
   become typical after launch that p code gets copied to r code for
   cutstom patchces; and potential p code will first to be tested in
   more mutable r realms.

   2. a reentrancy issue exists where r realm's calls to some variable
   function/method `var A func(...)...` are usually of functions
   declared in external realms (such as callback functions expected to
   be provided by the external realm) but instead ends up being a
   function declared in the the same r realm, an expected realm
   boundary isn't there, and may lead to exploits.

 - Interrealm Spec 2: With explicit cross(fn)(...) and crossing()
 declarations. The previous problems were solved by explicit crossing()
 declarations in realm functions (solves 1), and explicit
 cross(fn)(...) calls (solves 2 for the most part). But more problems
 were identified after most of the migration was done for examples from
 spec 1 to spec 2:

   3. a reentrancy issue where if calls to r realm's function/method
   A() are usually expected to be done by external realms (creating a
   realm boundary), but the external caller does things to get the r
   realm to call its own A(), the expected realm boundary isn't created
   and may lead to exploits.

   3.A. As a more concrete example of problem 3, when a realm takes as
   parameter a callback function `cb func(...)...` that isn't expected
   to be a crossing function and thus not explicitly crossed into. An
   external user or realm can then craft a function literal expression
   that calls the aforementioned realm's crossing functions without an
   explicit cross(fn)(...) call, thereby again dissolving a realm
   function boundary where one should be.

   4. Users didn't like the cross(fn)(...) syntax.

 - Interrealm Spec 3: With @cross decorator and `cur realm` first
 argument type. Instead of declaring a crossing-function with
 `crossing()` as the first statement the @cross decorator is used for
 package/file level function/methods declarations. Function literals
 can likewise be declared crossing by being wrapped like
 cross(func(...)...{}). When calling from within the same realm
 (without creating a realm boundary), the `cur` value is passed through
 to the called function's via its first argument; but when a realm
 boundary is intended, `nil` is passed in instead. This resolves
 problem 3.A because a non-crossing function literal would not be
 declared with the `cur realm` first argument, and thus a non-crossing
 call of the same realm's crossing function would not be syntactically
 possible.

----------------------------------------

Also refer to the [Lint and Transpile ADR](./adr/pr4264_lint_transpile.md).
*/

type fixCmd struct {
	verbose       bool
	rootDir       string
	filetestsOnly bool
	// min_confidence: minimum confidence of a problem to print it
	// (default 0.8) auto-fix: apply suggested fixes automatically.
}

func newFixCmd(cio commands.IO) *commands.Command {
	cmd := &fixCmd{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "fix",
			ShortUsage: "fix [flags] <package> [<package>...]",
			ShortHelp:  "runs the fixer for the specified packages",
		},
		cmd,
		func(_ context.Context, args []string) error {
			return execFix(cmd, args, cio)
		},
	)
}

func (c *fixCmd) RegisterFlags(fs *flag.FlagSet) {
	rootdir := gnoenv.RootDir()

	fs.BoolVar(&c.verbose, "v", false, "verbose output when fixning")
	fs.StringVar(&c.rootDir, "root-dir", rootdir, "clone location of github.com/gnolang/gno (gno tries to guess it)")
	fs.BoolVar(&c.filetestsOnly, "only-filetests", false, "dir only contains filetests. not recursive.")
}

func execFix(cmd *fixCmd, args []string, cio commands.IO) error {
	// Show a help message by default.
	if len(args) == 0 {
		return flag.ErrHelp
	}

	// Guess cmd.RootDir.
	if cmd.rootDir == "" {
		cmd.rootDir = gnoenv.RootDir()
	}

	var dirs []string
	var hasError bool
	var err error

	if cmd.filetestsOnly {
		dirs = slices.Clone(args)
	} else {
		dirs, err = gnoPackagesFromArgsRecursively(args)
		if err != nil {
			return fmt.Errorf("list packages from args: %w", err)
		}
	}

	bs, ts := test.StoreWithOptions(
		cmd.rootDir, goio.Discard,
		test.StoreOptions{PreprocessOnly: true, FixFrom: gno.GnoVerMissing},
	)
	ppkgs := map[string]processedPackage{}

	if cmd.verbose {
		cio.ErrPrintfln("flinting directories: %v", dirs)
	}
	//----------------------------------------
	// FIX STAGE 1: Type-check and lint.
	for _, dir := range dirs {
		if cmd.verbose {
			cio.ErrPrintfln("fixing %q", dir)
		}

		// Only supports directories.
		// You should fix all directories at once to avoid dependency issues.
		info, err := os.Stat(dir)
		if err == nil && !info.IsDir() {
			dir = filepath.Dir(dir)
		}

		// Read and parse gno.mod directly.
		fpath := path.Join(dir, "gno.mod")
		mod, err := gnomod.ParseFilepath(fpath)
		if errors.Is(err, fs.ErrNotExist) {
			// Make a temporary gno.mod (but don't write it yet)
			modstr := gno.GenGnoModMissing("gno.land/r/xxx_myrealm_xxx/xxx_fixme_xxx")
			mod, err = gnomod.ParseBytes("gno.mod", []byte(modstr))
			if err != nil {
				panic(fmt.Errorf("unexpected panic parsing default gno.mod bytes: %w", err))
			}
		} else {
			switch mod.GetGno() {
			case gno.GnoVerLatest:
				if cmd.verbose {
					cio.ErrPrintfln("%s: module is up to date, skipping fix", dir)
				}
				continue // nothing to do.
			case gno.GnoVerMissing:
				// good, fix it.
			default:
				cio.ErrPrintfln("%s: unrecognized gno.mod version %q, skipping fix", dir, mod.GetGno())
				continue // skip it.
			}
		}
		if err != nil {
			issue := gnoIssue{
				Code:       gnoGnoModError,
				Confidence: 1, // ??
				Location:   fpath,
				Msg:        err.Error(),
			}
			cio.ErrPrintln(issue)
			hasError = true
			return commands.ExitCodeError(1)
		}
		if mod.Draft {
			cio.ErrPrintfln("%s: module is draft, skipping fix", dir)
			continue
		}

		// See adr/pr4264_fix_transpile.md
		// FIX STEP 1: ReadMemPackage()
		// Read MemPackage with pkgPath.
		pkgPath, _ := determinePkgPath(mod, dir, cmd.rootDir)
		mpkg, err := gno.ReadMemPackage(dir, pkgPath)
		if err != nil {
			printError(cio.Err(), dir, pkgPath, err)
			hasError = true
			continue
		}

		// Filter out filetests that fail type-check.
		if cmd.filetestsOnly {
			filterInvalidFiletests(cio, mpkg)
		}

		// Perform imports using the parent store.
		if err := test.LoadImports(ts, mpkg); err != nil {
			printError(cio.Err(), dir, pkgPath, err)
			hasError = true
			continue
		}

		// Handle runtime errors
		didPanic := catchPanic(dir, pkgPath, cio.Err(), func() {
			// Wrap in cache wrap so execution of the fixer
			// doesn't impact other packages.
			cw := bs.CacheWrap()
			gs := ts.BeginTransaction(cw, cw, nil)

			// These are Go types.
			ppkg := processedPackage{mpkg: mpkg, dir: dir}
			// Run type checking
			// FIX STEP 2: ParseGnoMod()
			// FIX STEP 3: GoParse*()
			//
			// lintTypeCheck(mpkg) -->
			//   TypeCheckMemPackage(mpkg) -->
			//     imp.typeCheckMemPackage(mpkg)
			//       ParseGnoMod(mpkg);
			//       GoParseMemPackage(mpkg);
			//       g.cmd.Check();

			_, tfiles, errs := lintTypeCheck(cio, dir, mpkg, gs)
			if errs != nil {
				// cio.ErrPrintln(errs) already printed.
				hasError = true
				return
			}

			// FIX STEP 4: Prepare*()
			// Construct machine for preprocessing.
			tm := test.Machine(gs, goio.Discard, pkgPath, false)
			defer tm.Release()

			// Prepare Go AST for preprocessing.
			allgofs := append(tfiles.SourceFiles, tfiles.TestPackageFiles...)
			allgofs = append(allgofs, tfiles.TestFiles...)
			errs = gno.PrepareGno0p9(tfiles.FileSet, allgofs, mpkg)
			if errs != nil {
				cio.ErrPrintln(errs)
				hasError = true
				return // Prepare must succeed.
			}

			// FIX STEP 5: re-parse
			// Gno parse source fileset and test filesets.
			_, fset, _tests, ftests := sourceAndTestFileset(mpkg) // , cmd.filetestsOnly)

			{
				// FIX STEP 6: PreprocessFiles()
				// Preprocess fset files (w/ some _test.gno).
				cw := bs.CacheWrap()
				gs := ts.BeginTransaction(cw, cw, nil)
				tm.Store = gs
				pn, _ := tm.PreprocessFiles(
					mpkg.Name, mpkg.Path, fset, false, false, gno.GnoVerMissing)
				ppkg.AddNormal(pn, fset)

				// FIX STEP 7: FindXforms():
				// FindXforms for all files if outdated.
				// Use the preprocessor to collect the
				// transformations needed to be done.
				// They are collected in
				// pn.GetAttribute("XREALMFORM")
				for _, fn := range fset.Files {
					gno.FindXformsGno0p9(gs, pn, fn)
				}
			}
			{
				// FIX STEP 6: PreprocessFiles()
				// Preprocess xxx_test files (some _test.gno).
				cw := bs.CacheWrap()
				gs := ts.BeginTransaction(cw, cw, nil)
				tm.Store = gs
				pn, _ := tm.PreprocessFiles(
					mpkg.Name+"_test", mpkg.Path+"_test", _tests, false, false, gno.GnoVerMissing)
				ppkg.AddUnderscoreTests(pn, _tests)

				// FIX STEP 7: FindXforms():
				// FindXforms for all files if outdated.
				// Use the preprocessor to collect the
				// transformations needed to be done.
				// They are collected in
				// pn.GetAttribute("XREALMFORM")
				for _, fn := range _tests.Files {
					gno.FindXformsGno0p9(gs, pn, fn)
				}
			}
			{
				// FIX STEP 6: PreprocessFiles()
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
						cio.ErrPrintln(err)
						hasError = true
						continue
					}
					pkgName := string(fset.Files[0].PkgName)
					pn, _ := tm.PreprocessFiles(pkgName, pkgPath, fset, false, false, gno.GnoVerMissing)
					ppkg.AddFileTest(pn, fset)

					// FIX STEP 7: FindXforms():
					// FindXforms for all files if outdated.
					// Use the preprocessor to collect the
					// transformations needed to be done.
					// They are collected in
					// pn.GetAttribute("XREALMFORM")
					for _, fn := range fset.Files {
						gno.FindXformsGno0p9(gs, pn, fn)
					}
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
	// FIX STAGE 2: Transpile to Gno 0.9
	// Must be a separate stage because dirs depend on each other.
	for _, dir := range dirs {
		ppkg, ok := ppkgs[dir]
		if !ok {
			// Happens when fixing a file, (XXX fix this case)
			// but also happens when preprocessing isn't needed.
			continue
		}

		// Sanity check.
		mod, err := gno.ParseCheckGnoMod(ppkg.mpkg)
		if mod.GetGno() != gno.GnoVerMissing {
			panic("should not happen")
		}

		// FIX STEP 8 & 9: gno.TranspileGno0p9() Part 1 & 2
		mpkg := ppkg.mpkg
		transpileProcessedFileSet := func(pfs processedFileSet) error {
			pn, fset := pfs.pn, pfs.fset
			xforms1, _ := pn.GetAttribute(gno.ATTR_GNO0P9_XFORMS).(map[string]struct{})
			err = gno.TranspileGno0p9(mpkg, dir, pn, fset.GetFileNames(), xforms1)
			return err
		}
		err = transpileProcessedFileSet(ppkg.normal)
		if err != nil {
			return err
		}
		err = transpileProcessedFileSet(ppkg._tests)
		if err != nil {
			return err
		}
		for _, ftest := range ppkg.ftests {
			err = transpileProcessedFileSet(ftest)
			if err != nil {
				return err
			}
		}
	}
	if hasError {
		return commands.ExitCodeError(1)
	}

	//----------------------------------------
	// FIX STAGE 3: Write.
	// Must be a separate stage to prevent partial writes.
	for _, dir := range dirs {
		ppkg, ok := ppkgs[dir]
		if !ok {
			// Happens when fixing a file, (XXX fix this case)
			// but also happens when preprocessing isn't needed.
			continue
		}

		// Write version to gno.mod.
		mod, err := gno.ParseCheckGnoMod(ppkg.mpkg)
		if err != nil {
			// should have been auto-generated.
			panic("missing gno.mod")
		}
		mod.SetGno(gno.GnoVerLatest)
		ppkg.mpkg.SetFile("gno.mod", mod.WriteString())

		// FIX STEP 10: mpkg.WriteTo():
		err = ppkg.mpkg.WriteTo(dir)
		if err != nil {
			return err
		}
	}

	return nil
}

// When in filetestsOnly mode, filter out files that are known to have a
// type-check error. These files will be deleted from the mpkg.
// They are only deleted from mpkg; gno fix will not affect the files
// already on disk.
func filterInvalidFiletests(cio commands.IO, mpkg *std.MemPackage) {
	for _, mfile := range mpkg.Files {
		dirs, err := test.ParseDirectives(bytes.NewReader([]byte(mfile.Body)))
		if err != nil {
			panic(fmt.Errorf("error parsing directives: %w", err))
		}
		tcErr := dirs.FirstDefault(test.DirectiveTypeCheckError, "")
		if tcErr != "" {
			cio.Printfln("skipping invalid filetest %q", mfile.Name)
			mpkg.DeleteFile(mfile.Name)
			continue
		}
	}
}
