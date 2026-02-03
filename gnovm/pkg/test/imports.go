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
	gnostdlibs "github.com/gnolang/gno/gnovm/stdlibs"
	teststdlibs "github.com/gnolang/gno/gnovm/tests/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

type StoreOptions struct {
	// WithExamples if true includes the examples/ folder in the gno project.
	WithExamples bool

	// Testing if true includes tests/stdlibs. If false, WithExtern omitted.
	Testing bool

	// WithExtern lets imports of packages "filetests/extern/"
	// as imports under the directory in "gnovm/tests/files/extern".
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

	// Preloaded packages
	Packages packages.PkgList

	// SourceStore, if given, is used to process imports, whenever a custom
	// version doesn't exist in the testing standard libraries.
	// This ignores the value of WithExtern.
	SourceStore gno.Store
}

// This store without options supports stdlibs without test/stdlibs overrides.
// It is used for type-checking gno files without any test files.
// NOTE: It's called "Prod*" because it's suitable for type-checking non-test
// (production) files of a mem package, but it shouldn't be used for production
// systems.
func ProdStore(
	rootDir string,
	output io.Writer,
	pkgs packages.PkgList,
) (
	baseStore storetypes.CommitStore,
	gnoStore gno.Store,
) {
	return StoreWithOptions(
		rootDir,
		output,
		StoreOptions{
			WithExamples: true,
			Testing:      false,
			Packages:     pkgs,
		},
	)
}

// This store without options supports stdlibs with test/stdlibs overrides.  It
// is used for type-checking normal + non-xxx_test *_test.gno files, as well as
// xxx_test integrataion files, and filetest files.
func TestStore(
	rootDir string,
	output io.Writer,
	pkgs packages.PkgList,
) (
	baseStore storetypes.CommitStore,
	gnoStore gno.Store,
) {
	return StoreWithOptions(
		rootDir,
		output,
		StoreOptions{
			WithExamples: true,
			Testing:      true,
			Packages:     pkgs,
		},
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
	gnoStore gno.Store,
) {
	//----------------------------------------
	// process the non-stdlib mempackage after gno.MustReadMemPackage().
	// * m.PreprocessFiles() if opts.PreprocessOnly.
	// * m.RunMemPackage() otherwise.
	_processMemPackage := func(
		m *gno.Machine, mpkg *std.MemPackage, save bool) (
		pn *gno.PackageNode, pv *gno.PackageValue,
	) {
		// _processMemPackage should only be called for "prod" packages.
		// filetests/extern are MPStdlibProd, and examples are MPUserProd.
		// (pkg/test/test.go Test() will filter for MPFTest and store
		// the MP*Test mpkg in the store before running tests, so
		// MPUserProd is all we need here.)
		mptype := mpkg.Type.(gno.MemPackageType)
		if !mptype.IsProd() {
			// For non-prod packages (like test packages during linting),
			// skip processing and return nil
			return nil, nil
		}
		if opts.PreprocessOnly {
			// Check the gno.mod gno version.
			mod, err := gno.ParseCheckGnoMod(mpkg)
			if err != nil {
				panic(fmt.Errorf("test store parsing gno.mod: %w", err))
			}
			if mod == nil || mod.GetGno() == gno.GnoVerMissing {
				panic(fmt.Errorf("cannot parse %q: transpile to %s first", mpkg.Path, gno.GnoVerLatest))
			}
			m.Store.AddMemPackage(mpkg, mptype)
			return m.PreprocessFiles(
				mpkg.Name, mpkg.Path,
				m.ParseMemPackageAsType(mpkg, mptype),
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
			// if _test package... pretend stdlib.
			const testPath = "filetests/extern/"
			if strings.HasPrefix(pkgPath, testPath) {
				baseDir := filepath.Join(rootDir, "gnovm", "tests", "files", "extern", pkgPath[len(testPath):])
				mpkg := gno.MustReadMemPackage(baseDir, pkgPath, gno.MPStdlibProd)
				send := std.Coins{}
				ctx := Context("", pkgPath, send)
				m2 := gno.NewMachineWithOptions(gno.MachineOptions{
					PkgPath:       pkgPath,
					Output:        output,
					Store:         store,
					Context:       ctx,
					ReviveEnabled: true,
					SkipPackage:   true,
				})
				return _processMemPackage(m2, mpkg, true)
			}
		}

		// Load normal stdlib.
		if opts.SourceStore != nil {
			// Only perform actual loading if there exists a testing stdlib.
			if gno.IsStdlib(pkgPath) {
				loc := testStdlibLocation(rootDir, pkgPath)
				if osm.DirExists(loc) {
					pn, pv = loadStdlib(rootDir, pkgPath, store, output, opts.PreprocessOnly, opts.Testing)
					if pn != nil {
						return
					}
				}
			}
			// Get the package from the source store.
			pv = opts.SourceStore.GetPackage(pkgPath, true)
			if pv != nil {
				pn = pv.GetPackageNode(opts.SourceStore)
				mp := opts.SourceStore.GetMemPackage(pkgPath)
				if mp != nil {
					store.AddMemPackage(mp, mp.Type.(gno.MemPackageType))
				}
			} else {
				pn = nil
			}
			return
		}
		if gno.IsStdlib(pkgPath) {
			pn, pv = loadStdlib(rootDir, pkgPath, store, output, opts.PreprocessOnly, opts.Testing)
			if pn != nil {
				return
			}
		}

		loadFromDir := func(dir string) (pn *gno.PackageNode, pv *gno.PackageValue) {
			mpkg := gno.MustReadMemPackage(dir, pkgPath, gno.MPUserProd)
			if mpkg.IsEmpty() {
				panic(fmt.Sprintf("found an empty package %q", pkgPath))
			}
			send := std.Coins{}
			ctx := Context("", pkgPath, send)
			m2 := gno.NewMachineWithOptions(gno.MachineOptions{
				PkgPath:       pkgPath,
				Output:        output,
				Store:         store,
				Context:       ctx,
				ReviveEnabled: true,
				SkipPackage:   true,
			})
			return _processMemPackage(m2, mpkg, true)
		}

		// If available in loaded packages
		if pkg := opts.Packages.Get(pkgPath); pkg != nil {
			return loadFromDir(pkg.Dir)
		}

		if opts.WithExamples {
			// if examples package...
			examplePath := filepath.Join(rootDir, "examples", pkgPath)
			if osm.DirExists(examplePath) {
				return loadFromDir(examplePath)
			}
		}

		return nil, nil
	}

	//----------------------------------------
	// Construct new stores
	db := memdb.NewMemDB()
	baseStore = dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	// Make a new gno store.
	gnoStore = gno.NewStore(nil, baseStore, baseStore)
	gnoStore.SetPackageGetter(getPackage)
	if opts.Testing {
		gnoStore.SetNativeResolver(teststdlibs.NativeResolver)
	} else {
		gnoStore.SetNativeResolver(gnostdlibs.NativeResolver)
	}
	return
}

func stdlibLocation(rootDir, pkgPath string) string {
	return filepath.Join(rootDir, "gnovm", "stdlibs", pkgPath)
}

func testStdlibLocation(rootDir, pkgPath string) string {
	return filepath.Join(rootDir, "gnovm", "tests", "stdlibs", pkgPath)
}

// if !testing, result must be safe for production type-checking.
func loadStdlib(
	rootDir, pkgPath string,
	store gno.Store,
	stdout io.Writer,
	preprocessOnly bool,
	testing bool,
) (*gno.PackageNode, *gno.PackageValue) {
	dirs := []string{
		// Normal stdlib path.
		stdlibLocation(rootDir, pkgPath),
	}
	mPkgType := gno.MPStdlibProd
	if testing {
		// Override path. Definitions here override the previous if duplicate.
		dirs = append(dirs, testStdlibLocation(rootDir, pkgPath))
		mPkgType = gno.MPStdlibTest
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

	mpkg := gno.MustReadMemPackageFromList(files, pkgPath, mPkgType)
	m2 := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath:       pkgPath,
		Output:        stdout,
		Store:         store,
		ReviveEnabled: true,
		SkipPackage:   true, // will PreprocessFiles() or RunMemPackage() after.
	})
	if preprocessOnly {
		m2.Store.AddMemPackage(mpkg, mPkgType)
		return m2.PreprocessFiles(mpkg.Name, mpkg.Path, m2.ParseMemPackageAsType(mpkg, mPkgType), true, true, "")
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
	imports := importsMap.Merge(
		packages.FileKindPackageSource,
		packages.FileKindTest,
		packages.FileKindXTest,
	)
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
		// Get package from store, recursively as necessary.
		pkg := store.GetPackage(imp.PkgPath, true)
		if abortOnError && pkg == nil {
			return gno.ImportNotFoundError{Location: fset.Position(imp.Spec.Pos()).String(), PkgPath: imp.PkgPath}
		}
	}
	return nil
}
