package cache_test

import (
	"crypto/rand"
	"sort"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/cache"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
)

func benchmarkCacheStoreIterator(b *testing.B, numKVs int) {
	b.Helper()

	mem := dbadapter.Store{DB: memdb.NewMemDB()}
	cstore := cache.New(mem)
	keys := make([]string, numKVs)

	for i := range numKVs {
		key := make([]byte, 32)
		value := make([]byte, 32)

		_, _ = rand.Read(key)
		_, _ = rand.Read(value)

		keys[i] = string(key)
		cstore.Set(key, value)
	}

	sort.Strings(keys)

	for n := 0; n < b.N; n++ {
		iter := cstore.Iterator([]byte(keys[0]), []byte(keys[numKVs-1]))

		for _ = iter.Key(); iter.Valid(); iter.Next() {
		}

		iter.Close()
	}
}

func BenchmarkCacheStoreIterator500(b *testing.B)    { benchmarkCacheStoreIterator(b, 500) }
func BenchmarkCacheStoreIterator1000(b *testing.B)   { benchmarkCacheStoreIterator(b, 1000) }
func BenchmarkCacheStoreIterator10000(b *testing.B)  { benchmarkCacheStoreIterator(b, 10000) }
func BenchmarkCacheStoreIterator50000(b *testing.B)  { benchmarkCacheStoreIterator(b, 50000) }
func BenchmarkCacheStoreIterator100000(b *testing.B) { benchmarkCacheStoreIterator(b, 100000) }
