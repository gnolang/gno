package cache_test

import (
	"crypto/rand"
	"sort"
	"testing"

	dbm "github.com/gnolang/gno/pkgs/db"

	"github.com/gnolang/gno/pkgs/store/cache"
	"github.com/gnolang/gno/pkgs/store/dbadapter"
)

func benchmarkCacheStoreIterator(numKVs int, b *testing.B) {
	mem := dbadapter.Store{DB: dbm.NewMemDB()}
	cstore := cache.New(mem)
	keys := make([]string, numKVs, numKVs)

	for i := 0; i < numKVs; i++ {
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

func BenchmarkCacheStoreIterator500(b *testing.B)    { benchmarkCacheStoreIterator(500, b) }
func BenchmarkCacheStoreIterator1000(b *testing.B)   { benchmarkCacheStoreIterator(1000, b) }
func BenchmarkCacheStoreIterator10000(b *testing.B)  { benchmarkCacheStoreIterator(10000, b) }
func BenchmarkCacheStoreIterator50000(b *testing.B)  { benchmarkCacheStoreIterator(50000, b) }
func BenchmarkCacheStoreIterator100000(b *testing.B) { benchmarkCacheStoreIterator(100000, b) }
