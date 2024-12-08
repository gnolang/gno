package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

const (
	opcodesPkgPath = "gno.land/r/x/benchmark/opcodes"
	rounds         = 1000
)

func benchmarkOpCodes(bstore gno.Store, dir string) {
	opcodesPkgDir := filepath.Join(dir, "opcodes")

	pv := addPackage(bstore, opcodesPkgDir, opcodesPkgPath)
	for i := 0; i < rounds; i++ {
		callOpsBench(bstore, pv)
	}
}

func callOpsBench(bstore gno.Store, pv *gno.PackageValue) {
	// start
	pb := pv.GetBlock(bstore)
	for _, tv := range pb.Values {
		if fv, ok := tv.V.(*gno.FuncValue); ok {
			cx := gno.Call(fv.Name)
			callFunc(bstore, pv, cx)
		}
	}
}

const storagePkgPath = "gno.land/r/x/benchmark/storage"

func benchmarkStorage(bstore BenchStore, dir string) {
	gs := bstore.gnoStore
	avlPkgDir := filepath.Join(dir, "avl")
	addPackage(gs, avlPkgDir, "gno.land/p/demo/avl")

	storagePkgDir := filepath.Join(dir, "storage")
	pv := addPackage(gs, storagePkgDir, storagePkgPath)
	benchStoreSet(bstore, pv)
	benchStoreGet(bstore, pv)
}

func benchStoreSet(bstore BenchStore, pv *gno.PackageValue) {
	title := "1KB content"
	content := strings.Repeat("a", 1024)

	// in forum.gno: func AddPost(title, content string)
	// one AddPost will be added to three different boards in the forum.gno contract

	for i := 0; i < rounds; i++ {
		cx := gno.Call("AddPost", gno.Str(title), gno.Str(content))
		callFunc(bstore.gnoStore, pv, cx)
		bstore.Write()
		bstore.gnoStore.ClearObjectCache()
	}
}

func benchStoreGet(bstore BenchStore, pv *gno.PackageValue) {
	// in forum.gno: func GetPost(boardId, postId int) string  in forum.gno
	// there are three different boards on the benchmarking forum contract
	for i := 0; i < 3; i++ {
		for j := 0; j < rounds; j++ {
			cx := gno.Call("GetPost", gno.X(i), gno.X(j))
			callFunc(bstore.gnoStore, pv, cx)
			bstore.Write()
			bstore.gnoStore.ClearObjectCache()
		}
	}
}

func callFunc(gstore gno.Store, pv *gno.PackageValue, cx gno.Expr) []gno.TypedValue {
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath: pv.PkgPath,
			Output:  os.Stdout, // XXX
			Store:   gstore,
		})

	defer m.Release()

	m.SetActivePackage(pv)
	return m.Eval(cx)
}

// addPacakge

func addPackage(gstore gno.Store, dir string, pkgPath string) *gno.PackageValue {
	// load benchmark contract
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath: "",
			Output:  os.Stdout,
			Store:   gstore,
		})
	defer m.Release()

	memPkg := gno.MustReadMemPackage(dir, pkgPath)

	// pare the file, create pn, pv and save the values in m.store
	_, pv := m.RunMemPackage(memPkg, true)

	return pv
}

// load stdlibs
func loadStdlibs(bstore BenchStore) {
	// copied from vm/builtin.go
	getPackage := func(pkgPath string, newStore gno.Store) (pn *gno.PackageNode, pv *gno.PackageValue) {
		stdlibDir := filepath.Join(gnoenv.RootDir(), "gnovm", "stdlibs")
		stdlibPath := filepath.Join(stdlibDir, pkgPath)
		if !osm.DirExists(stdlibPath) {
			// does not exist.
			return nil, nil
		}

		memPkg := gno.MustReadMemPackage(stdlibPath, pkgPath)
		if memPkg.IsEmpty() {
			// no gno files are present, skip this package
			return nil, nil
		}

		m2 := gno.NewMachineWithOptions(gno.MachineOptions{
			PkgPath: "gno.land/r/stdlibs/" + pkgPath,
			// PkgPath: pkgPath,
			Output: os.Stdout,
			Store:  newStore,
		})
		defer m2.Release()
		return m2.RunMemPackage(memPkg, true)
	}

	bstore.gnoStore.SetPackageGetter(getPackage)
	bstore.gnoStore.SetNativeResolver(stdlibs.NativeResolver)
}
