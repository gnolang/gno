package test

import (
	"errors"
	"fmt"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/gnolang/gno/gnovm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/packages"
	teststdlibs "github.com/gnolang/gno/gnovm/tests/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

type StoreOptions struct {
	// WithExtern interprets imports of packages under "github.com/gnolang/gno/_test/"
	// as imports under the directory in gnovm/tests/files/extern.
	// This should only be used for GnoVM internal filetests (gnovm/tests/files).
	WithExtern bool

	// PreprocessOnly instructs the PackageGetter to run the imported files using
	// [gno.Machine.PreprocessFiles]. It avoids executing code for contexts
	// which only intend to perform a type check, ie. `gno lint`.
	PreprocessOnly bool
}

// NOTE: this isn't safe, should only be used for testing.
func Store(
	rootDir string,
	output io.Writer,
) (
	baseStore storetypes.CommitStore,
	resStore gno.Store,
) {
	return StoreWithOptions(rootDir, output, StoreOptions{})
}

// StoreWithOptions is a variant of [Store] which additionally accepts a
// [StoreOptions] argument.
func StoreWithOptions(
	rootDir string,
	output io.Writer,
	opts StoreOptions,
) (
	baseStore storetypes.CommitStore,
	resStore gno.Store,
) {
	processMemPackage := func(m *gno.Machine, memPkg *gnovm.MemPackage, save bool) (*gno.PackageNode, *gno.PackageValue) {
		return m.RunMemPackage(memPkg, save)
	}
	if opts.PreprocessOnly {
		processMemPackage = func(m *gno.Machine, memPkg *gnovm.MemPackage, save bool) (*gno.PackageNode, *gno.PackageValue) {
			m.Store.AddMemPackage(memPkg)
			return m.PreprocessFiles(memPkg.Name, memPkg.Path, gno.ParseMemPackage(memPkg), save, false)
		}
	}
	getPackage := func(pkgPath string, store gno.Store) (pn *gno.PackageNode, pv *gno.PackageValue) {
		if pkgPath == "" {
			panic(fmt.Sprintf("invalid zero package path in testStore().pkgGetter"))
		}

		if opts.WithExtern {
			// if _test package...
			const testPath = "github.com/gnolang/gno/_test/"
			if strings.HasPrefix(pkgPath, testPath) {
				baseDir := filepath.Join(rootDir, "gnovm", "tests", "files", "extern", pkgPath[len(testPath):])
				memPkg := gno.MustReadMemPackage(baseDir, pkgPath)
				send := std.Coins{}
				ctx := Context("", pkgPath, send)
				m2 := gno.NewMachineWithOptions(gno.MachineOptions{
					PkgPath:       "test",
					Output:        output,
					Store:         store,
					Context:       ctx,
					ReviveEnabled: true,
				})
				return processMemPackage(m2, memPkg, true)
			}
		}

		// Load normal stdlib.
		pn, pv = loadStdlib(rootDir, pkgPath, store, output, opts.PreprocessOnly)
		if pn != nil {
			return
		}

		// if examples package...
		examplePath := filepath.Join(rootDir, "examples", pkgPath)
		if osm.DirExists(examplePath) {
			memPkg := gno.MustReadMemPackage(examplePath, pkgPath)
			if memPkg.IsEmpty() {
				panic(fmt.Sprintf("found an empty package %q", pkgPath))
			}

			send := std.Coins{}
			ctx := Context("", pkgPath, send)
			m2 := gno.NewMachineWithOptions(gno.MachineOptions{
				PkgPath:       "test",
				Output:        output,
				Store:         store,
				Context:       ctx,
				ReviveEnabled: true,
			})
			return processMemPackage(m2, memPkg, true)
		}
		return nil, nil
	}
	db := memdb.NewMemDB()
	baseStore = dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	// Make a new store.
	resStore = gno.NewStore(nil, baseStore, baseStore)
	resStore.SetPackageGetter(getPackage)
	resStore.SetNativeResolver(teststdlibs.NativeResolver)
	return
}

func loadStdlib(rootDir, pkgPath string, store gno.Store, stdout io.Writer, preprocessOnly bool) (*gno.PackageNode, *gno.PackageValue) {
	dirs := [...]string{
		// Normal stdlib path.
		filepath.Join(rootDir, "gnovm", "stdlibs", pkgPath),
		// Override path. Definitions here override the previous if duplicate.
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

	memPkg := gno.MustReadMemPackageFromList(files, pkgPath)
	m2 := gno.NewMachineWithOptions(gno.MachineOptions{
		// NOTE: see also pkgs/sdk/vm/builtins.go
		// Needs PkgPath != its name because TestStore.getPackage is the package
		// getter for the store, which calls loadStdlib, so it would be recursively called.
		PkgPath:       "stdlibload",
		Output:        stdout,
		Store:         store,
		ReviveEnabled: true,
	})
	if preprocessOnly {
		m2.Store.AddMemPackage(memPkg)
		return m2.PreprocessFiles(memPkg.Name, memPkg.Path, gno.ParseMemPackage(memPkg), true, true)
	}
	// TODO: make this work when using gno lint.
	return m2.RunMemPackageWithOverrides(memPkg, true)
}

type stackWrappedError struct {
	err   error
	stack []byte
}

func (e *stackWrappedError) Error() string { return e.err.Error() }
func (e *stackWrappedError) Unwrap() error { return e.err }
func (e *stackWrappedError) String() string {
	return fmt.Sprintf("%v\nstack:\n%v", e.err, string(e.stack))
}

// LoadImports parses the given MemPackage and attempts to retrieve all pure packages
// from the store. This is mostly useful for "eager import loading", whereby all
// imports are pre-loaded in a permanent store, so that the tests can use
// ephemeral transaction stores.
func LoadImports(store gno.Store, memPkg *gnovm.MemPackage) (err error) {
	defer func() {
		// This is slightly different from other similar error handling; we do not have a
		// machine to work with, as this comes from an import; so we need
		// "machine-less" alternatives. (like v.String instead of v.Sprint)
		if r := recover(); r != nil {
			switch v := r.(type) {
			case *gno.TypedValue:
				err = errors.New(v.String())
			case *gno.PreprocessError:
				err = &stackWrappedError{v.Unwrap(), debug.Stack()}
			case gno.UnhandledPanicError:
				err = v
			case error:
				err = &stackWrappedError{v, debug.Stack()}
			default:
				err = &stackWrappedError{fmt.Errorf("%v", v), debug.Stack()}
			}
		}
	}()

	fset := token.NewFileSet()
	importsMap, err := packages.Imports(memPkg, fset)
	if err != nil {
		return err
	}
	imports := importsMap.Merge(packages.FileKindPackageSource, packages.FileKindTest, packages.FileKindXTest)
	for _, imp := range imports {
		if gno.IsRealmPath(imp.PkgPath) {
			// Don't eagerly load realms.
			// Realms persist state and can change the state of other realms in initialization.
			continue
		}
		pkg := store.GetPackage(imp.PkgPath, true)
		if pkg == nil {
			return fmt.Errorf("%v: unknown import path %v", fset.Position(imp.Spec.Pos()).String(), imp.PkgPath)
		}
	}
	return nil
}
