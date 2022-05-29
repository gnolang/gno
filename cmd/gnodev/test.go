package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/gnolang/gno"
	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/gno/tests"
	"go.uber.org/multierr"
)

type testOptions struct {
	Verbose bool   `flag:"verbose" help:"verbose"`
	RootDir string `flag:"root-dir" help:"clone location of github.com/gnolang/gno (gnodev tries to guess it)"`
	// Run string `flag:"run" help:"test name filtering pattern"`
	// Timeout time.Duration `flag:"timeout" help:"max execution time"`
	// VM Options
	// A flag about if we should download the production realms
	// UseNativeLibs bool // experimental, but could be useful for advanced developer needs
}

var DefaultTestOptions = testOptions{
	Verbose: false,
	RootDir: "",
}

func testApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(testOptions)
	if len(args) < 1 {
		cmd.ErrPrintfln("Usage: test [test flags] [packages]")
		return errors.New("invalid args")
	}

	// guess opts.RootDir
	if opts.RootDir == "" {
		cmd := exec.Command("go", "list", "-m", "-mod=mod", "-f", "{{.Dir}}", "github.com/gnolang/gno")
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal("can't guess --root-dir, please fill it manually.")
		}
		rootDir := strings.TrimSpace(string(out))
		opts.RootDir = rootDir
	}

	pkgPaths, err := gnoPackagesFromArgs(args)
	if err != nil {
		return fmt.Errorf("list packages from args: %w", err)
	}

	errCount := 0
	for _, pkgPath := range pkgPaths {
		unittestFiles, err := filepath.Glob(filepath.Join(pkgPath, "*_test.gno"))
		if err != nil {
			log.Fatal(err)
		}
		filetestFiles, err := filepath.Glob(filepath.Join(pkgPath, "*_filetest.gno"))
		if err != nil {
			log.Fatal(err)
		}
		if len(unittestFiles) == 0 && len(filetestFiles) == 0 {
			cmd.ErrPrintfln("?       %s \t[no test files]", pkgPath)
			continue
		}

		sort.Strings(unittestFiles)
		sort.Strings(filetestFiles)

		startedAt := time.Now()
		err = gnoTestPkg(cmd, pkgPath, unittestFiles, filetestFiles, opts)
		duration := time.Since(startedAt)
		dstr := fmtDuration(duration)

		if err != nil {
			err = fmt.Errorf("%s: test pkg: %w", pkgPath, err)
			cmd.ErrPrintfln("FAIL")
			cmd.ErrPrintfln("FAIL    %s \t%s", pkgPath, dstr)
			cmd.ErrPrintfln("FAIL")
			errCount++
		} else {
			cmd.ErrPrintfln("ok      %s \t%s", pkgPath, dstr)
		}
	}
	if errCount > 0 {
		cmd.ErrPrintfln("FAIL")
		return fmt.Errorf("FAIL: %d go test errors", errCount)
	}

	return nil
}

func gnoTestPkg(cmd *command.Command, pkgPath string, unittestFiles, filetestFiles []string, opts testOptions) error {
	verbose := opts.Verbose
	rootDir := opts.RootDir
	var errs error

	testStore := tests.TestStore(rootDir, "", os.Stdin, os.Stdout, os.Stderr, false)
	if verbose {
		testStore.SetLogStoreOps(true)
	}

	// testing with *_test.gno
	if len(unittestFiles) > 0 {
		stdout := new(bytes.Buffer)
		memPkg := gno.ReadMemPackage(pkgPath, pkgPath)

		//tfiles, ifiles := gno.ParseMemPackageTests(memPkg)
		tfiles, ifiles := parseMemPackageTests(memPkg)

		// run test files in pkg
		{
			m := tests.TestMachine(testStore, stdout, "main")
			m.RunMemPackage(memPkg, true)
			err := runTestFiles(cmd, testStore, m, tfiles, memPkg.Name, verbose)
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
				err := runTestFiles(cmd, testStore, m, ifiles, testPkgName, verbose)
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
			startedAt := time.Now()
			if verbose {
				cmd.ErrPrintfln("=== RUN   %s", testName)
			}

			var closer func() string
			if !verbose {
				var err error
				closer, err = captureStdoutAndStderr()
				if err != nil {
					panic(err)
				}
			}

			testFilePath := filepath.Join(pkgPath, testFileName)
			err := tests.RunFileTest(rootDir, testFilePath, false, nil)
			duration := time.Since(startedAt)
			dstr := fmtDuration(duration)

			if err != nil {
				errs = multierr.Append(errs, err)
				cmd.ErrPrintfln("--- FAIL: %s (%s)", testName, dstr)
				if verbose {
					stdouterr := closer()
					fmt.Fprintln(os.Stderr, stdouterr)
				}
				continue
			}

			if verbose {
				cmd.ErrPrintfln("--- PASS: %s (%s)", testName, dstr)
			}
		}
	}

	return errs
}

func runTestFiles(cmd *command.Command, testStore gno.Store, m *gno.Machine, files *gno.FileSet, pkgName string, verbose bool) error {
	var errs error

	testFuncs := &testFuncs{
		PackageName: pkgName,
		Verbose:     verbose,
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
			cmd.ErrPrintfln("=== RUN   %s", test.Name)
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
			cmd.ErrPrintfln("--- FAIL: %s (%v)", test.Name, duration)
			continue
		}

		// TODO: replace with amino or send native type?
		var rep report
		err = json.Unmarshal([]byte(ret), &rep)
		if err != nil {
			errs = multierr.Append(errs, err)
			cmd.ErrPrintfln("--- FAIL: %s (%s)", test.Name, dstr)
			continue
		}

		switch {
		case rep.Skipped:
			if verbose {
				cmd.ErrPrintfln("--- SKIP: %s", test.Name)
			}
		case rep.Failed:
			cmd.ErrPrintfln("--- FAIL: %s (%s)", test.Name, dstr)
		default:
			if verbose {
				cmd.ErrPrintfln("--- PASS: %s (%s)", test.Name, dstr)
			}
		}

		if rep.Output != "" && (verbose || rep.Failed) {
			cmd.ErrPrintfln("output: %s", rep.Output)
		}
	}

	return errs
}

// mirror of stdlibs/testing.Report
type report struct {
	Name    string
	Verbose bool
	Failed  bool
	Skipped bool
	Output  string
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
			return testing.RunTest({{.Verbose}}, test)
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
