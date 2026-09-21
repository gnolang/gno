package gnolang

import (
	"fmt"
	"io"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// BenchmarkPackageLoadFromStore measures loading a persisted N-file package
// from the store and touching a single file's block — the shape of a
// transaction that enters one function of an imported package. Each iteration
// uses a fresh transaction store (empty object cache) so the load cost is paid
// every time, mirroring per-message isolation on chain.
//
// This benchmarks the path this change optimizes: on master, GetPackage
// eagerly materializes all N file blocks (deriveFBlocksMap in fillPackage); with
// lazy loading, only the one touched block is materialized. Compare across
// versions with benchstat:
//
//	git stash && go test ./gnovm/pkg/gnolang/ -run x -bench BenchmarkPackageLoadFromStore -benchmem -count=6 > old.txt
//	git stash pop && go test ./gnovm/pkg/gnolang/ -run x -bench BenchmarkPackageLoadFromStore -benchmem -count=6 > new.txt
//	benchstat old.txt new.txt
//
// The single-file case (files=1) is the guard case: it must not regress, since
// there is no unused file to skip.
func BenchmarkPackageLoadFromStore(b *testing.B) {
	for _, nFiles := range []int{1, 4, 12} {
		b.Run(fmt.Sprintf("files=%d", nFiles), func(b *testing.B) {
			const pkgPath = "gno.vm/t/bench"
			db := memdb.NewMemDB()
			tm2Store := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
			st := NewStore(nil, tm2Store, tm2Store)

			// Create and persist the package once.
			wrapped := tm2Store.CacheWrap()
			txSt := st.BeginTransaction(wrapped, wrapped, nil, nil)
			files := make([]*std.MemFile, nFiles)
			for i := range files {
				files[i] = &std.MemFile{
					Name: fmt.Sprintf("f%d.gno", i),
					Body: fmt.Sprintf("package bench\n\nfunc F%d() int { return %d }", i, i),
				}
			}
			m := NewMachineWithOptions(MachineOptions{
				PkgPath: pkgPath,
				Store:   txSt,
				Output:  io.Discard,
			})
			m.RunMemPackage(&std.MemPackage{
				Type:  MPUserProd,
				Name:  "bench",
				Path:  pkgPath,
				Files: files,
			}, true)
			txSt.Write()
			wrapped.Write()

			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				cw := tm2Store.CacheWrap()
				txSt2 := st.BeginTransaction(cw, cw, nil, nil)
				pv := txSt2.GetPackage(pkgPath, false)
				// Touch one file, as entering one function would.
				_ = pv.GetFileBlock(txSt2, "f0.gno")
			}
		})
	}
}
