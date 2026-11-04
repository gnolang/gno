package cache_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/cache"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// The genesis balance-loading shape: for each of n accounts, do a narrow
// prefix iteration over that account's own key range, then write its key.
// Every iteration triggers dirtyItems, which scans the whole unsortedCache.
func benchAccountLoad(b *testing.B, n int) {
	b.Helper()
	for i := 0; i < b.N; i++ {
		parent := dbadapter.Store{DB: memdb.NewMemDB()}
		st := cache.New(parent)
		var gctx *types.GasContext // nil is valid: GasContext methods are nil-safe
		for a := range n {
			pfx := fmt.Appendf(nil, "bal/%08d/", a)
			// the splitCoins read: narrow prefix iteration
			start, end := pfx, types.PrefixEndBytes(pfx)
			it := st.Iterator(gctx, start, end)
			for ; it.Valid(); it.Next() {
				_ = it.Key()
			}
			it.Close()
			// the setSplitBalance write
			st.Set(gctx, append(append([]byte{}, pfx...), []byte("ugnot")...), []byte("1000"))
		}
	}
}

func BenchmarkAccountLoad1000(b *testing.B)  { benchAccountLoad(b, 1000) }
func BenchmarkAccountLoad2000(b *testing.B)  { benchAccountLoad(b, 2000) }
func BenchmarkAccountLoad4000(b *testing.B)  { benchAccountLoad(b, 4000) }
func BenchmarkAccountLoad8000(b *testing.B)  { benchAccountLoad(b, 8000) }
func BenchmarkAccountLoad16000(b *testing.B) { benchAccountLoad(b, 16000) }

// Isolate the two suspected O(n) terms independently.

// only writes, no iteration -> should be linear
func BenchmarkWritesOnly16000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		st := cache.New(dbadapter.Store{DB: memdb.NewMemDB()})
		var gctx *types.GasContext // nil is valid: GasContext methods are nil-safe
		for a := range 16000 {
			st.Set(gctx, fmt.Appendf(nil, "bal/%08d/ugnot", a), []byte("1000"))
		}
	}
}
