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
	"runtime/debug"
	"sort"
	"strings"
	"text/template"
	"time"

	"go.uber.org/multierr"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/transpiler"
	"github.com/gnolang/gno/gnovm/tests"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/random"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
)

type testCfg struct {
	verbose             bool
	rootDir             string
	run                 string
	timeout             time.Duration
	transpile           bool // TODO: transpile should be the default, but it needs to automatically transpile dependencies in memory.
	updateGoldenTests   bool
	printRuntimeMetrics bool
	withNativeFallback  bool
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
specific instructions that can be added using code comments.

"*_filetest.gno" must be declared in the 'main' package and so must have a
'main' function, that will be executed to test the target contract.

List of available instructions that can be used in "*_filetest.gno" files:
	- "PKGPATH:" is a single line instruction that can be used to define the
	package used to interact with the tested package. If not specified, "main" is
	used.
	- "MAXALLOC:" is a signle line instruction that can be used to define a limit
	to the VM allocator. If this limit is exceeded, the VM will panic. Default to
	0, no limit.
	- "SEND:" is a single line instruction that can be used to send an amount of
	token along with the transaction. The format is for example "1000000ugnot".
	Default is empty.
	- "Output:\n" (*) is a multiple lines instruction that can be used to assert
	the output of the "*_filetest.gno" file. Any prints executed inside the
	'main' function must match the lines that follows the "Output:\n"
	instruction, or else the test fails.
	- "Error:\n" works similarly to "Output:\n", except that it asserts the
	stderr of the program, which in that case, comes from the VM because of a
	panic, rather than the 'main' function.
	- "Realm:\n" (*) is a multiple lines instruction that can be used to assert
	what has been recorded in the store following the execution of the 'main'
	function.

(*) The 'update-golden-tests' flag can be set to fill out the content of the
instruction with the actual content of the test instead of failing.
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
		&c.transpile,
		"transpile",
		false,
		"transpile gno to go before testing",
	)

	fs.BoolVar(
		&c.updateGoldenTests,
		"update-golden-tests",
		false,
		`writes actual as wanted for "Output:" and "Realm:" instructions`,
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

func execTest(cfg *testCfg, args []string, io commands.IO) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	verbose := cfg.verbose

	tempdirRoot, err := os.MkdirTemp("", "gno-transpile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempdirRoot)

	// go.mod
	modPath := filepath.Join(tempdirRoot, "go.mod")
	err = makeTestGoMod(modPath, transpiler.ImportPrefix, "1.21")
	if err != nil {
		return fmt.Errorf("write .mod file: %w", err)
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

	buildErrCount := 0
	testErrCount := 0
	for _, pkg := range subPkgs {
		if cfg.transpile {
			if verbose {
				io.ErrPrintfln("=== PREC  %s", pkg.Dir)
			}
			transpileOpts := newTranspileOptions(&transpileCfg{
				output: tempdirRoot,
			})
			err := transpilePkg(importPath(pkg.Dir), transpileOpts)
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
			err = goBuildFileOrPkg(tempDir, defaultTranspileCfg)
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
	io commands.IO,
) error {
	var (
		verbose             = cfg.verbose
		rootDir             = cfg.rootDir
		runFlag             = cfg.run
		printRuntimeMetrics = cfg.printRuntimeMetrics

		stdin  = io.In()
		stdout = io.Out()
		stderr = io.Err()
		errs   error
	)

	mode := tests.ImportModeStdlibsOnly
	if cfg.withNativeFallback {
		// XXX: display a warn?
		mode = tests.ImportModeStdlibsPreferred
	}
	if !verbose {
		// TODO: speedup by ignoring if filter is file/*?
		mockOut := bytes.NewBufferString("")
		stdout = commands.WriteNopCloser(mockOut)
	}

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
				io.ErrPrintfln("--- WARNING: unable to read package path from gno.mod or gno root directory; try creating a gno.mod file")
				gnoPkgPath = transpiler.GnoRealmPkgsPrefixBefore + random.RandStr(8)
			}
		}
		memPkg := gno.ReadMemPackage(pkgPath, gnoPkgPath)

		// tfiles, ifiles := gno.ParseMemPackageTests(memPkg)
		tfiles, ifiles := parseMemPackageTests(memPkg)
		testPkgName := getPkgNameFromFileset(ifiles)

		// run test files in pkg
		if len(tfiles.Files) > 0 {
			testStore := tests.TestStore(
				rootDir, "",
				stdin, stdout, stderr,
				mode,
			)
			if verbose {
				testStore.SetLogStoreOps(true)
			}

			m := tests.TestMachine(testStore, stdout, gnoPkgPath)
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

		// test xxx_test pkg
		if len(ifiles.Files) > 0 {
			testStore := tests.TestStore(
				rootDir, "",
				stdin, stdout, stderr,
				mode,
			)
			if verbose {
				testStore.SetLogStoreOps(true)
			}

			m := tests.TestMachine(testStore, stdout, testPkgName)

			memFiles := make([]*std.MemFile, 0, len(ifiles.FileNames())+1)
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
			m.RunMemPackage(memPkg, true)

			err := runTestFiles(m, ifiles, testPkgName, verbose, printRuntimeMetrics, runFlag, io)
			if err != nil {
				errs = multierr.Append(errs, err)
			}
		}
	}

	// testing with *_filetest.gno
	{
		filter := splitRegexp(runFlag)
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

func runTestFiles(
	m *gno.Machine,
	files *gno.FileSet,
	pkgName string,
	verbose bool,
	printRuntimeMetrics bool,
	runFlag string,
	io commands.IO,
) (errs error) {
	defer func() {
		if r := recover(); r != nil {
			errs = multierr.Append(fmt.Errorf("panic: %v\nstack:\n%v\ngno machine: %v", r, string(debug.Stack()), m.String()), errs)
		}
	}()

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
	n := gno.MustParseFile("main_test.gno", testmain)
	m.RunFiles(n)

	for _, test := range testFuncs.Tests {
		testFuncStr := fmt.Sprintf("%q", test.Name)

		eval := m.Eval(gno.Call("runtest", testFuncStr))

		ret := eval[0].GetString()
		if ret == "" {
			err := errors.New("failed to execute unit test: %q", test.Name)
			errs = multierr.Append(errs, err)
			io.ErrPrintfln("--- FAIL: %s [internal gno testing error]", test.Name)
			continue
		}

		// TODO: replace with amino or send native type?
		var rep report
		err = json.Unmarshal([]byte(ret), &rep)
		if err != nil {
			errs = multierr.Append(errs, err)
			io.ErrPrintfln("--- FAIL: %s [internal gno testing error]", test.Name)
			continue
		}

		if rep.Failed {
			err := errors.New("failed: %q", test.Name)
			errs = multierr.Append(errs, err)
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
	Failed  bool
	Skipped bool
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
