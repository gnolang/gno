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

	// When transpiling code in examples/ we use the test store. gno fix may need
	// gno.mod to not be auto-generated when importing from the test store.
	DoNotGenerateGnoMod bool

	// When fixing code from an earler gno version. Not supported for stdlibs.
	FixFrom string
}

// NOTE: this isn't safe, should only be used for testing.
func Store(
	rootDir string,
	output io.Writer,
) (
	baseStore storetypes.CommitStore,
	resStore gno.Store,
) {
	return StoreWithOptions(
		rootDir,
		output,
		StoreOptions{},
	)
}

// ========================================
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
	//----------------------------------------
	// process the mempackage after gno.MustReadMemPackage().
	// * m.PreprocessFiles() if opts.PreprocessOnly.
	// * m.RunMemPackage() otherwise.
	_processMemPackage := func(
		m *gno.Machine, mpkg *std.MemPackage, save bool) (
		pn *gno.PackageNode, pv *gno.PackageValue,
	) {
		if opts.PreprocessOnly {
			// Check the gno.mod gno version.
			mod, err := gno.ParseCheckGnoMod(mpkg)
			if err != nil {
				panic(fmt.Errorf("test store parsing gno.mod: %w", err))
			}
			if mod == nil || mod.GetGno() == gno.GnoVerMissing {
				// In order to translate into a newer Gno version with
				// the preprocessor make a slight modifications to the
				// AST. This needs to happen even for imports, because
				// the preprocessor requires imports also preprocessed.
				// This is because the linter uses pkg/test/imports.go.
				gofset, _, gofs, _gofs, tgofs, errs := gno.GoParseMemPackage(
					mpkg, gno.ParseModeAll)
				if errs != nil {
					panic(fmt.Errorf("test store parsing: %w", errs))
				}
				allgofs := append(gofs, _gofs...)
				allgofs = append(allgofs, tgofs...)
				errs = gno.PrepareGno0p9(gofset, allgofs, mpkg)
				if errs != nil {
					panic(fmt.Errorf("test store preparing AST: %w", errs))
				}
			}
			m.Store.AddMemPackage(mpkg, gno.MemPackageTypeAny)
			return m.PreprocessFiles(
				mpkg.Name, mpkg.Path,
				gno.ParseMemPackage(mpkg),
				save, false, opts.FixFrom)
		} else {
			return m.RunMemPackage(mpkg, save)
		}
	}

	//----------------------------------------
	// Main entrypoint for new test imports.
	getPackage := func(pkgPath string, store gno.Store) (pn *gno.PackageNode, pv *gno.PackageValue) {
		if pkgPath == "" {
			panic("invalid zero package path in testStore().pkgGetter")
		}
		if opts.WithExtern {
			// if _test package...
			const testPath = "github.com/gnolang/gno/_test/"
			if strings.HasPrefix(pkgPath, testPath) {
				baseDir := filepath.Join(rootDir, "gnovm", "tests", "files", "extern", pkgPath[len(testPath):])
				mpkg := gno.MustReadMemPackage(baseDir, pkgPath)
				send := std.Coins{}
				ctx := Context("", pkgPath, send)
				m2 := gno.NewMachineWithOptions(gno.MachineOptions{
					PkgPath:       "test",
					Output:        output,
					Store:         store,
					Context:       ctx,
					ReviveEnabled: true,
				})
				return _processMemPackage(m2, mpkg, true)
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
			mpkg := gno.MustReadMemPackage(examplePath, pkgPath)
			if mpkg.IsEmpty() {
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
			return _processMemPackage(m2, mpkg, true)
		}

		return nil, nil
	}

	//----------------------------------------
	// Construct new stores
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

	mpkg := gno.MustReadMemPackageFromList(files, pkgPath, gno.MemPackageTypeStdlib)
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
		m2.Store.AddMemPackage(mpkg, gno.MemPackageTypeStdlib)
		return m2.PreprocessFiles(mpkg.Name, mpkg.Path, gno.ParseMemPackage(mpkg), true, true, "")
	}
	// TODO: make this work when using gno lint.
	return m2.RunMemPackageWithOverrides(mpkg, true)
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
func LoadImports(store gno.Store, mpkg *std.MemPackage, abortOnError bool) (err error) {
	// If this gets out of hand (e.g. with nested catchPanic with need for
	// selective catching) then pass in a bool instead.
	// See also cmd/gno/common.go.
	if os.Getenv("DEBUG_PANIC") == "1" {
		fmt.Println("DEBUG_PANIC=1 (will not recover)")
	} else {
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
	}

	fset := token.NewFileSet()
	importsMap, err := packages.Imports(mpkg, fset)
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
		if !abortOnError {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("ignoring panic: %v\n", r)
				}
			}()
		}
		pkg := store.GetPackage(imp.PkgPath, true)
		if abortOnError && pkg == nil {
			return gno.ImportNotFoundError{Location: fset.Position(imp.Spec.Pos()).String(), PkgPath: imp.PkgPath}
		}
	}
	return nil
}
