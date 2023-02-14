package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/gnolang/gno/pkgs/commands"
	"github.com/gnolang/gno/pkgs/errors"
	gno "github.com/gnolang/gno/pkgs/gnolang"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/gno/pkgs/testutils"
	"github.com/gnolang/gno/tests"
	"go.uber.org/multierr"
)

type testCfg struct {
	verbose    bool
	rootDir    string
	run        string
	timeout    time.Duration
	precompile bool // TODO: precompile should be the default, but it needs to automatically precompile dependencies in memory.
}

func newTestCmd() *commands.Command {
	cfg := &testCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "test",
			ShortUsage: "test [flags] <package> [<package>...]",
			ShortHelp:  "Runs the tests for the specified packages",
		},
		cfg,
		func(_ context.Context, args []string) error {
			return execTest(cfg, args)
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
}

func execTest(cfg *testCfg, args []string) error {
	if len(args) < 1 {
		return errors.New("invalid args")
	}

	verbose := cfg.verbose

	tempdirRoot, err := os.MkdirTemp("", "gno-precompile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempdirRoot)

	// go.mod
	modPath := filepath.Join(tempdirRoot, "go.mod")
	err = makeTestGoMod(modPath, gno.ImportPrefix, "1.18")
	if err != nil {
		return fmt.Errorf("write .mod file: %w", err)
	}

	// guess opts.RootDir
	if cfg.rootDir == "" {
		cfg.rootDir = guessRootDir()
	}

	pkgPaths, err := gnoPackagesFromArgs(args)
	if err != nil {
		return fmt.Errorf("list packages from args: %w", err)
	}

	if cfg.timeout > 0 {
		go func() {
			time.Sleep(cfg.timeout)
			panic("test timed out after " + cfg.timeout.String())
		}()
	}

	buildErrCount := 0
	testErrCount := 0
	for _, pkgPath := range pkgPaths {
		if cfg.precompile {
			if verbose {
				fmt.Printf("=== PREC  %s\n", pkgPath)
			}
			precompileOpts := newPrecompileOptions(&precompileCfg{
				output: tempdirRoot,
			})
			err := precompilePkg(importPath(pkgPath), precompileOpts)
			if err != nil {
				fmt.Println(err)
				fmt.Println("FAIL")
				fmt.Printf("FAIL    %s\n", pkgPath)
				fmt.Println("FAIL")

				buildErrCount++
				continue
			}

			if verbose {
				fmt.Printf("=== BUILD %s", pkgPath)
			}
			tempDir, err := ResolvePath(tempdirRoot, importPath(pkgPath))
			if err != nil {
				return errors.New("cannot resolve build dir")
			}
			err = goBuildFileOrPkg(tempDir, defaultBuildOptions)
			if err != nil {
				fmt.Println(err)
				fmt.Println("FAIL")
				fmt.Printf("FAIL    %s\n", pkgPath)
				fmt.Println("FAIL")

				buildErrCount++
				continue
			}
		}

		unittestFiles, err := filepath.Glob(filepath.Join(pkgPath, "*_test.gno"))
		if err != nil {
			log.Fatal(err)
		}
		filetestFiles, err := filepath.Glob(filepath.Join(pkgPath, "*_filetest.gno"))
		if err != nil {
			log.Fatal(err)
		}
		if len(unittestFiles) == 0 && len(filetestFiles) == 0 {
			fmt.Printf("?       %s \t[no test files]\n", pkgPath)
			continue
		}

		sort.Strings(unittestFiles)
		sort.Strings(filetestFiles)

		startedAt := time.Now()
		err = gnoTestPkg(pkgPath, unittestFiles, filetestFiles, cfg)
		duration := time.Since(startedAt)
		dstr := fmtDuration(duration)

		if err != nil {
			fmt.Printf("%s: test pkg: %v\n", pkgPath, err)
			fmt.Println("FAIL")
			fmt.Printf("FAIL    %s \t%s\n", pkgPath, dstr)
			fmt.Println("FAIL")
			testErrCount++
		} else {
			fmt.Printf("ok      %s \t%s\n", pkgPath, dstr)
		}
	}
	if testErrCount > 0 || buildErrCount > 0 {
		fmt.Println("FAIL")
		return fmt.Errorf("FAIL: %d build errors, %d test errors", buildErrCount, testErrCount)
	}

	return nil
}

func gnoTestPkg(pkgPath string, unittestFiles, filetestFiles []string, cfg *testCfg) error {
	verbose := cfg.verbose
	rootDir := cfg.rootDir
	runFlag := cfg.run
	filter := splitRegexp(runFlag)

	var errs error

	testStore := tests.TestStore(rootDir, "", os.Stdin, os.Stdout, os.Stderr, tests.ImportModeStdlibsOnly)
	if verbose {
		testStore.SetLogStoreOps(true)
	}

	// testing with *_test.gno
	if len(unittestFiles) > 0 {
		// TODO: speedup by ignoring if filter is file/*?
		var stdout io.Writer = new(bytes.Buffer)
		if verbose {
			stdout = os.Stdout
		}
		memPkg := gno.ReadMemPackage(pkgPath, pkgPath)

		// tfiles, ifiles := gno.ParseMemPackageTests(memPkg)
		tfiles, ifiles := parseMemPackageTests(memPkg)

		// run test files in pkg
		{
			m := tests.TestMachine(testStore, stdout, "main")
			m.RunMemPackage(memPkg, true)
			err := runTestFiles(m, tfiles, memPkg.Name, verbose, runFlag)
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
				err := runTestFiles(m, ifiles, testPkgName, verbose, runFlag)
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
				fmt.Printf("=== RUN   %s\n", testName)
			}

			var closer func() (string, error)
			if !verbose {
				closer = testutils.CaptureStdoutAndStderr()
			}

			testFilePath := filepath.Join(pkgPath, testFileName)
			err := tests.RunFileTest(rootDir, testFilePath, false, nil)
			duration := time.Since(startedAt)
			dstr := fmtDuration(duration)

			if err != nil {
				errs = multierr.Append(errs, err)
				fmt.Printf("--- FAIL: %s (%s)\n", testName, dstr)
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
				fmt.Printf("--- PASS: %s (%s)\n", testName, dstr)
			}
		}
	}

	return errs
}

func runTestFiles(m *gno.Machine, files *gno.FileSet, pkgName string, verbose bool, runFlag string) error {
	var errs error

	testFuncs := &testFuncs{
		PackageName: pkgName,
		Verbose:     verbose,
		RunFlag:     runFlag,
	}
	loadTestFuncs(pkgName, testFuncs, files)

	testmain, err := formatTestmain(testFuncs)
	if err != nil {
		log.Fatal(err)
	}

	m.RunFiles(files.Files...)
	n := gno.MustParseFile("testmain.go", testmain)
	m.RunFiles(n)

	for _, test := range testFuncs.Tests {
		if verbose {
			fmt.Printf("=== RUN   %s\n", test.Name)
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
			fmt.Printf("--- FAIL: %s (%v)\n", test.Name, duration)
			continue
		}

		// TODO: replace with amino or send native type?
		var rep report
		err = json.Unmarshal([]byte(ret), &rep)
		if err != nil {
			errs = multierr.Append(errs, err)
			fmt.Printf("--- FAIL: %s (%s)\n", test.Name, dstr)
			continue
		}

		switch {
		case rep.Filtered:
			fmt.Printf("--- FILT: %s\n", test.Name)
			// noop
		case rep.Skipped:
			if verbose {
				fmt.Printf("--- SKIP: %s\n", test.Name)
			}
		case rep.Failed:
			err := errors.New("failed: %q", test.Name)
			errs = multierr.Append(errs, err)
			fmt.Printf("--- FAIL: %s (%s)\n", test.Name, dstr)
		default:
			if verbose {
				fmt.Printf("--- PASS: %s (%s)\n", test.Name, dstr)
			}
		}

		if rep.Output != "" && (verbose || rep.Failed) {
			fmt.Printf("output: %s\n", rep.Output)
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
