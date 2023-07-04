package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/tests"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"go.uber.org/multierr"
)

type testCfg struct {
	verbose             bool
	rootDir             string
	run                 string
	timeout             time.Duration
	precompile          bool // TODO: precompile should be the default, but it needs to automatically precompile dependencies in memory.
	updateGoldenTests   bool
	printRuntimeMetrics bool
	withNativeFallback  bool
}

func newTestCmd(io *commands.IO) *commands.Command {
	cfg := &testCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "test",
			ShortUsage: "test [flags] <package> [<package>...]",
			ShortHelp:  "Runs the tests for the specified packages",
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
		"verbose",
		false,
		"verbose output when running",
	)

	fs.BoolVar(
		&c.precompile,
		"precompile",
		false,
		"precompile gno to go before testing",
	)

	fs.BoolVar(
		&c.updateGoldenTests,
		"update-golden-tests",
		false,
		"writes actual as wanted in test comments",
	)

	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"",
		"clone location of github.com/gnolang/gno (gnodev tries to guess it)",
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
		&c.withNativeFallback,
		"with-native-fallback",
		false,
		"use stdlibs/* if present, otherwise use supported native Go packages",
	)

	fs.BoolVar(
		&c.printRuntimeMetrics,
		"print-runtime-metrics",
		false,
		"print runtime metrics (gas, memory, cpu cycles)",
	)
}

func execTest(cfg *testCfg, args []string, io *commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	verbose := cfg.verbose

	tempdirRoot, err := os.MkdirTemp("", "gno-precompile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempdirRoot)

	// go.mod
	modPath := filepath.Join(tempdirRoot, "go.mod")
	err = makeTestGoMod(modPath, gno.ImportPrefix, "1.19")
	if err != nil {
		return fmt.Errorf("write .mod file: %w", err)
	}

	// guess opts.RootDir
	if cfg.rootDir == "" {
		cfg.rootDir = guessRootDir()
	}

	paths, err := gnoPackagePathsFromPattern(args)
	if err != nil {
		return fmt.Errorf("list package paths from args: %w", err)
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

	buildErrCount := 0
	testErrCount := 0
	for _, pkg := range subPkgs {
		if cfg.precompile {
			if verbose {
				io.ErrPrintfln("=== PREC  %s", pkg.Dir)
			}
			precompileOpts := newPrecompileOptions(&precompileCfg{
				output: tempdirRoot,
			})
			err := precompilePkg(importPath(pkg.Dir), precompileOpts)
			if err != nil {
				io.ErrPrintln(err)
				io.ErrPrintln("FAIL")
				io.ErrPrintfln("FAIL    %s", pkg.Dir)
				io.ErrPrintln("FAIL")

				buildErrCount++
				continue
			}

			if verbose {
				io.ErrPrintfln("=== BUILD %s", pkg.Dir)
			}
			tempDir, err := ResolvePath(tempdirRoot, importPath(pkg.Dir))
			if err != nil {
				return errors.New("cannot resolve build dir")
			}
			err = goBuildFileOrPkg(tempDir, defaultBuildOptions)
			if err != nil {
				io.ErrPrintln(err)
				io.ErrPrintln("FAIL")
				io.ErrPrintfln("FAIL    %s", pkg.Dir)
				io.ErrPrintln("FAIL")

				buildErrCount++
				continue
			}
		}

		if len(pkg.TestGnoFiles) == 0 && len(pkg.FiletestGnoFiles) == 0 {
			io.ErrPrintfln("?       %s \t[no test files]", pkg.Dir)
			continue
		}

		sort.Strings(pkg.TestGnoFiles)
		sort.Strings(pkg.FiletestGnoFiles)

		startedAt := time.Now()
		err = gnoTestPkg(pkg.Dir, pkg.TestGnoFiles, pkg.FiletestGnoFiles, cfg, io)
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

func gnoTestPkg(
	pkgPath string,
	unittestFiles,
	filetestFiles []string,
	cfg *testCfg,
	io *commands.IO,
) error {
	var (
		verbose             = cfg.verbose
		rootDir             = cfg.rootDir
		runFlag             = cfg.run
		printRuntimeMetrics = cfg.printRuntimeMetrics

		stdin  = io.In
		stdout = io.Out
		stderr = io.Err
	)

	filter := splitRegexp(runFlag)
	var errs error

	mode := tests.ImportModeStdlibsOnly
	if cfg.withNativeFallback {
		// XXX: display a warn?
		mode = tests.ImportModeStdlibsPreferred
	}
	testStore := tests.TestStore(
		rootDir, "",
		stdin, stdout, stderr,
		mode,
	)
	if verbose {
		testStore.SetLogStoreOps(true)
	}

	if !verbose {
		// TODO: speedup by ignoring if filter is file/*?
		mockOut := bytes.NewBufferString("")
		stdout = commands.WriteNopCloser(mockOut)
	}

	// testing with *_test.gno
	if len(unittestFiles) > 0 {
		memPkg := gno.ReadMemPackage(pkgPath, pkgPath)

		// tfiles, ifiles := gno.ParseMemPackageTests(memPkg)
		tfiles, ifiles := parseMemPackageTests(memPkg)

		// run test files in pkg
		{
			m := tests.TestMachine(testStore, stdout, "main")
			if printRuntimeMetrics {
				// from tm2/pkg/sdk/vm/keeper.go
				// XXX: make maxAllocTx configurable.
				maxAllocTx := int64(500 * 1000 * 1000)

				m.Alloc = gno.NewAllocator(maxAllocTx)
			}
			m.RunMemPackage(memPkg, true)
			err := runTestFiles(m, tfiles, memPkg.Name, verbose, printRuntimeMetrics, runFlag, io)
			if err != nil {
				errs = multierr.Append(errs, err)
			}
		}

		// run test files in xxx_test pkg
		{
			testPkgName := getPkgNameFromFileset(ifiles)
			if testPkgName != "" {
				m := tests.TestMachine(testStore, stdout, testPkgName)
				m.RunMemPackage(memPkg, true)
				err := runTestFiles(m, ifiles, testPkgName, verbose, printRuntimeMetrics, runFlag, io)
				if err != nil {
					errs = multierr.Append(errs, err)
				}
			}
		}
	}

	// testing with *_filetest.gno
	{
		for _, testFile := range filetestFiles {
			testFileName := filepath.Base(testFile)
			testName := "file/" + testFileName
			if !shouldRun(filter, testName) {
				continue
			}

			startedAt := time.Now()
			if verbose {
				io.ErrPrintfln("=== RUN   %s", testName)
			}

			var closer func() (string, error)
			if !verbose {
				closer = testutils.CaptureStdoutAndStderr()
			}

			testFilePath := filepath.Join(pkgPath, testFileName)
			err := tests.RunFileTest(rootDir, testFilePath, tests.WithSyncWanted(cfg.updateGoldenTests))
			duration := time.Since(startedAt)
			dstr := fmtDuration(duration)

			if err != nil {
				errs = multierr.Append(errs, err)
				io.ErrPrintfln("--- FAIL: %s (%s)", testName, dstr)
				if verbose {
					stdouterr, err := closer()
					if err != nil {
						panic(err)
					}
					fmt.Fprintln(os.Stderr, stdouterr)
				}
				continue
			}

			if verbose {
				io.ErrPrintfln("--- PASS: %s (%s)", testName, dstr)
			}
			// XXX: add per-test metrics
		}
	}

	return errs
}

func runTestFiles(
	m *gno.Machine,
	files *gno.FileSet,
	pkgName string,
	verbose bool,
	printRuntimeMetrics bool,
	runFlag string,
	io *commands.IO,
) error {
	var errs error

	testFuncs := &testFuncs{
		PackageName: pkgName,
		Verbose:     verbose,
		RunFlag:     runFlag,
	}
	loadTestFuncs(pkgName, testFuncs, files)

	// before/after statistics
	numPackagesBefore := m.Store.NumMemPackages()

	testmain, err := formatTestmain(testFuncs)
	if err != nil {
		log.Fatal(err)
	}

	m.RunFiles(files.Files...)
	n := gno.MustParseFile("testmain.go", testmain)
	m.RunFiles(n)

	for _, test := range testFuncs.Tests {
		if verbose {
			io.ErrPrintfln("=== RUN   %s", test.Name)
		}

		testFuncStr := fmt.Sprintf("%q", test.Name)

		startedAt := time.Now()
		eval := m.Eval(gno.Call("runtest", testFuncStr))
		duration := time.Since(startedAt)
		dstr := fmtDuration(duration)

		ret := eval[0].GetString()
		if ret == "" {
			err := errors.New("failed to execute unit test: %q", test.Name)
			errs = multierr.Append(errs, err)
			io.ErrPrintfln("--- FAIL: %s (%v)", test.Name, duration)
			continue
		}

		// TODO: replace with amino or send native type?
		var rep report
		err = json.Unmarshal([]byte(ret), &rep)
		if err != nil {
			errs = multierr.Append(errs, err)
			io.ErrPrintfln("--- FAIL: %s (%s)", test.Name, dstr)
			continue
		}

		switch {
		case rep.Filtered:
			io.ErrPrintfln("--- FILT: %s", test.Name)
			// noop
		case rep.Skipped:
			if verbose {
				io.ErrPrintfln("--- SKIP: %s", test.Name)
			}
		case rep.Failed:
			err := errors.New("failed: %q", test.Name)
			errs = multierr.Append(errs, err)
			io.ErrPrintfln("--- FAIL: %s (%s)", test.Name, dstr)
		default:
			if verbose {
				io.ErrPrintfln("--- PASS: %s (%s)", test.Name, dstr)
			}
		}

		if rep.Output != "" && (verbose || rep.Failed) {
			io.ErrPrintfln("output: %s", rep.Output)
		}

		if printRuntimeMetrics {
			imports := m.Store.NumMemPackages() - numPackagesBefore - 1
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
			io.ErrPrintfln("---       runtime: cycle=%s imports=%d allocs=%s",
				prettySize(m.Cycles),
				imports,
				allocsVal,
			)
		}
	}

	return errs
}

// mirror of stdlibs/testing.Report
type report struct {
	Name     string
	Verbose  bool
	Failed   bool
	Skipped  bool
	Filtered bool
	Output   string
}

var testmainTmpl = template.Must(template.New("testmain").Parse(`
package {{ .PackageName }}

import (
	"testing"
)

var tests = []testing.InternalTest{
{{range .Tests}}
    {"{{.Name}}", {{.Name}}},
{{end}}
}

func runtest(name string) (report string) {
	for _, test := range tests {
		if test.Name == name {
			return testing.RunTest({{printf "%q" .RunFlag}}, {{.Verbose}}, test)
		}
	}
	panic("no such test: " + name)
	return ""
}
`))

type testFuncs struct {
	Tests       []testFunc
	PackageName string
	Verbose     bool
	RunFlag     string
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

func formatTestmain(t *testFuncs) (string, error) {
	var buf bytes.Buffer
	if err := testmainTmpl.Execute(&buf, t); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func loadTestFuncs(pkgName string, t *testFuncs, tfiles *gno.FileSet) *testFuncs {
	for _, tf := range tfiles.Files {
		for _, d := range tf.Decls {
			if fd, ok := d.(*gno.FuncDecl); ok {
				fname := string(fd.Name)
				if strings.HasPrefix(fname, "Test") {
					tf := testFunc{
						Package: pkgName,
						Name:    fname,
					}
					t.Tests = append(t.Tests, tf)
				}
			}
		}
	}
	return t
}

// parseMemPackageTests is copied from gno.ParseMemPackageTests
// for except to _filetest.gno
func parseMemPackageTests(memPkg *std.MemPackage) (tset, itset *gno.FileSet) {
	tset = &gno.FileSet{}
	itset = &gno.FileSet{}
	for _, mfile := range memPkg.Files {
		if !strings.HasSuffix(mfile.Name, ".gno") {
			continue // skip this file.
		}
		if strings.HasSuffix(mfile.Name, "_filetest.gno") {
			continue
		}
		n, err := gno.ParseFile(mfile.Name, mfile.Body)
		if err != nil {
			panic(errors.Wrap(err, "parsing file "+mfile.Name))
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
