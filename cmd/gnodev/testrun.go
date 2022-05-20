package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/gnolang/gno"
	dbm "github.com/gnolang/gno/pkgs/db"
	osm "github.com/gnolang/gno/pkgs/os"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/gno/pkgs/store/dbadapter"
	"github.com/gnolang/gno/pkgs/store/iavl"
	stypes "github.com/gnolang/gno/pkgs/store/types"
)

type testFuncs struct {
	Tests   []testFunc
	Package *std.MemPackage
	Verbose bool
}

type testFunc struct {
	Package string
	Name    string
}

var testmainTmpl = template.Must(template.New("testmain").Parse(`
package {{ .Package.Name }} 

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
	tfiles, _ := gno.ParseMemPackageTests(memPkg)
	testFuncs := loadTestFuncs(memPkg, tfiles)
	testFuncs.Verbose = verbose
	testmain, err := formatTestmain(testFuncs)
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Printf("testmain: %s\n", testmain)

	m := gno.NewMachine(memPkg.Name, testStore)
	m.RunMemPackage(memPkg, false)
	m.RunFiles(tfiles.Files...)

	n := gno.MustParseFile("testmain.go", testmain)
	m.RunFiles(n)

	res := m.Eval(gno.Call("testrun"))[0].GetBool()
	return res
}

func formatTestmain(t *testFuncs) (string, error) {
	var buf bytes.Buffer
	if err := testmainTmpl.Execute(&buf, t); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func loadTestFuncs(memPkg *std.MemPackage, tfiles *gno.FileSet) *testFuncs {
	t := &testFuncs{
		Package: memPkg,
	}

	for _, tf := range tfiles.Files {
		for _, d := range tf.Decls {
			if fd, ok := d.(*gno.FuncDecl); ok {
				fname := string(fd.Name)
				if strings.HasPrefix(fname, "Test") {
					tf := testFunc{
						Package: memPkg.Name,
						Name:    fname,
					}
					t.Tests = append(t.Tests, tf)
				}
			}
		}
	}
	return t
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
	/*
			store.SetStrictGo2GnoMapping(false)
		    // native mappings
		    stdlibs.InjectNativeMappings(store)
	*/

	return store
}
