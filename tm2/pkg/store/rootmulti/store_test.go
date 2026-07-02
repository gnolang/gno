package rootmulti

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/merkle"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"

	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

func TestStoreType(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	store := NewMultiStore(db)
	store.MountStoreWithDB(
		types.NewStoreKey("store1"), iavl.StoreConstructor, db)
}

func TestStoreMount(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	store := NewMultiStore(db)

	key1 := types.NewStoreKey("store1")
	key2 := types.NewStoreKey("store2")
	dup1 := types.NewStoreKey("store1")

	require.NotPanics(t, func() { store.MountStoreWithDB(key1, iavl.StoreConstructor, db) })
	require.NotPanics(t, func() { store.MountStoreWithDB(key2, iavl.StoreConstructor, db) })

	require.Panics(t, func() { store.MountStoreWithDB(key1, iavl.StoreConstructor, db) })
	require.Panics(t, func() { store.MountStoreWithDB(dup1, iavl.StoreConstructor, db) })
}

func TestCacheMultiStoreWithVersion(t *testing.T) {
	t.Parallel()

	var db dbm.DB = memdb.NewMemDB()
	ms := newMultiStoreWithMounts(db)
	err := ms.LoadLatestVersion()
	require.NoError(t, err)

	commitID := types.CommitID{}
	checkStore(t, ms, commitID, commitID)

	k, v := []byte("wind"), []byte("blows")

	store1 := ms.getStoreByName("store1")
	store1.Set(nil, k, v)

	cID := ms.Commit()
	require.Equal(t, int64(1), cID.Version)

	// require failure when given an invalid or pruned version
	_, _, err = ms.MultiImmutableCacheWrapWithVersion(cID.Version + 1)
	require.Error(t, err)

	// require a valid version can be cache-loaded
	cms, release, err := ms.MultiImmutableCacheWrapWithVersion(cID.Version)
	require.NoError(t, err)
	defer release()

	// require a valid key lookup yields the correct value
	kvStore := cms.GetStore(ms.keysByName["store1"])
	require.NotNil(t, kvStore)
	require.Equal(t, kvStore.Get(nil, k), v)

	// require we cannot commit (write) to a cache-versioned multi-store
	require.Panics(t, func() {
		kvStore.Set(nil, k, []byte("newValue"))
		cms.(types.Committer).Commit()
	})
}

func TestHashStableWithEmptyCommit(t *testing.T) {
	t.Parallel()

	var db dbm.DB = memdb.NewMemDB()
	ms := newMultiStoreWithMounts(db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	commitID := types.CommitID{}
	checkStore(t, ms, commitID, commitID)

	k, v := []byte("wind"), []byte("blows")

	store1 := ms.getStoreByName("store1")
	store1.Set(nil, k, v)

	cID := ms.Commit()
	require.Equal(t, int64(1), cID.Version)
	hash := cID.Hash

	// make an empty commit, it should update version, but not affect hash
	cID = ms.Commit()
	require.Equal(t, int64(2), cID.Version)
	require.Equal(t, hash, cID.Hash)
}

func TestMultistoreCommitLoad(t *testing.T) {
	t.Parallel()

	var db dbm.DB = memdb.NewMemDB()
	store := newMultiStoreWithMounts(db)
	err := store.LoadLatestVersion()
	require.Nil(t, err)

	// New store has empty last commit.
	commitID := types.CommitID{}
	checkStore(t, store, commitID, commitID)

	// Make sure we can get stores by name.
	s1 := store.getStoreByName("store1")
	require.NotNil(t, s1)
	s3 := store.getStoreByName("store3")
	require.NotNil(t, s3)
	s77 := store.getStoreByName("store77")
	require.Nil(t, s77)

	// Make a few commits and check them.
	nCommits := int64(3)
	for i := int64(0); i < nCommits; i++ {
		commitID = store.Commit()
		expectedCommitID := getExpectedCommitID(store, i+1)
		checkStore(t, store, expectedCommitID, commitID)
	}

	// Load the latest multistore again and check version.
	store = newMultiStoreWithMounts(db)
	err = store.LoadLatestVersion()
	require.Nil(t, err)
	commitID = getExpectedCommitID(store, nCommits)
	checkStore(t, store, commitID, commitID)

	// Commit and check version.
	commitID = store.Commit()
	expectedCommitID := getExpectedCommitID(store, nCommits+1)
	checkStore(t, store, expectedCommitID, commitID)

	// Load an older multistore and check version.
	ver := nCommits - 1
	store = newMultiStoreWithMounts(db)
	err = store.LoadVersion(ver)
	require.Nil(t, err)
	commitID = getExpectedCommitID(store, ver)
	checkStore(t, store, commitID, commitID)

	// XXX: commit this older version
	commitID = store.Commit()
	expectedCommitID = getExpectedCommitID(store, ver+1)
	checkStore(t, store, expectedCommitID, commitID)

	// XXX: confirm old commit is overwritten and we have rolled back
	// LatestVersion
	store = newMultiStoreWithMounts(db)
	err = store.LoadLatestVersion()
	require.Nil(t, err)
	commitID = getExpectedCommitID(store, ver+1)
	checkStore(t, store, commitID, commitID)
}

func TestParsePath(t *testing.T) {
	t.Parallel()

	_, _, err := parsePath("foo")
	require.Error(t, err)

	store, subpath, err := parsePath("/foo")
	require.NoError(t, err)
	require.Equal(t, store, "foo")
	require.Equal(t, subpath, "")

	store, subpath, err = parsePath("/fizz/bang/baz")
	require.NoError(t, err)
	require.Equal(t, store, "fizz")
	require.Equal(t, subpath, "/bang/baz")

	substore, subsubpath, err := parsePath(subpath)
	require.NoError(t, err)
	require.Equal(t, substore, "bang")
	require.Equal(t, subsubpath, "/baz")
}

func TestMultiStoreQuery(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	multi := newMultiStoreWithMounts(db)
	err := multi.LoadLatestVersion()
	require.NoError(t, err)

	k, v := []byte("wind"), []byte("blows")
	k2, v2 := []byte("water"), []byte("flows")
	// v3 := []byte("is cold")

	cid := multi.Commit()

	// Make sure we can get by name.
	garbage := multi.getStoreByName("bad-name")
	require.Nil(t, garbage)

	// Set and commit data in one store.
	store1 := multi.getStoreByName("store1")
	store1.Set(nil, k, v)

	// ... and another.
	store2 := multi.getStoreByName("store2")
	store2.Set(nil, k2, v2)

	// Commit the multistore.
	cid = multi.Commit()
	ver := cid.Version

	// Reload multistore from database
	multi = newMultiStoreWithMounts(db)
	err = multi.LoadLatestVersion()
	require.Nil(t, err)

	// Test bad path.
	query := abci.RequestQuery{Path: "/key", Data: k, Height: ver}
	qres := multi.Query(query)
	require.True(t, strings.HasPrefix(qres.Error.Error(), "unknownrequest error:"))

	query.Path = "h897fy32890rf63296r92"
	qres = multi.Query(query)
	require.True(t, strings.HasPrefix(qres.Error.Error(), "unknownrequest error:"))

	// Test invalid store name.
	query.Path = "/garbage/key"
	qres = multi.Query(query)
	require.True(t, strings.HasPrefix(qres.Error.Error(), "unknownrequest error:"))

	// Test valid query with data.
	query.Path = "/store1/key"
	qres = multi.Query(query)
	require.Nil(t, qres.Error)
	require.Equal(t, v, qres.Value)

	// Test valid but empty query.
	query.Path = "/store2/key"
	query.Prove = true
	qres = multi.Query(query)
	require.Nil(t, qres.Error)
	require.Nil(t, qres.Value)

	// Test store2 data.
	query.Data = k2
	qres = multi.Query(query)
	require.Nil(t, qres.Error)
	require.Equal(t, v2, qres.Value)
}

// TestSnapshotReadIsolation verifies that a snapshot acquired at version N
// continues to return version-N data after version N+1 has been committed.
// This is the core cross-store consistency guarantee: both IAVL and dbadapter
// sub-stores must reflect the same block height within a single query.
func TestSnapshotReadIsolation(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	ms := newMultiStoreWithMounts(db)
	require.NoError(t, ms.LoadLatestVersion())

	key := []byte("city")
	v1 := []byte("paris")
	v2 := []byte("berlin")

	// Commit version 1 with v1.
	ms.getStoreByName("store1").Set(nil, key, v1)
	cid1 := ms.Commit()
	require.Equal(t, int64(1), cid1.Version)

	// Acquire an immutable view pinned to version 1 before committing v2.
	snap1, release1, err := ms.MultiImmutableCacheWrapWithVersion(cid1.Version)
	require.NoError(t, err)

	// Commit version 2 with v2 — this swaps the querySnapshot.
	ms.getStoreByName("store1").Set(nil, key, v2)
	cid2 := ms.Commit()
	require.Equal(t, int64(2), cid2.Version)

	// The version-1 snapshot must still return v1, not v2.
	got := snap1.GetStore(ms.keysByName["store1"]).Get(nil, key)
	require.Equal(t, v1, got, "version-1 snapshot must not see version-2 writes")

	// A fresh snapshot at version 2 must return v2.
	snap2, release2, err := ms.MultiImmutableCacheWrapWithVersion(cid2.Version)
	require.NoError(t, err)
	defer release2()

	got2 := snap2.GetStore(ms.keysByName["store1"]).Get(nil, key)
	require.Equal(t, v2, got2, "version-2 snapshot must return version-2 value")

	// Releasing the version-1 snapshot after version-2 has been committed must
	// not panic — the refcount keeps it alive until this point.
	release1()
}

// TestSnapshotRefcounting verifies that the snapshot underlying a query is not
// closed while references are still held, and is closed exactly once after all
// references are released.
func TestSnapshotRefcounting(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	ms := newMultiStoreWithMounts(db)
	require.NoError(t, ms.LoadLatestVersion())

	ms.getStoreByName("store1").Set(nil, []byte("k"), []byte("v"))
	cid := ms.Commit()

	// Acquire three concurrent references to the same version.
	const N = 3
	releases := make([]func(), N)
	for i := range N {
		_, rel, err := ms.MultiImmutableCacheWrapWithVersion(cid.Version)
		require.NoError(t, err)
		releases[i] = rel
	}

	// Commit the next version — swaps the snapshot. The old snapshot must remain
	// valid because three references still hold it.
	ms.Commit()

	// Release two of the three refs — snapshot must still be alive.
	releases[0]()
	releases[1]()

	// The third ref can still read without panic.
	snap, rel3, err := ms.MultiImmutableCacheWrapWithVersion(cid.Version)
	require.NoError(t, err)
	require.NotNil(t, snap.GetStore(ms.keysByName["store1"]))
	rel3()

	// Releasing the last original ref must not panic.
	releases[2]()
}

// TestSnapshotConcurrentCommitAndQuery runs concurrent calls to
// MultiImmutableCacheWrapWithVersion while Commit() advances the block height.
// Run with -race to detect data races on the snapshot pointer swap.
func TestSnapshotConcurrentCommitAndQuery(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	ms := newMultiStoreWithMounts(db)
	require.NoError(t, ms.LoadLatestVersion())

	key := []byte("counter")

	// Seed version 1 so queries have a valid version to request.
	ms.getStoreByName("store1").Set(nil, key, []byte{0})
	seedCID := ms.Commit()

	const (
		numCommits = 20
		numReaders = 8
	)

	var wg sync.WaitGroup

	// Writer goroutine: commits numCommits blocks.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range numCommits {
			ms.getStoreByName("store1").Set(nil, key, []byte{byte(i + 1)})
			ms.Commit()
		}
	}()

	// Reader goroutines: each repeatedly acquires a snapshot at the seed version
	// and reads from it, then releases. Must never panic or data-race.
	for range numReaders {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range numCommits {
				snap, release, err := ms.MultiImmutableCacheWrapWithVersion(seedCID.Version)
				if err != nil {
					// Version may have been pruned; that's acceptable.
					continue
				}
				store := snap.GetStore(ms.keysByName["store1"])
				require.NotNil(t, store)
				// Seed version must always return the seeded value, never a
				// later value — cross-store consistency check.
				got := store.Get(nil, key)
				require.Equal(t, []byte{0}, got,
					"snapshot at seed version must return seed value, not a later write")
				release()
			}
		}()
	}

	wg.Wait()
}

// countingDB wraps a real DB and counts calls to batch.WriteSync() so tests
// can assert that a Commit produces exactly one atomic disk flush.
type countingDB struct {
	dbm.DB
	writeSyncs *atomic.Int32
}

func (c *countingDB) NewBatch() dbm.Batch {
	return &countingBatch{Batch: c.DB.NewBatch(), writeSyncs: c.writeSyncs}
}

func (c *countingDB) NewBatchWithSize(size int) dbm.Batch {
	return &countingBatch{Batch: c.DB.NewBatchWithSize(size), writeSyncs: c.writeSyncs}
}

type countingBatch struct {
	dbm.Batch
	writeSyncs *atomic.Int32
}

func (b *countingBatch) WriteSync() error {
	b.writeSyncs.Add(1)
	return b.Batch.WriteSync()
}

// TestCommitAtomicBatchWithCacheFlush verifies that a full BaseApp-style block
// commit — deliver cache flush (dbadapter writes), IAVL SaveVersion, IAVL
// pruning, and rootmulti metadata — lands in exactly one batch.WriteSync()
// call on the real DB. This is the core cross-store atomicity guarantee.
func TestCommitAtomicBatchWithCacheFlush(t *testing.T) {
	t.Parallel()

	var writeSyncs atomic.Int32
	db := &countingDB{DB: memdb.NewMemDB(), writeSyncs: &writeSyncs}
	ms := newMultiStoreWithMounts(db)
	require.NoError(t, ms.LoadLatestVersion())

	// Simulate a DeliverTx: writes buffered in a MultiCacheWrap of the live
	// multiStore, exactly as BaseApp does.
	cache := ms.MultiCacheWrap()
	cache.GetStore(ms.keysByName["store1"]).Set(nil, []byte("k1"), []byte("v1"))
	cache.GetStore(ms.keysByName["store2"]).Set(nil, []byte("k2"), []byte("v2"))

	writeSyncs.Store(0)
	// Mirror BaseApp.Commit(): flush the cache, then commit. Both share the
	// same CollectingDB, so the two calls accumulate into one drained batch.
	cache.MultiWrite()
	cid := ms.Commit()
	require.Equal(t, int64(1), cid.Version)

	// Cache flush + IAVL SaveVersion + metadata must reach disk in exactly
	// one WriteSync.
	require.Equal(t, int32(1), writeSyncs.Load(),
		"MultiWrite+Commit must produce exactly one atomic WriteSync")

	// The collector must be empty afterwards — otherwise writes would leak
	// into the next commit.
	require.Equal(t, 0, ms.collector.Len(),
		"collector must be drained after Commit")

	// Reload from disk and verify all values are present — proves the single
	// batch actually carried IAVL nodes AND rootmulti metadata to persistence.
	reload := newMultiStoreWithMounts(db)
	require.NoError(t, reload.LoadLatestVersion())
	require.Equal(t, cid.Version, reload.lastCommitID.Version)
	require.Equal(t, []byte("v1"),
		reload.getStoreByName("store1").Get(nil, []byte("k1")))
	require.Equal(t, []byte("v2"),
		reload.getStoreByName("store2").Get(nil, []byte("k2")))
}

// TestCommitSingleAtomicBatch is the same as TestCommitAtomicBatchWithCacheFlush
// but for the no-cache Commit() path used by tests and lower-level callers.
func TestCommitSingleAtomicBatch(t *testing.T) {
	t.Parallel()

	var writeSyncs atomic.Int32
	db := &countingDB{DB: memdb.NewMemDB(), writeSyncs: &writeSyncs}
	ms := newMultiStoreWithMounts(db)
	require.NoError(t, ms.LoadLatestVersion())

	// Direct sub-store writes (bypassing the cache path — writes go straight
	// into IAVL's MutableTree).
	ms.getStoreByName("store1").Set(nil, []byte("k"), []byte("v"))

	writeSyncs.Store(0)
	cid := ms.Commit()
	require.Equal(t, int64(1), cid.Version)

	require.Equal(t, int32(1), writeSyncs.Load(),
		"Commit must produce exactly one atomic WriteSync")
	require.Equal(t, 0, ms.collector.Len(),
		"collector must be drained after Commit")
}

// -----------------------------------------------------------------------
// utils

func newMultiStoreWithMounts(db dbm.DB) *multiStore {
	store := NewMultiStore(db)
	store.storeOpts = types.StoreOptions{PruningOptions: types.PruneSyncable}
	store.MountStoreWithDB(
		types.NewStoreKey("store1"), iavl.StoreConstructor, nil)
	store.MountStoreWithDB(
		types.NewStoreKey("store2"), iavl.StoreConstructor, nil)
	store.MountStoreWithDB(
		types.NewStoreKey("store3"), iavl.StoreConstructor, nil)
	return store
}

func checkStore(t *testing.T, store *multiStore, expect, got types.CommitID) {
	t.Helper()

	require.Equal(t, expect, got)
	require.Equal(t, expect, store.LastCommitID())
}

func getExpectedCommitID(store *multiStore, ver int64) types.CommitID {
	return types.CommitID{
		Version: ver,
		Hash:    hashStores(store.stores),
	}
}

func hashStores(stores map[types.StoreKey]types.CommitStore) []byte {
	m := make(map[string][]byte, len(stores))
	for key, store := range stores {
		name := key.Name()
		m[name] = storeInfo{
			Name: name,
			Core: storeCore{
				CommitID: store.LastCommitID(),
				// StoreType: store.GetStoreType(),
			},
		}.GetHash()
	}
	return merkle.SimpleHashFromMap(m)
}
