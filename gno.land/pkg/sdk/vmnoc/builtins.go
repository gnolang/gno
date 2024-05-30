package vm

import (
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func (vm *VMKeeper) initBuiltinPackagesAndTypes(store gno.Store) {
	getPackage := func(pkgPath string, newStore gno.Store) (pn *gno.PackageNode, pv *gno.PackageValue) {
		stdlibPath := filepath.Join(vm.stdlibsDir, pkgPath)
		checkStdlibs := !osm.DirExists(stdlibPath)
		if checkStdlibs {
			newPath := filepath.Join("stdlibs", pkgPath)
			if _, err := gnovm.StdLibsFS.ReadDir(newPath); err != nil {
				// no gno files are present, skip this package
				return nil, nil
			}

			stdlibPath = newPath
		}

		var memPkg *std.MemPackage
		if checkStdlibs {
			memPkg = gno.ReadMemPackageFS(gnovm.StdLibsFS, stdlibPath, pkgPath)
		} else {
			memPkg = gno.ReadMemPackage(stdlibPath, pkgPath)
		}

		if memPkg.IsEmpty() {
			// no gno files are present, skip this package
			return nil, nil
		}

		m2 := gno.NewMachineWithOptions(gno.MachineOptions{
			PkgPath: "gno.land/r/stdlibs/" + pkgPath,
			Output:  os.Stdout,
			Store:   newStore,
		})
		defer m2.Release()
		return m2.RunMemPackage(memPkg, true)
	}
	store.SetPackageGetter(getPackage)
	store.SetNativeStore(stdlibs.NativeStore)
}
