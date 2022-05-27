package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/gnolang/gno"
	dbm "github.com/gnolang/gno/pkgs/db"
	"github.com/gnolang/gno/pkgs/errors"
	osm "github.com/gnolang/gno/pkgs/os"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/gno/pkgs/store/dbadapter"
	"github.com/gnolang/gno/pkgs/store/iavl"
	stypes "github.com/gnolang/gno/pkgs/store/types"
	"github.com/gnolang/gno/stdlibs"
)

type testFuncs struct {
	Tests       []testFunc
	PackageName string
	Verbose     bool
}

type testFunc struct {
	Package string
	Name    string
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

func testrun() (ok bool) {
 	return testing.RunTests({{.Verbose}},tests)
}

`))

func runTest(testStore gno.Store, pkgPath string, verbose bool) (ok bool) {
	memPkg := gno.ReadMemPackage(pkgPath, pkgPath)
	m := gno.NewMachine(memPkg.Name, testStore)
	m.RunMemPackage(memPkg, true)

	//tfiles, ifiles := gno.ParseMemPackageTests(memPkg)
	tfiles, ifiles := parseMemPackageTests(memPkg)

	// run test files in pkg
	testok := runTestFiles(testStore, m, tfiles, memPkg.Name, verbose)

	// run test files in xxx_test pkg
	testPkgName := getPkgNameFromFileset(ifiles)
	if testPkgName == "" { // empty test funcs
		return testok
	}

	m2 := gno.NewMachine(testPkgName, testStore)
	itestok := runTestFiles(testStore, m2, ifiles, testPkgName, verbose)

	if testok && itestok {
		return true
	}
	return false
}

func runTestFiles(testStore gno.Store, m *gno.Machine, files *gno.FileSet, pkgName string, verbose bool) bool {
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

	res := m.Eval(gno.Call("testrun"))[0].GetBool()
	return res
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

func newTestStore(rootDir string, stdin io.Reader, stdout, stderr io.Writer) gno.Store {
	db := dbm.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
	store := gno.NewStore(nil, baseStore, iavlStore)

	getPackage := func(pkgPath string) (pn *gno.PackageNode, pv *gno.PackageValue) {
		switch pkgPath {
		case "os":
			pkg := gno.NewPackageNode("os", pkgPath, nil)
			pkg.DefineGoNativeValue("Stdin", stdin)
			pkg.DefineGoNativeValue("Stdout", stdout)
			pkg.DefineGoNativeValue("Stderr", stderr)
			return pkg, pkg.NewPackage()
		case "fmt":
			pkg := gno.NewPackageNode("fmt", pkgPath, nil)
			pkg.DefineGoNativeType(reflect.TypeOf((*fmt.Stringer)(nil)).Elem())
			pkg.DefineGoNativeType(reflect.TypeOf((*fmt.Formatter)(nil)).Elem())
			pkg.DefineGoNativeValue("Println", func(a ...interface{}) (n int, err error) {
				// NOTE: uncomment to debug long running tests
				res := fmt.Sprintln(a...)
				return stdout.Write([]byte(res))
			})
			pkg.DefineGoNativeValue("Printf", func(format string, a ...interface{}) (n int, err error) {
				res := fmt.Sprintf(format, a...)
				return stdout.Write([]byte(res))
			})
			pkg.DefineGoNativeValue("Print", func(a ...interface{}) (n int, err error) {
				res := fmt.Sprint(a...)
				return stdout.Write([]byte(res))
			})
			pkg.DefineGoNativeValue("Sprint", fmt.Sprint)
			pkg.DefineGoNativeValue("Sprintf", fmt.Sprintf)
			pkg.DefineGoNativeValue("Sprintln", fmt.Sprintln)
			pkg.DefineGoNativeValue("Sscanf", fmt.Sscanf)
			pkg.DefineGoNativeValue("Errorf", fmt.Errorf)
			pkg.DefineGoNativeValue("Fprintln", fmt.Fprintln)
			pkg.DefineGoNativeValue("Fprintf", fmt.Fprintf)
			pkg.DefineGoNativeValue("Fprint", fmt.Fprint)

			return pkg, pkg.NewPackage()
		}

		// stdlib
		stdlibPath := filepath.Join(rootDir, "stdlibs", pkgPath)
		if osm.DirExists(stdlibPath) {
			memPkg := gno.ReadMemPackage(stdlibPath, pkgPath)
			m := gno.NewMachineWithOptions(gno.MachineOptions{
				PkgPath: "test",
				Output:  stdout,
				Store:   store,
			})
			return m.RunMemPackage(memPkg, false)
		}
		// examples
		examplePath := filepath.Join(rootDir, "examples", pkgPath)
		if osm.DirExists(examplePath) {
			memPkg := gno.ReadMemPackage(examplePath, pkgPath)
			m := gno.NewMachineWithOptions(gno.MachineOptions{
				PkgPath: "test",
				Output:  stdout,
				Store:   store,
			})
			return m.RunMemPackage(memPkg, false)
		}
		return nil, nil
	}

	store.SetPackageGetter(getPackage)
	store.SetStrictGo2GnoMapping(false)
	store.SetPackageInjector(testPackageInjector)
	// native mappings
	stdlibs.InjectNativeMappings(store)

	return store
}

func testPackageInjector(store gno.Store, pn *gno.PackageNode) {
	stdlibs.InjectPackage(store, pn)
	switch pn.PkgPath {
	case "strconv":
		// NOTE: Itoa and Atoi are already injected
		// from stdlibs.InjectNatives.
		pn.DefineGoNativeType(reflect.TypeOf(strconv.NumError{}))
		pn.DefineGoNativeValue("ParseInt", strconv.ParseInt)
	}
}
