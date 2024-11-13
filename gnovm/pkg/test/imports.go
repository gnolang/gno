package test

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	teststdlibs "github.com/gnolang/gno/gnovm/tests/stdlibs"
	teststd "github.com/gnolang/gno/gnovm/tests/stdlibs/std"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// NOTE: this isn't safe, should only be used for testing.
func TestStore(
	rootDir string,
	withExtern bool,
	stdin io.Reader,
	stdout, stderr io.Writer,
) (
	baseStore storetypes.CommitStore,
	resStore gno.Store,
) {
	getPackage := func(pkgPath string, store gno.Store) (pn *gno.PackageNode, pv *gno.PackageValue) {
		if pkgPath == "" {
			panic(fmt.Sprintf("invalid zero package path in testStore().pkgGetter"))
		}

		if withExtern {
			// if _test package...
			const testPath = "github.com/gnolang/gno/_test/"
			if strings.HasPrefix(pkgPath, testPath) {
				baseDir := filepath.Join(rootDir, "gnovm", "tests", "files", "extern", pkgPath[len(testPath):])
				memPkg := gno.ReadMemPackage(baseDir, pkgPath)
				send := std.Coins{}
				ctx := TestContext(pkgPath, send)
				m2 := gno.NewMachineWithOptions(gno.MachineOptions{
					PkgPath: "test",
					Output:  stdout,
					Store:   store,
					Context: ctx,
				})
				return m2.RunMemPackage(memPkg, true)
			}
		}

		// gonative exceptions.
		switch pkgPath {
		case "os":
			pkg := gno.NewPackageNode("os", pkgPath, nil)
			pkg.DefineGoNativeValue("Stdin", stdin)
			pkg.DefineGoNativeValue("Stdout", stdout)
			pkg.DefineGoNativeValue("Stderr", stderr)
			return pkg, pkg.NewPackage()
		case "fmt":
			pkg := gno.NewPackageNode("fmt", pkgPath, nil)
			pkg.DefineGoNativeValue("Println", func(a ...interface{}) (n int, err error) {
				// NOTE: uncomment to debug long running tests
				// fmt.Println(a...)
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
		case "encoding/json":
			pkg := gno.NewPackageNode("json", pkgPath, nil)
			pkg.DefineGoNativeValue("Unmarshal", json.Unmarshal)
			pkg.DefineGoNativeValue("Marshal", json.Marshal)
			return pkg, pkg.NewPackage()
		case "encoding/xml":
			pkg := gno.NewPackageNode("xml", pkgPath, nil)
			pkg.DefineGoNativeValue("Unmarshal", xml.Unmarshal)
			return pkg, pkg.NewPackage()
		case "internal/os_test":
			pkg := gno.NewPackageNode("os_test", pkgPath, nil)
			pkg.DefineNative("Sleep",
				gno.Flds( // params
					"d", gno.AnyT(), // NOTE: should be time.Duration
				),
				gno.Flds( // results
				),
				func(m *gno.Machine) {
					// For testing purposes here, nanoseconds are separately kept track.
					arg0 := m.LastBlock().GetParams1().TV
					d := arg0.GetInt64()
					sec := d / int64(time.Second)
					nano := d % int64(time.Second)
					ctx := m.Context.(*teststd.TestExecContext)
					ctx.Timestamp += sec
					ctx.TimestampNano += nano
					if ctx.TimestampNano >= int64(time.Second) {
						ctx.Timestamp += 1
						ctx.TimestampNano -= int64(time.Second)
					}
					m.Context = ctx
				},
			)
			return pkg, pkg.NewPackage()
		case "math/big":
			pkg := gno.NewPackageNode("big", pkgPath, nil)
			pkg.DefineGoNativeValue("NewInt", big.NewInt)
			return pkg, pkg.NewPackage()
		case "context":
			pkg := gno.NewPackageNode("context", pkgPath, nil)
			pkg.DefineGoNativeValue("WithValue", context.WithValue)
			pkg.DefineGoNativeValue("Background", context.Background)
			return pkg, pkg.NewPackage()
		}

		// load normal stdlib.
		pn, pv = loadStdlib(rootDir, pkgPath, store, stdout)
		if pn != nil {
			return
		}

		// if examples package...
		examplePath := filepath.Join(rootDir, "examples", pkgPath)
		if osm.DirExists(examplePath) {
			memPkg := gno.ReadMemPackage(examplePath, pkgPath)
			if memPkg.IsEmpty() {
				panic(fmt.Sprintf("found an empty package %q", pkgPath))
			}

			send := std.Coins{}
			ctx := TestContext(pkgPath, send)
			m2 := gno.NewMachineWithOptions(gno.MachineOptions{
				PkgPath: "test",
				Output:  stdout,
				Store:   store,
				Context: ctx,
			})
			pn, pv = m2.RunMemPackage(memPkg, true)
			return
		}
		return nil, nil
	}
	db := memdb.NewMemDB()
	baseStore = dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	// make a new store
	resStore = gno.NewStore(nil, baseStore, baseStore)
	resStore.SetPackageGetter(getPackage)
	resStore.SetNativeStore(teststdlibs.NativeStore)
	resStore.SetStrictGo2GnoMapping(false)
	return
}

func loadStdlib(rootDir, pkgPath string, store gno.Store, stdout io.Writer) (*gno.PackageNode, *gno.PackageValue) {
	dirs := [...]string{
		// normal stdlib path.
		filepath.Join(rootDir, "gnovm", "stdlibs", pkgPath),
		// override path. definitions here override the previous if duplicate.
		filepath.Join(rootDir, "gnovm", "tests", "stdlibs", pkgPath),
	}
	files := make([]string, 0, 32) // pre-alloc 32 as a likely high number of files
	for _, path := range dirs {
		dl, err := os.ReadDir(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			panic(fmt.Errorf("could not access dir %q: %w", path, err))
		}

		for _, f := range dl {
			// NOTE: RunMemPackage has other rules; those should be mostly useful
			// for on-chain packages (ie. include README and gno.mod).
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".gno") {
				files = append(files, filepath.Join(path, f.Name()))
			}
		}
	}
	if len(files) == 0 {
		return nil, nil
	}

	memPkg := gno.ReadMemPackageFromList(files, pkgPath)
	m2 := gno.NewMachineWithOptions(gno.MachineOptions{
		// NOTE: see also pkgs/sdk/vm/builtins.go
		// Needs PkgPath != its name because TestStore.getPackage is the package
		// getter for the store, which calls loadStdlib, so it would be recursively called.
		PkgPath: "stdlibload",
		Output:  stdout,
		Store:   store,
	})
	save := pkgPath != "testing" // never save the "testing" package
	return m2.RunMemPackageWithOverrides(memPkg, save)
}
