package cache_batch_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	_ "github.com/gnolang/gno/tm2/pkg/db/pebbledb"
	"github.com/gnolang/gno/tm2/pkg/store/cache"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	store "github.com/gnolang/gno/tm2/pkg/store/types"
)

func TestCacheBatchWritePebbleDB(t *testing.T) {
	t.Parallel()
	db, err := dbm.NewDB("test_pebble", dbm.PebbleDBBackend, t.TempDir())
	require.NoError(t, err)
	defer db.Close()
	testCacheBatchWrite(t, db)
}

func testCacheBatchWrite(t *testing.T, db dbm.DB) {
	t.Helper()
	parent := dbadapter.Store{DB: db}
	cs := cache.New(parent)

	const n = 500
	for i := 0; i < n; i++ {
		k := fmt.Sprintf("key%04d", i)
		v := fmt.Sprintf("val%04d", i)
		cs.Set(nil, []byte(k), []byte(v))
	}
	for i := 0; i < 50; i++ {
		k := fmt.Sprintf("key%04d", i)
		cs.Delete(nil, []byte(k))
	}

	// Not visible in parent yet.
	val := parent.Get(nil, []byte("key0100"))
	require.Nil(t, val)

	cs.Write()

	for i := 50; i < n; i++ {
		k := fmt.Sprintf("key%04d", i)
		v := fmt.Sprintf("val%04d", i)
		got := parent.Get(nil, []byte(k))
		require.Equal(t, []byte(v), got, "key %s", k)
	}
	for i := 0; i < 50; i++ {
		k := fmt.Sprintf("key%04d", i)
		got := parent.Get(nil, []byte(k))
		require.Nil(t, got, "key %s should be deleted", k)
	}
}

func TestCacheBatchWriteOverwritePebbleDB(t *testing.T) {
	t.Parallel()
	db, err := dbm.NewDB("test_pebble_ow", dbm.PebbleDBBackend, t.TempDir())
	require.NoError(t, err)
	defer db.Close()
	testCacheBatchWriteOverwrite(t, db)
}

func testCacheBatchWriteOverwrite(t *testing.T, db dbm.DB) {
	t.Helper()
	parent := dbadapter.Store{DB: db}

	require.NoError(t, db.Set([]byte("existing"), []byte("old")))

	cs := cache.New(parent)
	cs.Set(nil, []byte("existing"), []byte("new"))
	cs.Set(nil, []byte("fresh"), []byte("val"))
	cs.Write()

	got := parent.Get(nil, []byte("existing"))
	require.Equal(t, []byte("new"), got)
	got = parent.Get(nil, []byte("fresh"))
	require.Equal(t, []byte("val"), got)
}

func TestCacheBatchWriteEmptyPebbleDB(t *testing.T) {
	t.Parallel()
	db, err := dbm.NewDB("test_pebble_empty", dbm.PebbleDBBackend, t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	parent := dbadapter.Store{DB: db}
	cs := cache.New(parent)
	cs.Write()
}

func TestCacheBatchWriteSetThenDeletePebbleDB(t *testing.T) {
	t.Parallel()
	db, err := dbm.NewDB("test_pebble_sd", dbm.PebbleDBBackend, t.TempDir())
	require.NoError(t, err)
	defer db.Close()
	testCacheBatchWriteSetThenDelete(t, db)
}

func testCacheBatchWriteSetThenDelete(t *testing.T, db dbm.DB) {
	t.Helper()
	parent := dbadapter.Store{DB: db}
	cs := cache.New(parent)

	cs.Set(nil, []byte("k"), []byte("v"))
	cs.Delete(nil, []byte("k"))
	cs.Write()

	got := parent.Get(nil, []byte("k"))
	require.Nil(t, got)
}

// nonBatchStore wraps a DB but does NOT implement GetDB(),
// forcing the cache to use the non-batch fallback path.
type nonBatchStore struct {
	db dbm.DB
}

func (s nonBatchStore) Get(gctx *store.GasContext, key []byte) []byte {
	v, _ := s.db.Get(key)
	return v
}
func (s nonBatchStore) Has(gctx *store.GasContext, key []byte) bool {
	v, _ := s.db.Has(key)
	return v
}
func (s nonBatchStore) Set(gctx *store.GasContext, key, value []byte) { s.db.Set(key, value) }
func (s nonBatchStore) Delete(gctx *store.GasContext, key []byte)     { s.db.Delete(key) }
func (s nonBatchStore) Iterator(gctx *store.GasContext, start, end []byte) store.Iterator {
	it, _ := s.db.Iterator(start, end)
	return it
}
func (s nonBatchStore) ReverseIterator(gctx *store.GasContext, start, end []byte) store.Iterator {
	it, _ := s.db.ReverseIterator(start, end)
	return it
}
func (s nonBatchStore) CacheWrap() store.Store               { return cache.New(s) }
func (s nonBatchStore) Write()                               { panic("unexpected") }
func (s nonBatchStore) Flush()                               {}
func (s nonBatchStore) SetStoreOptions(_ store.StoreOptions) {}
func (s nonBatchStore) GetStoreOptions() store.StoreOptions  { return store.StoreOptions{} }

func TestCacheFallbackWritePebbleDB(t *testing.T) {
	t.Parallel()
	db, err := dbm.NewDB("test_fb_pebble", dbm.PebbleDBBackend, t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	parent := nonBatchStore{db: db}
	cs := cache.New(parent)
	cs.Set(nil, []byte("k1"), []byte("v1"))
	cs.Set(nil, []byte("k2"), []byte("v2"))
	cs.Delete(nil, []byte("k2"))
	cs.Write()

	v, err := db.Get([]byte("k1"))
	require.NoError(t, err)
	require.Equal(t, []byte("v1"), v)

	v, err = db.Get([]byte("k2"))
	require.NoError(t, err)
	require.Nil(t, v)
}

func TestCacheClearedAfterWritePebbleDB(t *testing.T) {
	t.Parallel()
	db, err := dbm.NewDB("test_clear_pebble", dbm.PebbleDBBackend, t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	parent := dbadapter.Store{DB: db}
	cs := cache.New(parent)
	cs.Set(nil, []byte("k"), []byte("v1"))
	cs.Write()

	// After Write, cache is cleared. A second Write is a no-op.
	cs.Write()

	// Update the parent directly — cache should not mask it.
	require.NoError(t, db.Set([]byte("k"), []byte("v2")))
	got := cs.Get(nil, []byte("k"))
	require.Equal(t, []byte("v2"), got)
}
