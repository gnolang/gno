package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	goio "io"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/test"
	teststd "github.com/gnolang/gno/gnovm/tests/stdlibs/std"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/random"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"go.uber.org/multierr"
)

type testCfg struct {
	verbose             bool
	rootDir             string
	run                 string
	timeout             time.Duration
	updateGoldenTests   bool
	printRuntimeMetrics bool
	printEvents         bool
}

func newTestCmd(io commands.IO) *commands.Command {
	cfg := &testCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "test",
			ShortUsage: "test [flags] <package> [<package>...]",
			ShortHelp:  "runs the tests for the specified packages",
			LongHelp: `Runs the tests for the specified packages.

'gno test' recompiles each package along with any files with names matching the
file pattern "*_test.gno" or "*_filetest.gno".

The <package> can be directory or file path (relative or absolute).

- "*_test.gno" files work like "*_test.go" files, but they contain only test
functions. Benchmark and fuzz functions aren't supported yet. Similarly, only
tests that belong to the same package are supported for now (no "xxx_test").

The package path used to execute the "*_test.gno" file is fetched from the
module name found in 'gno.mod', or else it is randomly generated like
"gno.land/r/XXXXXXXX".

- "*_filetest.gno" files on the other hand are kind of unique. They exist to
provide a way to interact and assert a gno contract, thanks to a set of
specific directives that can be added using code comments.

"*_filetest.gno" must be declared in the 'main' package and so must have a
'main' function, that will be executed to test the target contract.

These single-line directives can set "input parameters" for the machine used
to perform the test:
	- "PKGPATH:" is a single line directive that can be used to define the
	package used to interact with the tested package. If not specified, "main" is
	used.
	- "MAXALLOC:" is a single line directive that can be used to define a limit
	to the VM allocator. If this limit is exceeded, the VM will panic. Default to
	0, no limit.
	- "SEND:" is a single line directive that can be used to send an amount of
	token along with the transaction. The format is for example "1000000ugnot".
	Default is empty.

These directives, instead, match the comment that follows with the result
of the GnoVM, acting as a "golden test":
	- "Output:" tests the following comment with the standard output of the
	filetest.
	- "Error:" tests the following comment with any panic, or other kind of
	error that the filetest generates (like a parsing or preprocessing error).
	- "Realm:" tests the following comment against the store log, which can show
	what realm information is stored.
	- "Stacktrace:" can be used to verify the following lines against the
	stacktrace of the error.
	- "Events:" can be used to verify the emitted events against a JSON.

To speed up execution, imports of pure packages are processed separately from
the execution of the tests. This makes testing faster, but means that the
initialization of imported pure packages cannot be checked in filetests.
`,
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execTest(cfg, args, io)
		},
	)
}

func (c *testCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.verbose,
		"v",
		false,
		"verbose output when running",
	)

	fs.BoolVar(
		&c.updateGoldenTests,
		"update-golden-tests",
		false,
		`writes actual as wanted for "golden" directives in filetests`,
	)

	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"",
		"clone location of github.com/gnolang/gno (gno tries to guess it)",
	)

	fs.StringVar(
		&c.run,
		"run",
		"",
		"test name filtering pattern",
	)

	fs.DurationVar(
		&c.timeout,
		"timeout",
		0,
		"max execution time",
	)

	fs.BoolVar(
		&c.printRuntimeMetrics,
		"print-runtime-metrics",
		false,
		"print runtime metrics (gas, memory, cpu cycles)",
	)

	fs.BoolVar(
		&c.printEvents,
		"print-events",
		false,
		"print emitted events",
	)
}

// proxyWriter is a simple wrapper around a io.Writer, it exists so that the
// underlying writer can be swapped with another when necessary.
type proxyWriter struct {
	goio.Writer
}

// tee temporarily appends the writer w to an underlying MultiWriter, which
// should then be reverted using revert().
func (p *proxyWriter) tee(w goio.Writer) (revert func()) {
	save := p.Writer
	if save == goio.Discard {
		p.Writer = w
	} else {
		p.Writer = goio.MultiWriter(save, w)
	}
	return func() {
		p.Writer = save
	}
}

func execTest(cfg *testCfg, args []string, io commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	// guess opts.RootDir
	if cfg.rootDir == "" {
		cfg.rootDir = gnoenv.RootDir()
	}

	paths, err := targetsFromPatterns(args)
	if err != nil {
		return fmt.Errorf("list targets from patterns: %w", err)
	}
	if len(paths) == 0 {
		io.ErrPrintln("no packages to test")
		return nil
	}

	if cfg.timeout > 0 {
		go func() {
			time.Sleep(cfg.timeout)
			panic("test timed out after " + cfg.timeout.String())
		}()
	}

	subPkgs, err := gnomod.SubPkgsFromPaths(paths)
	if err != nil {
		return fmt.Errorf("list sub packages: %w", err)
	}

	// Set-up testStore.
	// Use a proxyWriter for stdout so that filetests can plug in a writer when
	// necessary.
	pkgCfg := &testPkgCfg{
		testCfg: cfg,
		io:      io,
	}
	outWriter := &proxyWriter{goio.Discard}
	if cfg.verbose {
		outWriter.Writer = io.Out()
	}
	pkgCfg.baseStore, pkgCfg.testStore = test.Store(
		cfg.rootDir, false,
		io.In(), outWriter, io.Err(),
	)
	pkgCfg.outWriter = outWriter
	if cfg.verbose {
		pkgCfg.testStore.SetLogStoreOps(true)
	}

	buildErrCount := 0
	testErrCount := 0
	for _, pkg := range subPkgs {
		if len(pkg.TestGnoFiles) == 0 && len(pkg.FiletestGnoFiles) == 0 {
			io.ErrPrintfln("?       %s \t[no test files]", pkg.Dir)
			continue
		}

		sort.Strings(pkg.TestGnoFiles)
		sort.Strings(pkg.FiletestGnoFiles)

		startedAt := time.Now()
		err = pkgCfg.gnoTestPkg(pkg.Dir, pkg.TestGnoFiles, pkg.FiletestGnoFiles)
		duration := time.Since(startedAt)
		dstr := fmtDuration(duration)

		if err != nil {
			io.ErrPrintfln("%s: test pkg: %v", pkg.Dir, err)
			io.ErrPrintfln("FAIL")
			io.ErrPrintfln("FAIL    %s \t%s", pkg.Dir, dstr)
			io.ErrPrintfln("FAIL")
			testErrCount++
		} else {
			io.ErrPrintfln("ok      %s \t%s", pkg.Dir, dstr)
		}
	}
	if testErrCount > 0 || buildErrCount > 0 {
		io.ErrPrintfln("FAIL")
		return fmt.Errorf("FAIL: %d build errors, %d test errors", buildErrCount, testErrCount)
	}

	return nil
}

type testPkgCfg struct {
	*testCfg
	baseStore storetypes.CommitStore
	testStore gno.Store
	outWriter *proxyWriter
	io        commands.IO
}

func (cfg testPkgCfg) gnoTestPkg(
	pkgPath string,
	unittestFiles,
	filetestFiles []string,
) error {
	var (
		verbose = cfg.verbose
		rootDir = cfg.rootDir
		runFlag = cfg.run

		errs error
	)

	// testing with *_test.gno
	if len(unittestFiles) > 0 {
		// Determine gnoPkgPath by reading gno.mod
		var gnoPkgPath string
		modfile, err := gnomod.ParseAt(pkgPath)
		if err == nil {
			gnoPkgPath = modfile.Module.Mod.Path
		} else {
			gnoPkgPath = pkgPathFromRootDir(pkgPath, rootDir)
			if gnoPkgPath == "" {
				// unable to read pkgPath from gno.mod, generate a random realm path
				cfg.io.ErrPrintfln("--- WARNING: unable to read package path from gno.mod or gno root directory; try creating a gno.mod file")
				gnoPkgPath = gno.RealmPathPrefix + strings.ToLower(random.RandStr(8))
			}
		}
		memPkg := gno.ReadMemPackage(pkgPath, gnoPkgPath)

		var tfiles, ifiles *gno.FileSet

		hasError := catchRuntimeError(gnoPkgPath, cfg.io.Err(), func() {
			tfiles, ifiles = parseMemPackageTests(cfg.testStore, memPkg)
		})

		if hasError {
			return commands.ExitCodeError(1)
		}
		testPkgName := getPkgNameFromFileset(ifiles)

		// create a common cw/gs for both the `pkg` tests as well as the `pkg_test`
		// tests. this allows us to "export" symbols from the pkg tests and
		// import them from the `pkg_test` tests.
		cw := cfg.baseStore.CacheWrap()
		gs := cfg.testStore.BeginTransaction(cw, cw)

		// run test files in pkg
		if len(tfiles.Files) > 0 {
			err := cfg.runTestFiles(memPkg, tfiles, cw, gs, runFlag)
			if err != nil {
				errs = multierr.Append(errs, err)
			}
		}

		// test xxx_test pkg
		if len(ifiles.Files) > 0 {
			memFiles := make([]*gnovm.MemFile, 0, len(ifiles.FileNames())+1)
			for _, f := range memPkg.Files {
				for _, ifileName := range ifiles.FileNames() {
					if f.Name == "gno.mod" || f.Name == ifileName {
						memFiles = append(memFiles, f)
						break
					}
				}
			}
			memPkg.Files = memFiles
			memPkg.Name = testPkgName
			memPkg.Path = memPkg.Path + "_test"

			err := cfg.runTestFiles(memPkg, ifiles, cw, gs, runFlag)
			if err != nil {
				errs = multierr.Append(errs, err)
			}
		}
	}

	// testing with *_filetest.gno
	{
		opts := test.FileTestOptions{
			Store:     cfg.testStore,
			BaseStore: cfg.baseStore,
			Output:    cfg.outWriter,
		}
		revert := cfg.outWriter.tee(opts.Stdout())
		defer revert()

		filter := splitRegexp(runFlag)
		for _, testFile := range filetestFiles {
			testFileName := filepath.Base(testFile)
			testFilePath := filepath.Join(pkgPath, testFileName)
			testName := "file/" + testFileName
			if !shouldRun(filter, testName) {
				continue
			}

			startedAt := time.Now()
			if verbose {
				cfg.io.ErrPrintfln("=== RUN   %s", testName)
			}

			content, err := os.ReadFile(testFilePath)
			if err != nil {
				return err
			}

			if cfg.updateGoldenTests {
				var changed string
				changed, err = opts.RunSync(testFileName, content)
				if changed != "" {
					err = os.WriteFile(testFilePath, []byte(changed), 0o644)
					if err != nil {
						panic(fmt.Errorf("could not fix golden file: %w", err))
					}
				}
			} else {
				err = opts.Run(testFileName, content)
			}
			duration := time.Since(startedAt)
			dstr := fmtDuration(duration)
			if err != nil {
				cfg.io.ErrPrintfln("--- FAIL: %s (%s)", testName, dstr)
				cfg.io.ErrPrintln(err.Error())
				errs = multierr.Append(errs, fmt.Errorf("%s failed", testName))
			} else if verbose {
				cfg.io.ErrPrintfln("--- PASS: %s (%s)", testName, dstr)
			}

			// XXX: add per-test metrics
		}
	}

	return errs
}

// attempts to determine the full gno pkg path by analyzing the directory.
func pkgPathFromRootDir(pkgPath, rootDir string) string {
	abPkgPath, err := filepath.Abs(pkgPath)
	if err != nil {
		log.Printf("could not determine abs path: %v", err)
		return ""
	}
	abRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		log.Printf("could not determine abs path: %v", err)
		return ""
	}
	abRootDir += string(filepath.Separator)
	if !strings.HasPrefix(abPkgPath, abRootDir) {
		return ""
	}
	impPath := strings.ReplaceAll(abPkgPath[len(abRootDir):], string(filepath.Separator), "/")
	for _, prefix := range [...]string{
		"examples/",
		"gnovm/stdlibs/",
		"gnovm/tests/stdlibs/",
	} {
		if strings.HasPrefix(impPath, prefix) {
			return impPath[len(prefix):]
		}
	}
	return ""
}

func (cfg *testPkgCfg) runTestFiles(
	memPkg *gnovm.MemPackage,
	files *gno.FileSet,
	cw storetypes.Store, gs gno.TransactionStore,
	runFlag string,
) (errs error) {
	var m *gno.Machine
	defer func() {
		if r := recover(); r != nil {
			if st := m.ExceptionsStacktrace(); st != "" {
				errs = multierr.Append(errors.New(st), errs)
			}
			errs = multierr.Append(
				fmt.Errorf("panic: %v\ngo stacktrace:\n%v\ngno machine: %v\ngno stacktrace:\n%v",
					r, string(debug.Stack()), m.String(), m.Stacktrace()),
				errs,
			)
		}
	}()

	tests := loadTestFuncs(memPkg.Name, files)

	var alloc *gno.Allocator

	// run package and write to upper cw / gs
	if cfg.printRuntimeMetrics {
		alloc = gno.NewAllocator(math.MaxInt64)
	}
	// Check if we already have the package - it may have been eagerly
	// loaded.
	m = test.Machine(gs, cfg.outWriter, memPkg.Path)
	m.Alloc = alloc
	if cfg.testStore.GetMemPackage(memPkg.Path) == nil {
		m.RunMemPackage(memPkg, true)
	} else {
		// ensure to set the active package.
		m.SetActivePackage(gs.GetPackage(memPkg.Path, false))
	}
	pv := m.Package

	m.RunFiles(files.Files...)

	for _, tf := range tests {
		// TODO(morgan): we could theoretically use wrapping on the baseStore
		// and gno store to achieve per-test isolation. However, that requires
		// some deeper changes, as ideally we'd:
		// - Run the MemPackage independently (so it can also be run as a
		//   consequence of an import)
		// - Run the test files before this for loop (but persist it to store;
		//   RunFiles doesn't do that currently)
		// - Wrap here.
		m = test.Machine(gs, cfg.outWriter, memPkg.Path)
		m.Alloc = alloc
		m.SetActivePackage(pv)

		testingpv := m.Store.GetPackage("testing", false)
		testingtv := gno.TypedValue{T: &gno.PackageType{}, V: testingpv}
		testingcx := &gno.ConstExpr{TypedValue: testingtv}

		eval := m.Eval(gno.Call(
			gno.Sel(testingcx, "RunTest"),           // Call testing.RunTest
			gno.Str(runFlag),                        // run flag
			gno.Nx(strconv.FormatBool(cfg.verbose)), // is verbose?
			&gno.CompositeLitExpr{ // Third param, the testing.InternalTest
				Type: gno.Sel(testingcx, "InternalTest"),
				Elts: gno.KeyValueExprs{
					{Key: gno.X("Name"), Value: gno.Str(tf.Name)},
					{Key: gno.X("F"), Value: gno.Nx(tf.Name)},
				},
			},
		))

		if cfg.printEvents {
			events := m.Context.(*teststd.TestExecContext).EventLogger.Events()
			if events != nil {
				res, err := json.Marshal(events)
				if err != nil {
					panic(err)
				}
				cfg.io.ErrPrintfln("EVENTS: %s", string(res))
			}
		}

		ret := eval[0].GetString()
		if ret == "" {
			err := errors.New("failed to execute unit test: %q", tf.Name)
			errs = multierr.Append(errs, err)
			cfg.io.ErrPrintfln("--- FAIL: %s [internal gno testing error]", tf.Name)
			continue
		}

		// TODO: replace with amino or send native type?
		var rep report
		err := json.Unmarshal([]byte(ret), &rep)
		if err != nil {
			errs = multierr.Append(errs, err)
			cfg.io.ErrPrintfln("--- FAIL: %s [internal gno testing error]", tf.Name)
			continue
		}

		if rep.Failed {
			err := errors.New("failed: %q", tf.Name)
			errs = multierr.Append(errs, err)
		}

		if cfg.printRuntimeMetrics {
			// XXX: store changes
			// XXX: max mem consumption
			allocsVal := "n/a"
			if m.Alloc != nil {
				maxAllocs, allocs := m.Alloc.Status()
				allocsVal = fmt.Sprintf("%s(%.2f%%)",
					prettySize(allocs),
					float64(allocs)/float64(maxAllocs)*100,
				)
			}
			cfg.io.ErrPrintfln("---       runtime: cycle=%s allocs=%s",
				prettySize(m.Cycles),
				allocsVal,
			)
		}
	}

	return errs
}

// mirror of stdlibs/testing.Report
type report struct {
	Failed  bool
	Skipped bool
}

type testFunc struct {
	Package string
	Name    string
}

func getPkgNameFromFileset(files *gno.FileSet) string {
	if len(files.Files) <= 0 {
		return ""
	}
	return string(files.Files[0].PkgName)
}

func loadTestFuncs(pkgName string, tfiles *gno.FileSet) (rt []testFunc) {
	for _, tf := range tfiles.Files {
		for _, d := range tf.Decls {
			if fd, ok := d.(*gno.FuncDecl); ok {
				fname := string(fd.Name)
				if strings.HasPrefix(fname, "Test") {
					tf := testFunc{
						Package: pkgName,
						Name:    fname,
					}
					rt = append(rt, tf)
				}
			}
		}
	}
	return
}

// parses test files (skipping filetests) in the memPkg.
func parseMemPackageTests(store gno.Store, memPkg *gnovm.MemPackage) (tset, itset *gno.FileSet) {
	tset = &gno.FileSet{}
	itset = &gno.FileSet{}
	var errs error
	for _, mfile := range memPkg.Files {
		if !strings.HasSuffix(mfile.Name, ".gno") {
			continue // skip this file.
		}
		if strings.HasSuffix(mfile.Name, "_filetest.gno") {
			continue
		}

		if err := test.LoadImports(store, path.Join(memPkg.Path, mfile.Name), []byte(mfile.Body)); err != nil {
			errs = multierr.Append(errs, err)
		}

		n, err := gno.ParseFile(mfile.Name, mfile.Body)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		if n == nil {
			panic("should not happen")
		}
		if strings.HasSuffix(mfile.Name, "_test.gno") {
			// add test file.
			if memPkg.Name+"_test" == string(n.PkgName) {
				itset.AddFiles(n)
			} else {
				tset.AddFiles(n)
			}
		} else if memPkg.Name == string(n.PkgName) {
			// skip package file.
		} else {
			panic(fmt.Sprintf(
				"expected package name [%s] or [%s_test] but got [%s] file [%s]",
				memPkg.Name, memPkg.Name, n.PkgName, mfile))
		}
	}
	if errs != nil {
		panic(errs)
	}
	return tset, itset
}

func shouldRun(filter filterMatch, path string) bool {
	if filter == nil {
		return true
	}
	elem := strings.Split(path, "/")
	ok, _ := filter.matches(elem, matchString)
	return ok
}
