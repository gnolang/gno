//go:build cgo

package cache_batch_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/lmdbdb"
	"github.com/gnolang/gno/tm2/pkg/db/mdbxdb"
	"github.com/gnolang/gno/tm2/pkg/store/cache"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
)

func TestCacheBatchWriteLMDB(t *testing.T) {
	t.Parallel()
	db, err := lmdbdb.NewLMDB("test_lmdb", t.TempDir())
	require.NoError(t, err)
	defer db.Close()
	testCacheBatchWrite(t, db)
}

func TestCacheBatchWriteMDBX(t *testing.T) {
	t.Parallel()
	db, err := mdbxdb.NewMDBX("test_mdbx", t.TempDir())
	require.NoError(t, err)
	defer db.Close()
	testCacheBatchWrite(t, db)
}

func TestCacheBatchWriteOverwriteLMDB(t *testing.T) {
	t.Parallel()
	db, err := lmdbdb.NewLMDB("test_lmdb_ow", t.TempDir())
	require.NoError(t, err)
	defer db.Close()
	testCacheBatchWriteOverwrite(t, db)
}

func TestCacheBatchWriteOverwriteMDBX(t *testing.T) {
	t.Parallel()
	db, err := mdbxdb.NewMDBX("test_mdbx_ow", t.TempDir())
	require.NoError(t, err)
	defer db.Close()
	testCacheBatchWriteOverwrite(t, db)
}

func TestCacheBatchWriteEmptyLMDB(t *testing.T) {
	t.Parallel()
	db, err := lmdbdb.NewLMDB("test_lmdb_empty", t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	parent := dbadapter.Store{DB: db}
	cs := cache.New(parent)
	cs.Write()
}

func TestCacheBatchWriteEmptyMDBX(t *testing.T) {
	t.Parallel()
	db, err := mdbxdb.NewMDBX("test_mdbx_empty", t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	parent := dbadapter.Store{DB: db}
	cs := cache.New(parent)
	cs.Write()
}

func TestCacheBatchWriteSetThenDeleteLMDB(t *testing.T) {
	t.Parallel()
	db, err := lmdbdb.NewLMDB("test_lmdb_sd", t.TempDir())
	require.NoError(t, err)
	defer db.Close()
	testCacheBatchWriteSetThenDelete(t, db)
}

func TestCacheBatchWriteSetThenDeleteMDBX(t *testing.T) {
	t.Parallel()
	db, err := mdbxdb.NewMDBX("test_mdbx_sd", t.TempDir())
	require.NoError(t, err)
	defer db.Close()
	testCacheBatchWriteSetThenDelete(t, db)
}

func TestCacheBatchUsesDBBatch(t *testing.T) {
	t.Parallel()
	db, err := lmdbdb.NewLMDB("test_batch_path", t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	parent := dbadapter.Store{DB: db}
	var iface interface{} = parent
	_, ok := iface.(interface{ GetDB() dbm.DB })
	require.True(t, ok, "dbadapter.Store should implement GetDB()")

	cs := cache.New(parent)
	cs.Set(nil, []byte("k"), []byte("v"))
	cs.Write()

	got := parent.Get(nil, []byte("k"))
	require.Equal(t, []byte("v"), got)
}

func TestCacheBatchUsesDBBatchMDBX(t *testing.T) {
	t.Parallel()
	db, err := mdbxdb.NewMDBX("test_batch_path_mdbx", t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	parent := dbadapter.Store{DB: db}
	var iface interface{} = parent
	_, ok := iface.(interface{ GetDB() dbm.DB })
	require.True(t, ok, "dbadapter.Store should implement GetDB()")

	cs := cache.New(parent)
	cs.Set(nil, []byte("k"), []byte("v"))
	cs.Write()

	got := parent.Get(nil, []byte("k"))
	require.Equal(t, []byte("v"), got)
}

func TestCacheFallbackWriteLMDB(t *testing.T) {
	t.Parallel()
	db, err := lmdbdb.NewLMDB("test_fb_lmdb", t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	parent := nonBatchStore{db: db}
	cs := cache.New(parent)
	cs.Set(nil, []byte("k1"), []byte("v1"))
	cs.Write()

	v, err := db.Get([]byte("k1"))
	require.NoError(t, err)
	require.Equal(t, []byte("v1"), v)
}

func TestCacheFallbackWriteMDBX(t *testing.T) {
	t.Parallel()
	db, err := mdbxdb.NewMDBX("test_fb_mdbx", t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	parent := nonBatchStore{db: db}
	cs := cache.New(parent)
	cs.Set(nil, []byte("k1"), []byte("v1"))
	cs.Write()

	v, err := db.Get([]byte("k1"))
	require.NoError(t, err)
	require.Equal(t, []byte("v1"), v)
}

func TestCacheClearedAfterWriteLMDB(t *testing.T) {
	t.Parallel()
	db, err := lmdbdb.NewLMDB("test_clear_lmdb", t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	parent := dbadapter.Store{DB: db}
	cs := cache.New(parent)
	cs.Set(nil, []byte("k"), []byte("v1"))
	cs.Write()

	cs.Write() // second Write is no-op

	require.NoError(t, db.Set([]byte("k"), []byte("v2")))
	got := cs.Get(nil, []byte("k"))
	require.Equal(t, []byte("v2"), got)
}

func TestCacheClearedAfterWriteMDBX(t *testing.T) {
	t.Parallel()
	db, err := mdbxdb.NewMDBX("test_clear_mdbx", t.TempDir())
	require.NoError(t, err)
	defer db.Close()

	parent := dbadapter.Store{DB: db}
	cs := cache.New(parent)
	cs.Set(nil, []byte("k"), []byte("v1"))
	cs.Write()

	cs.Write() // second Write is no-op

	require.NoError(t, db.Set([]byte("k"), []byte("v2")))
	got := cs.Get(nil, []byte("k"))
	require.Equal(t, []byte("v2"), got)
}
