package rootmulti

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	_ "github.com/gnolang/gno/tm2/pkg/db/pebbledb"
	storebptree "github.com/gnolang/gno/tm2/pkg/store/bptree"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

type rootReadBarrierDB struct {
	dbm.DB
	armed   atomic.Bool
	once    sync.Once
	reached chan struct{}
	resume  chan struct{}
	batches atomic.Int64
}

func (db *rootReadBarrierDB) waitOnRoot(key []byte) {
	if db.armed.Load() && bytes.HasPrefix(key, []byte("s/_/R")) {
		db.once.Do(func() {
			db.armed.Store(false)
			close(db.reached)
			<-db.resume
		})
	}
}

func (db *rootReadBarrierDB) Get(key []byte) ([]byte, error) {
	db.waitOnRoot(key)
	return db.DB.Get(key)
}

func (db *rootReadBarrierDB) NewSnapshot() (dbm.Snapshot, error) {
	snapshot, err := db.DB.NewSnapshot()
	if err != nil {
		return nil, err
	}
	return &rootReadBarrierSnapshot{Snapshot: snapshot, db: db}, nil
}

func (db *rootReadBarrierDB) NewBatch() dbm.Batch {
	db.batches.Add(1)
	return db.DB.NewBatch()
}

func (db *rootReadBarrierDB) NewBatchWithSize(size int) dbm.Batch {
	db.batches.Add(1)
	return db.DB.NewBatchWithSize(size)
}

type rootReadBarrierSnapshot struct {
	dbm.Snapshot
	db *rootReadBarrierDB
}

func (s *rootReadBarrierSnapshot) Get(key []byte) ([]byte, error) {
	s.db.waitOnRoot(key)
	return s.Snapshot.Get(key)
}

type noSnapshotDB struct {
	dbm.DB
}

func (db *noSnapshotDB) NewSnapshot() (dbm.Snapshot, error) {
	return nil, fmt.Errorf("snapshots unsupported")
}

// TestImmutableDedicatedMountWithoutSnapshotSupport ensures the read-only
// fallback can load a dedicated B+Tree store without constructing a live batch.
func TestImmutableDedicatedMountWithoutSnapshotSupport(t *testing.T) {
	db := &noSnapshotDB{DB: memdb.NewMemDB()}
	ms := NewMultiStore(db)
	key := types.NewStoreKey("main")
	ms.MountStoreWithDB(key, storebptree.FastStoreConstructor, db)
	require.NoError(t, ms.LoadLatestVersion())
	ms.GetStore(key).Set(nil, []byte("key"), []byte("value"))
	version := ms.Commit().Version

	require.NotPanics(t, func() {
		snapshot, release, err := ms.MultiImmutableCacheWrapWithVersion(version)
		require.NoError(t, err)
		defer release()
		require.Equal(t, []byte("value"), snapshot.GetStore(key).Get(nil, []byte("key")))
	})
}

// TestImmutableDedicatedMountCannotRebuildFromStaleRoot reproduces the query
// versus commit TOCTOU and verifies the query stays snapshot-backed and write-free.
func TestImmutableDedicatedMountCannotRebuildFromStaleRoot(t *testing.T) {
	db := &rootReadBarrierDB{
		DB:      memdb.NewMemDB(),
		reached: make(chan struct{}),
		resume:  make(chan struct{}),
	}
	ms := NewMultiStore(db)
	ms.SetStoreOptions(types.StoreOptions{PruningOptions: types.PruneNothing})
	key := types.NewStoreKey("main")
	ms.MountStoreWithDB(key, storebptree.FastStoreConstructor, db)
	require.NoError(t, ms.LoadLatestVersion())

	live := ms.GetStore(key)
	live.Set(nil, []byte("account"), []byte("v1"))
	v1 := ms.Commit().Version
	batchesBeforeQuery := db.batches.Load()
	db.armed.Store(true)

	type queryResult struct {
		value   []byte
		release func()
		err     error
	}
	result := make(chan queryResult, 1)
	go func() {
		snapshot, release, err := ms.MultiImmutableCacheWrapWithVersion(v1)
		if err != nil {
			result <- queryResult{err: err}
			return
		}
		result <- queryResult{
			value:   snapshot.GetStore(key).Get(nil, []byte("account")),
			release: release,
		}
	}()

	<-db.reached
	t.Log("query loaded root v1; committing writer root v2 before query continues")
	queryLiveBatches := db.batches.Load() - batchesBeforeQuery
	live.Set(nil, []byte("account"), []byte("v2"))
	v2 := ms.Commit().Version
	close(db.resume)
	require.Equal(t, int64(2), v2)

	query := <-result
	require.NoError(t, query.err)
	defer query.release()
	require.Equal(t, []byte("v1"), query.value)

	require.Equal(t, int64(3), ms.Commit().Version)
	fastValue := live.Get(nil, []byte("account"))

	plain := storebptree.StoreConstructor(
		dbm.NewPrefixDB(db, []byte("s/_/")),
		types.StoreOptions{Immutable: true},
	).(*storebptree.Store)
	require.NoError(t, plain.LoadVersion(3))
	treeValue := plain.Get(nil, []byte("account"))
	t.Logf("query live batches=%d fast=%q tree=%q", queryLiveBatches, fastValue, treeValue)

	require.Equal(t, []byte("v2"), fastValue)
	require.Equal(t, treeValue, fastValue)
	require.Zero(t, queryLiveBatches, "immutable query constructed batches against the live dedicated DB")
}

// TestFastIndexConcurrentSnapshotsKeepTreeParity ensures Pebble-backed query
// snapshots remain pinned while concurrent commits preserve fast/tree parity.
func TestFastIndexConcurrentSnapshotsKeepTreeParity(t *testing.T) {
	db, err := dbm.NewDB("fast-index-race", dbm.PebbleDBBackend, t.TempDir())
	require.NoError(t, err)
	ms := NewMultiStore(db)
	t.Cleanup(func() {
		require.NoError(t, ms.Close())
		require.NoError(t, db.Close())
	})
	ms.SetStoreOptions(types.StoreOptions{PruningOptions: types.PruneNothing})
	storeKey := types.NewStoreKey("main")
	ms.MountStoreWithDB(storeKey, storebptree.FastStoreConstructor, db)
	require.NoError(t, ms.LoadLatestVersion())

	live := ms.GetStore(storeKey)
	for i := range 16 {
		live.Set(nil, []byte{byte(i)}, []byte{0})
	}
	seedVersion := ms.Commit().Version

	const readers = 4
	errs := make(chan error, readers)
	var wg sync.WaitGroup
	for range readers {
		wg.Go(func() {
			for range 50 {
				snapshot, release, err := ms.MultiImmutableCacheWrapWithVersion(seedVersion)
				if err != nil {
					errs <- err
					return
				}
				for i := range 16 {
					if got := snapshot.GetStore(storeKey).Get(nil, []byte{byte(i)}); !bytes.Equal(got, []byte{0}) {
						release()
						errs <- fmt.Errorf("seed key %d = %x", i, got)
						return
					}
				}
				release()
			}
		})
	}
	for version := byte(1); version <= 50; version++ {
		live.Set(nil, []byte{version % 16}, []byte{version})
		ms.Commit()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	latest := ms.LastCommitID().Version
	plain := storebptree.StoreConstructor(
		dbm.NewPrefixDB(db, []byte("s/_/")),
		types.StoreOptions{Immutable: true},
	).(*storebptree.Store)
	require.NoError(t, plain.LoadVersion(latest))
	for i := range 16 {
		require.Equal(t,
			plain.Get(nil, []byte{byte(i)}),
			live.Get(nil, []byte{byte(i)}),
			"key %d fast/tree mismatch", i,
		)
	}
}
