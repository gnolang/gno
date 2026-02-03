package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	goio "io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gnolang/gno/gnovm/cmd/gno/internal/cmdutil"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/lint"
	"github.com/gnolang/gno/gnovm/pkg/lint/reporters"
	_ "github.com/gnolang/gno/gnovm/pkg/lint/rules"
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
	verbose      bool
	rootDir      string
	autoGnomod   bool
	mode         string
	format       string
	listRules    bool
	disableRules string
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

	fs.BoolVar(&c.verbose, "v", false, "verbose output when linting")
	fs.StringVar(&c.rootDir, "root-dir", rootdir, "clone location of github.com/gnolang/gno (gno tries to guess it)")
	fs.BoolVar(&c.autoGnomod, "auto-gnomod", true, "auto-generate gnomod.toml file if not already present")
	fs.StringVar(&c.mode, "mode", "default", "lint mode: default, strict, warn-only")
	fs.StringVar(&c.format, "format", "text", "output format: text, json")
	fs.BoolVar(&c.listRules, "list-rules", false, "list available lint rules and exit")
	fs.StringVar(&c.disableRules, "disable-rules", "", "comma-separated list of rules to disable (e.g., AVL001,GLOBAL001)")
}

func execLint(cmd *lintCmd, args []string, io commands.IO) error {
	// Handle --list-rules first
	if cmd.listRules {
		return listLintRules(io)
	}

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

	reporter, err := reporters.NewReporter(cmd.format, io.Err())
	if err != nil {
		return err
	}

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
	// LINT STAGE 1: Preprocessing.
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
			reportIssue(reporter, gnoGnoModError, err.Error(), fpath)
			reporter.Flush()
			return commands.ExitCodeError(1)
		}

		// See adr/pr4264_lint_transpile.md
		// STEP 1.1: ReadMemPackage()
		// Read MemPackage with pkgPath.
		pkgPath, _ := determinePkgPath(mod, dir, cmd.rootDir)
		mpkg, err := gno.ReadMemPackage(dir, pkgPath, gno.MPAnyAll)
		if err != nil {
			reportError(reporter, dir, pkgPath, err)
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
			reportError(reporter, dir, pkgPath, err)
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
		didPanic := catchPanicWithReporter(reporter, dir, pkgPath, func() {
			// Memo process results here.
			ppkg := cmdutil.ProcessedPackage{MPkg: mpkg, Dir: dir}

			// Run type checking
			// STEP 1.2: ParseGnoMod()
			// STEP 1.3: GoParse*()
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
			errs := lintTypeCheck(reporter, dir, mpkg, gno.TypeCheckOptions{
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

			// STEP 1.4: re-parse for preprocessor.
			// While lintTypeCheck > TypeCheckMemPackage will find
			// most issues, the preprocessor may have additional
			// checks.
			// Gno parse source fileset and test filesets.
			_, fset, tfset, _tests, ftests := sourceAndTestFileset(mpkg, false)

			{
				// STEP 1.5: PreprocessFiles()
				// Preprocess fset files (no test files)
				tm.Store = newProdGnoStore()
				pn, _ := tm.PreprocessFiles(
					mpkg.Name, mpkg.Path, fset, false, false, "")
				ppkg.AddNormal(pn, fset)
			}
			{
				// STEP 1.5: PreprocessFiles()
				// Preprocess fset files (w/ some *_test.gno).
				tm.Store = newTestGnoStore(false)
				pn, _ := tm.PreprocessFiles(
					mpkg.Name, mpkg.Path, tfset, false, false, "")
				ppkg.AddTest(pn, fset)
			}
			{
				// STEP 1.5: PreprocessFiles()
				// Preprocess _test files (all xxx_test *_test.gno).
				tm.Store = newTestGnoStore(true)
				pn, _ := tm.PreprocessFiles(
					mpkg.Name+"_test", mpkg.Path+"_test", _tests, false, false, "")
				ppkg.AddUnderscoreTests(pn, _tests)
			}
			{
				// STEP 1.5: PreprocessFiles()
				// Preprocess _filetest.gno files.
				for i, fset := range ftests {
					tm.Store = newTestGnoStore(true)
					fname := fset.Files[0].FileName
					mfile := mpkg.GetFile(fname)
					pkgPath := fmt.Sprintf("%s_filetest%d", mpkg.Path, i)
					pkgPath, err = parsePkgPathDirective(mfile.Body, pkgPath)
					if err != nil {
						reportError(reporter, dir, pkgPath, err)
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
		reporter.Flush()
		return commands.ExitCodeError(1)
	}

	//----------------------------------------
	// LINT STAGE 2: Lint rules.
	lintCfg := lint.DefaultConfig()
	switch cmd.mode {
	case "default":
		lintCfg.Mode = lint.ModeDefault
	case "strict":
		lintCfg.Mode = lint.ModeStrict
	case "warn-only":
		lintCfg.Mode = lint.ModeWarnOnly
	default:
		return fmt.Errorf("invalid lint mode: %q", cmd.mode)
	}

	if cmd.disableRules != "" {
		for _, rule := range strings.Split(cmd.disableRules, ",") {
			lintCfg.Disable[rule] = true
		}
	}

	engine := lint.NewEngine(lintCfg, lint.DefaultRegistry, reporter)

	for _, ppkg := range ppkgs {
		sources := make(map[string]string)
		for _, mf := range ppkg.MPkg.Files {
			sources[mf.Name] = mf.Body
		}

		if ppkg.Prod.Fset != nil {
			engine.Run(ppkg.Prod.Fset, sources)
		}
		if ppkg.Test.Fset != nil {
			engine.Run(ppkg.Test.Fset, sources)
		}
		if ppkg.XTest.Fset != nil {
			engine.Run(ppkg.XTest.Fset, sources)
		}
		for _, ftest := range ppkg.FTest {
			if ftest.Fset != nil {
				engine.Run(ftest.Fset, sources)
			}
		}
	}

	if err := engine.Flush(); err != nil {
		return err
	}

	_, _, lintErrors := engine.Summary()
	if lintErrors > 0 {
		hasError = true
	}

	//----------------------------------------
	// LINT STAGE 3: Write.
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

		// STEP 3.1: mpkg.WriteTo()
		err := ppkg.MPkg.WriteTo(pkg.Dir)
		if err != nil {
			return err
		}
	}

	if hasError {
		return commands.ExitCodeError(1)
	}

	return nil
}

func lintTypeCheck(
	reporter lint.Reporter,
	dir string,
	mpkg *std.MemPackage,
	opts gno.TypeCheckOptions,
) (lerr error) {
	_, tcErrs := gno.TypeCheckMemPackage(mpkg, opts)

	errors := multierr.Errors(tcErrs)
	for _, err := range errors {
		reportError(reporter, dir, mpkg.Path, err)
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

func listLintRules(io commands.IO) error {
	rules := lint.DefaultRegistry.All()

	ruleInfos := make([]lint.RuleInfo, len(rules))
	for i, r := range rules {
		ruleInfos[i] = r.Info()
	}
	sort.Slice(ruleInfos, func(i, j int) bool {
		return ruleInfos[i].ID < ruleInfos[j].ID
	})

	io.Println("Available lint rules:\n")
	for _, info := range ruleInfos {
		io.Printf("  %-12s %-30s (%s)\n", info.ID, info.Name, info.Severity)
	}
	return nil
}
