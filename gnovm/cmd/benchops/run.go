package main

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

const (
	opcodesPkgPath = "gno.land/r/x/benchmark/opcodes"
	storagePkgPath = "gno.land/r/x/benchmark/storage"
	nativePkgPath  = "gno.land/r/x/benchmark/native"
	rounds         = 1000
)

// benchPackages holds pre-loaded package values so that
// package loading doesn't contaminate benchmark measurements.
type benchPackages struct {
	opcodes *gno.PackageValue
	storage *gno.PackageValue
	native  *gno.PackageValue
}

// loadBenchPackages loads all benchmark packages before the
// exporter is initialized. Store operations during loading
// are measured by the VM but not exported.
func loadBenchPackages(bstore BenchStore, dir string) benchPackages {
	gs := bstore.gnoStore
	var pkgs benchPackages

	// Opcodes
	pkgs.opcodes = addPackage(gs, filepath.Join(dir, "opcodes"), opcodesPkgPath)

	// Storage (needs avl dependency)
	addPackage(gs, filepath.Join(dir, "avl"), "gno.land/p/nt/avl/v0")
	pkgs.storage = addPackage(gs, filepath.Join(dir, "storage"), storagePkgPath)

	// Native
	pkgs.native = addPackage(gs, filepath.Join(dir, "native"), nativePkgPath)

	return pkgs
}

func benchmarkOpCodes(bstore gno.Store, pv *gno.PackageValue) {
	for range rounds {
		callOpsBench(bstore, pv)
	}
}

func callOpsBench(bstore gno.Store, pv *gno.PackageValue) {
	pb := pv.GetBlock(bstore)
	for _, tv := range pb.Values {
		if fv, ok := tv.V.(*gno.FuncValue); ok {
			cx := gno.Call(fv.Name)
			callFunc(bstore, pv, cx)
		}
	}
}

func benchmarkStorage(bstore BenchStore, pv *gno.PackageValue) {
	benchStoreSet(bstore, pv)
	benchStoreGet(bstore, pv)
}

func benchStoreSet(bstore BenchStore, pv *gno.PackageValue) {
	title := "1KB content"
	content := strings.Repeat("a", 1024)

	// in forum.gno: func AddPost(title, content string)
	// one AddPost will be added to three different boards in the forum.gno contract

	for range rounds {
		cx := gno.Call("AddPost", gno.Str(title), gno.Str(content))
		callFunc(bstore.gnoStore, pv, cx)
		bstore.Write()
		bstore.gnoStore.ClearObjectCache()
	}
}

func benchStoreGet(bstore BenchStore, pv *gno.PackageValue) {
	// in forum.gno: func GetPost(boardId, postId int) string  in forum.gno
	// there are three different boards on the benchmarking forum contract
	for i := range 3 {
		for j := range rounds {
			cx := gno.Call("GetPost", gno.X(i), gno.X(j))
			callFunc(bstore.gnoStore, pv, cx)
			bstore.Write()
			bstore.gnoStore.ClearObjectCache()
		}
	}
}

func benchmarkNative(bstore gno.Store, pv *gno.PackageValue) {
	for range rounds {
		callOpsBench(bstore, pv)
	}
}

func callFunc(gstore gno.Store, pv *gno.PackageValue, cx gno.Expr) []gno.TypedValue {
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath: pv.PkgPath,
			Output:  io.Discard,
			Store:   gstore,
		})

	defer m.Release()

	m.SetActivePackage(pv)
	return m.Eval(cx)
}

func addPackage(gstore gno.Store, dir string, pkgPath string) *gno.PackageValue {
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath: "",
			Output:  io.Discard,
			Store:   gstore,
		})
	defer m.Release()

	mpkg := gno.MustReadMemPackage(dir, pkgPath, gno.MPAnyProd)

	_, pv := m.RunMemPackage(mpkg, true)

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

		mpkg := gno.MustReadMemPackage(stdlibPath, pkgPath, gno.MPStdlibProd)
		if mpkg.IsEmpty() {
			// no gno files are present, skip this package
			return nil, nil
		}

		m2 := gno.NewMachineWithOptions(gno.MachineOptions{
			PkgPath: "gno.land/r/stdlibs/" + pkgPath,
			// PkgPath: pkgPath,
			Output: io.Discard,
			Store:  newStore,
		})
		defer m2.Release()
		return m2.RunMemPackage(mpkg, true)
	}

	bstore.gnoStore.SetPackageGetter(getPackage)
	bstore.gnoStore.SetNativeResolver(stdlibs.NativeResolver)
}
