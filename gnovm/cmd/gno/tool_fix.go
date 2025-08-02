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
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"go.uber.org/multierr"
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
	verbose        bool
	rootDir        string
	filetestsOnly  bool
	filetestsMatch string
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

	fs.BoolVar(&c.verbose, "v", false, "verbose output when fixing")
	fs.StringVar(&c.rootDir, "root-dir", rootdir, "clone location of github.com/gnolang/gno (gno tries to guess it)")
	fs.BoolVar(&c.filetestsOnly, "filetests-only", false, "dir only contains filetests. not recursive.")
	fs.StringVar(&c.filetestsMatch, "filetests-match", "", "if --filetests-only=true, filters by substring match.")
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

	var dirs []string = nil
	var err error

	if cmd.filetestsOnly {
		dirs = append([]string(nil), args...)
	} else {
		dirs, err = gnoPackagesFromArgsRecursively(args)
		if err != nil {
			return fmt.Errorf("list packages from args: %w", err)
		}
	}

	testbs, testgs := test.StoreWithOptions(
		cmd.rootDir, goio.Discard,
		test.StoreOptions{PreprocessOnly: true, WithExtern: true, WithExamples: true, Testing: true, FixFrom: gno.GnoVerMissing},
	)

	if cmd.verbose {
		cio.ErrPrintfln("fixing directories: %v", dirs)
	}

	if !cmd.filetestsOnly {
		return fixDir(cmd, cio, dirs, testbs, testgs, "")
	} else {
		if len(dirs) != 1 {
			return fmt.Errorf("must specify one dir")
		}
		files, err := os.ReadDir(dirs[0])
		if err != nil {
			return fmt.Errorf("reading directory: %w", err)
		}
		fnames := make([]string, 0, len(files))
		for _, file := range files {
			// Ignore directories and hidden files, only include
			// allowed files & extensions, then exclude files that
			// are of the bad extensions.
			if file.IsDir() ||
				strings.HasPrefix(file.Name(), ".") ||
				!strings.HasSuffix(file.Name(), ".gno") {
				continue
			}
			fpath := filepath.Join(dirs[0], file.Name())
			if cmd.filetestsMatch != "" {
				if !strings.Contains(fpath, cmd.filetestsMatch) {
					continue
				}
			}
			fnames = append(fnames, filepath.Join(dirs[0], file.Name()))
		}
		for _, fname := range fnames {
			if cmd.verbose {
				fmt.Printf("fixing %q\n", fname)
			}
			err2 := fixDir(cmd, cio, dirs, testbs, testgs, fname)
			if err2 != nil {
				fmt.Printf("error fixing file %q: %v\n",
					fname, err2)
				err = multierr.Append(err, err2)
			}
		}
	}
	return err
}

// filetest: if cmd.filetestsOnly, a single filetest to run fixDir on.
func fixDir(cmd *fixCmd, cio commands.IO, dirs []string, testbs stypes.CommitStore, testgs gno.Store, filetest string) error {
	ppkgs := map[string]processedPackage{}
	hasError := false
	//----------------------------------------
	// FIX STAGE 1: Type-check and lint.
	for _, dir := range dirs {
		if cmd.verbose && !cmd.filetestsOnly {
			cio.ErrPrintfln("fixing %q", dir)
		}

		// Only supports directories.
		// You should fix all directories at once to avoid dependency issues.
		info, err := os.Stat(dir)
		if err == nil && !info.IsDir() {
			dir = filepath.Dir(dir)
		}

		// Read and parse gnomod.toml directly.
		fpath := filepath.Join(dir, "gnomod.toml")
		mod, err := gnomod.ParseFilepath(fpath)
		if errors.Is(err, fs.ErrNotExist) {
			// We try a lazy migration from gno.mod if it exists and is valid.
			deprecatedDotmod := filepath.Join(dir, "gno.mod")
			mod, err = gnomod.ParseFilepath(deprecatedDotmod)
			if err != nil {
				// It doesn't exist or we can't parse it.
				// Make a temporary gnomod.toml (but don't write it yet)
				modulePath, _ := determinePkgPath(nil, dir, cmd.rootDir)
				modstr := gno.GenGnoModLatest(modulePath)
				mod, err = gnomod.ParseBytes("gnomod.toml", []byte(modstr))
				if err != nil {
					panic(fmt.Errorf("unexpected panic parsing default gnomod.toml bytes: %w", err))
				}
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
				cio.ErrPrintfln("%s: unrecognized gnomod.toml version %q, skipping fix", dir, mod.GetGno())
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
		if mod.Ignore {
			cio.ErrPrintfln("%s: module is ignore, skipping fix", dir)
			continue
		}

		// See adr/pr4264_fix_transpile.md
		// FIX STEP 1: ReadMemPackage()
		// Read MemPackage with pkgPath.
		var mpkg *std.MemPackage
		pkgPath, _ := determinePkgPath(mod, dir, cmd.rootDir)
		if cmd.filetestsOnly {
			mpkg, err = gno.ReadMemPackageFromList(
				[]string{filetest}, pkgPath, gno.MPFiletests)
		} else {
			mpkg, err = gno.ReadMemPackage(dir, pkgPath, gno.MPUserAll) // stdlib not supported
		}
		if err != nil {
			printError(cio.Err(), dir, pkgPath, err)
			hasError = true
			continue
		}

		// Filter out filetests that fail type-check.
		if cmd.filetestsOnly && filterInvalidFiletest(cio, mpkg) {
			return nil // done
		}

		// Perform imports using the parent store.
		abortOnError := !cmd.filetestsOnly
		if err := test.LoadImports(testgs, mpkg, abortOnError); err != nil {
			printError(cio.Err(), dir, pkgPath, err)
			hasError = true
			continue
		}

		// Wrap in cache wrap so execution of the linter
		// doesn't impact other packages.
		newTestGnoStore := func() gno.Store {
			tcw := testbs.CacheWrap()
			tgs := testgs.BeginTransaction(tcw, tcw, nil)
			return tgs
		}

		// Handle runtime errors
		didPanic := catchPanic(dir, pkgPath, cio.Err(), func() {
			// Memo process results here.
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
			errs := lintTypeCheck(cio, dir, mpkg, gno.TypeCheckOptions{
				Getter:     newTestGnoStore(),
				TestGetter: newTestGnoStore(),
				Mode:       gno.TCGno0p0,
			})
			if errs != nil {
				// cio.ErrPrintln(errs) already printed.
				hasError = true
				return
			}

			// FIX STEP 4.a: Prepare*()
			tm := test.Machine(newTestGnoStore(), goio.Discard, pkgPath, false)
			defer tm.Release()
			// FIX STEP 4.b: Re-parse the mem package to Go AST.
			gofset, allgofs, _, _, _, errs := gno.GoParseMemPackage(mpkg)
			if errs != nil {
				cio.ErrPrintln(errs)
				hasError = true
				return // Go parse must succeed.
			}
			// FIX STEP 4.c: PrepareGno0p9() for Gno preprocessing.
			errs = gno.PrepareGno0p9(gofset, allgofs, mpkg)
			if errs != nil {
				cio.ErrPrintln(errs)
				hasError = true
				return // Prepare must succeed.
			}

			// FIX STEP 5: re-parse
			// Gno parse source fileset and test filesets.
			// The second result `fset` is ignored because we're
			// not interested in type-check veracity (but lint is).
			_, _, tfset, _tests, ftests := sourceAndTestFileset(mpkg, cmd.filetestsOnly)
			{
				// FIX STEP 6: PreprocessFiles()
				// Preprocess tfset files (w/ some *_test.gno).
				tm.Store = newTestGnoStore()
				pn, _ := tm.PreprocessFiles(
					mpkg.Name, mpkg.Path, tfset, false, false, gno.GnoVerMissing)
				ppkg.AddTest(pn, tfset)

				// FIX STEP 7: FindXforms():
				// FindXforms for all files if outdated.
				// Use the preprocessor to collect the
				// transformations needed to be done.
				// They are collected in
				// pn.GetAttribute("XREALMFORM")
				for _, fn := range tfset.Files {
					gno.FindXformsGno0p9(tm.Store, pn, fn)
					gno.FindMoreXformsGno0p9(tm.Store, pn, pn, fn)
				}
				for { // continue to find more until exhausted.
					xnewSum := 0
					for _, fn := range tfset.Files {
						xnew := gno.FindMoreXformsGno0p9(tm.Store, pn, pn, fn)
						xnewSum += xnew
					}
					if xnewSum == 0 {
						break // done
					}
				}
			}
			{
				// FIX STEP 6: PreprocessFiles()
				// Preprocess xxx_test files (all xxx_test *_test.gno).
				tm.Store = newTestGnoStore()
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
					gno.FindXformsGno0p9(tm.Store, pn, fn)
					gno.FindMoreXformsGno0p9(tm.Store, pn, pn, fn)
				}
				for { // continue to find more until exhausted.
					xnewSum := 0
					for _, fn := range _tests.Files {
						xnew := gno.FindMoreXformsGno0p9(tm.Store, pn, pn, fn)
						xnewSum += xnew
					}
					if xnewSum == 0 {
						break // done
					}
				}
			}
			{
				// FIX STEP 6: PreprocessFiles()
				// Preprocess _filetest.gno files.
				for i, fset := range ftests {
					tm.Store = newTestGnoStore()
					fname := fset.Files[0].FileName
					mfile := mpkg.GetFile(fname)
					pkgPath := fmt.Sprintf("%s_filetest%d", mpkg.Path, i)
					pkgPath, err = parsePkgPathDirective(mfile.Body, pkgPath)
					if err != nil {
						cio.ErrPrintln(err)
						hasError = true
						continue
					}
					pkgName := string(fset.Files[0].PkgName)
					pn, _ := tm.PreprocessFiles(
						pkgName, pkgPath, fset,
						false, false, gno.GnoVerMissing)
					ppkg.AddFileTest(pn, fset)

					// FIX STEP 7: FindXforms():
					// FindXforms for all files if outdated.
					// Use the preprocessor to collect the
					// transformations needed to be done.
					// They are collected in
					// pn.GetAttribute("XREALMFORM")
					for _, fn := range fset.Files {
						gno.FindXformsGno0p9(tm.Store, pn, fn)
						gno.FindMoreXformsGno0p9(tm.Store, pn, pn, fn)
					}
					for { // continue to find more until exhausted.
						xnewSum := 0
						for _, fn := range fset.Files {
							xnew := gno.FindMoreXformsGno0p9(tm.Store, pn, pn, fn)
							xnewSum += xnew
						}
						if xnewSum == 0 {
							break // done
						}
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
		if mod != nil && mod.GetGno() != gno.GnoVerMissing {
			panic("should not happen")
		}

		// FIX STEP 8 & 9: gno.TranspileGno0p9() Part 1 & 2
		mpkg := ppkg.mpkg
		transpileProcessedFileSet := func(pfs processedFileSet) error {
			pn, fset := pfs.pn, pfs.fset
			xforms1, _ := pn.GetAttribute(gno.ATTR_PN_XFORMS).(map[string]struct{})
			err = gno.TranspileGno0p9(mpkg, dir, pn, fset.GetFileNames(), xforms1)
			return err
		}
		err = transpileProcessedFileSet(ppkg.test)
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

		// Write version to gnomod.toml.
		mod, err := gno.ParseCheckGnoMod(ppkg.mpkg)
		if err != nil {
			panic(fmt.Sprintf("unhandled error: %v", err))
		}
		if mod == nil {
			panic("XXX: generate default gnomod.toml")
		}
		mod.SetGno(gno.GnoVerLatest)
		ppkg.mpkg.SetFile("gnomod.toml", mod.WriteString())
		// Cleanup gno.mod if it exists.
		ppkg.mpkg.DeleteFile("gno.mod")

		// FIX STEP 10: mpkg.WriteTo():
		err = ppkg.mpkg.WriteTo(dir)
		if err != nil {
			return err
		}
	}

	return nil
}

// Returns true if mpkg has a filetest that has a TypeCheckError directive,
// or has a type-check-like Error directive. Panics if it has anything
// but one testfile.
func filterInvalidFiletest(cio commands.IO, mpkg *std.MemPackage) bool {
	if len(mpkg.Files) != 1 {
		panic("expected 1 filetest but got something else")
	}
	mfile := mpkg.Files[0]
	dirs, err := test.ParseDirectives(bytes.NewReader([]byte(mfile.Body)))
	if err != nil {
		panic(fmt.Errorf("error parsing directives: %w", err))
	}
	// Filter filetests with Go type-check error.
	tcErr := dirs.FirstDefault(test.DirectiveTypeCheckError, "")
	if tcErr != "" {
		cio.Printfln("skipping filetest with type-check error %q", mfile.Name)
		return true
	}
	// Filter filetests with type-check-ish Error directives.
	// (most Error directives are fine).
	// Not sure why Go type-check doesn't catch this.
	dErr := dirs.FirstDefault(test.DirectiveError, "")
	if dErr != "" && strings.Contains(dErr, "import cycle detected") ||
		strings.Contains(dErr, "exceeded maximum VPBlock depth") ||
		strings.Contains(dErr, "cannot import realm path") ||
		strings.Contains(dErr, "cannot import stdlib internal") ||
		strings.Contains(dErr, "internal/ packages can only be") ||
		strings.Contains(dErr, "cannot find branch label") ||
		strings.Contains(dErr, "but is not natively defined") ||
		strings.Contains(dErr, "goroutines are not permitted") {
		cio.Printfln("skipping filetest with type-check-ish error %q", mfile.Name)
		return true
	}
	return false
}
