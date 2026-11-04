package rootmulti

import (
	"bytes"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	storebptree "github.com/gnolang/gno/tm2/pkg/store/bptree"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// newHoldingTestStore mounts what gno.land mounts, a bptree store with the
// fast index and a plain store, and prunes every older version, so held
// commits also prune versions that are only held.
func newHoldingTestStore(t *testing.T, db dbm.DB) (ms *multiStore, mainKey, baseKey types.StoreKey) {
	t.Helper()

	ms = NewMultiStore(db)
	ms.SetStoreOptions(types.StoreOptions{PruningOptions: types.PruneEverything})
	mainKey, baseKey = types.NewStoreKey("main"), types.NewStoreKey("base")
	ms.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, db)
	ms.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, db)
	require.NoError(t, ms.LoadLatestVersion())
	return ms, mainKey, baseKey
}

// commitTestBlock writes a key to each store the way BaseApp does, through a
// cache wrap, and commits.
func commitTestBlock(ms *multiStore, mainKey, baseKey types.StoreKey, n int) types.CommitID {
	cache := ms.MultiCacheWrap()
	cache.GetStore(mainKey).Set(nil, fmt.Appendf(nil, "k%d", n), fmt.Appendf(nil, "v%d", n))
	cache.GetStore(baseKey).Set(nil, fmt.Appendf(nil, "b%d", n), fmt.Appendf(nil, "w%d", n))
	cache.MultiWrite()
	return ms.Commit()
}

// TestHoldingCommits pins the holding contract: held commits compute the same
// versions and hashes as written ones, are readable, and reach the disk only
// at WriteHeld, as one synced batch; a store reopened before that, as after a
// crash, is at the last written version.
func TestHoldingCommits(t *testing.T) {
	t.Parallel()

	var writeSyncs atomic.Int32
	db := &countingDB{DB: memdb.NewMemDB(), writeSyncs: &writeSyncs}
	ms, mainKey, baseKey := newHoldingTestStore(t, db)
	plain, plainMain, plainBase := newHoldingTestStore(t, memdb.NewMemDB())

	require.Equal(t, commitTestBlock(plain, plainMain, plainBase, 1), commitTestBlock(ms, mainKey, baseKey, 1))
	require.Equal(t, int64(1), ms.WrittenCommitID().Version)

	ms.SetHolding(true)
	writeSyncs.Store(0)
	for n := 2; n <= 4; n++ {
		require.Equal(t, commitTestBlock(plain, plainMain, plainBase, n), commitTestBlock(ms, mainKey, baseKey, n),
			"a held commit must have the hash a written one has")
	}
	require.Zero(t, writeSyncs.Load(), "a held commit reached the disk")
	require.Equal(t, int64(4), ms.LastCommitID().Version)
	require.Equal(t, int64(1), ms.WrittenCommitID().Version)
	require.Equal(t, int64(1), getLatestVersion(db))
	require.Equal(t, []byte("v4"), ms.GetStore(mainKey).Get(nil, []byte("k4")))
	require.Equal(t, []byte("w3"), ms.GetStore(baseKey).Get(nil, []byte("b3")))
	require.Positive(t, ms.HeldBytes())

	// Queries read the disk: the written version, not a held one.
	_, release, err := ms.MultiImmutableCacheWrapWithVersion(1)
	require.NoError(t, err)
	release()
	_, _, err = ms.MultiImmutableCacheWrapWithVersion(4)
	require.Error(t, err, "a query view of a held version")

	// A crash now loses the held commits and nothing else.
	crashed, crashedMain, _ := newHoldingTestStore(t, db)
	require.Equal(t, int64(1), crashed.LastCommitID().Version)
	require.Equal(t, []byte("v1"), crashed.GetStore(crashedMain).Get(nil, []byte("k1")))
	require.Nil(t, crashed.GetStore(crashedMain).Get(nil, []byte("k2")))

	require.NoError(t, ms.WriteHeld())
	require.Equal(t, int32(1), writeSyncs.Load(), "WriteHeld must write one synced batch")
	require.Zero(t, ms.HeldBytes())
	require.Equal(t, ms.LastCommitID(), ms.WrittenCommitID())
	require.Equal(t, int64(4), getLatestVersion(db))

	reloaded, reloadedMain, reloadedBase := newHoldingTestStore(t, db)
	require.Equal(t, ms.LastCommitID(), reloaded.LastCommitID())
	for n := 1; n <= 4; n++ {
		require.Equal(t, fmt.Appendf(nil, "v%d", n), reloaded.GetStore(reloadedMain).Get(nil, fmt.Appendf(nil, "k%d", n)))
		require.Equal(t, fmt.Appendf(nil, "w%d", n), reloaded.GetStore(reloadedBase).Get(nil, fmt.Appendf(nil, "b%d", n)))
	}

	ms.SetHolding(false)
	require.Equal(t, commitTestBlock(plain, plainMain, plainBase, 5), commitTestBlock(ms, mainKey, baseKey, 5))
	require.Equal(t, int64(5), ms.WrittenCommitID().Version)
}

// TestHoldingRefusesDatabaseIterators pins that, while holding, an iterator
// over the database panics and is recorded, while the bptree store's own
// iterator, which walks its tree, still sees held writes.
func TestHoldingRefusesDatabaseIterators(t *testing.T) {
	t.Parallel()

	ms, mainKey, baseKey := newHoldingTestStore(t, memdb.NewMemDB())
	ms.SetHolding(true)
	commitTestBlock(ms, mainKey, baseKey, 1)

	it := ms.GetStore(mainKey).Iterator(nil, nil, nil)
	require.True(t, it.Valid())
	require.Equal(t, []byte("k1"), it.Key())
	it.Close()
	require.False(t, ms.TakeRefusedIterator())

	func() {
		defer func() {
			err, _ := recover().(error)
			require.True(t, errors.Is(err, dbm.ErrIteratorOverHeldWrites), "recovered %v", err)
		}()
		ms.GetStore(baseKey).Iterator(nil, nil, nil)
	}()
	require.True(t, ms.TakeRefusedIterator())
	require.False(t, ms.TakeRefusedIterator())
}

// TestHoldingOffWithCommitsHeldPanics pins that holding cannot be turned off
// while commits are held: they would be written with the next commit, late.
func TestHoldingOffWithCommitsHeldPanics(t *testing.T) {
	t.Parallel()

	ms, mainKey, baseKey := newHoldingTestStore(t, memdb.NewMemDB())
	ms.SetHolding(true)
	commitTestBlock(ms, mainKey, baseKey, 1)
	require.Panics(t, func() { ms.SetHolding(false) })
	require.NoError(t, ms.WriteHeld())
	require.NotPanics(t, func() { ms.SetHolding(false) })
}

// TestHoldingWritesWhatLoadCollected pins that turning holding on first
// writes what a store collected outside a commit, as the bptree fast index's
// rebuild at load does, so that it is on disk and a window with no commit
// holds nothing: holding turns off again at once, as block sync does when it
// switches to consensus without a block.
func TestHoldingWritesWhatLoadCollected(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	ms, mainKey, baseKey := newHoldingTestStore(t, db)
	commitTestBlock(ms, mainKey, baseKey, 1)

	// Drop the fast index's completeness stamp: the next load rebuilds it.
	var stamp []byte
	it, err := db.Iterator(nil, nil)
	require.NoError(t, err)
	for ; it.Valid(); it.Next() {
		if bytes.HasSuffix(it.Key(), []byte("fastidx")) {
			stamp = append([]byte(nil), it.Key()...)
		}
	}
	it.Close()
	require.NotNil(t, stamp, "no fast-index stamp on disk")
	require.NoError(t, db.Delete(stamp))

	ms, _, _ = newHoldingTestStore(t, db)
	require.Positive(t, ms.collector.Len(), "the rebuild at load collected nothing")
	ms.SetHolding(true)
	require.Zero(t, ms.collector.Len())
	v, err := db.Get(stamp)
	require.NoError(t, err)
	require.NotNil(t, v, "the rebuilt stamp is not on disk")
	require.NotPanics(t, func() { ms.SetHolding(false) })
}

// TestHoldingRefusesLiveQueries pins that, while holding, a query on the live
// stores is refused rather than answered from commits that are not on disk.
func TestHoldingRefusesLiveQueries(t *testing.T) {
	t.Parallel()

	ms, mainKey, baseKey := newHoldingTestStore(t, memdb.NewMemDB())
	commitTestBlock(ms, mainKey, baseKey, 1)
	query := abci.RequestQuery{Path: "/main/key", Data: []byte("k1"), Height: 1}
	require.Nil(t, ms.Query(query).Error)

	ms.SetHolding(true)
	commitTestBlock(ms, mainKey, baseKey, 2)
	require.NotNil(t, ms.Query(query).Error, "a live query while holding")
}
