package bptree

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bp "github.com/gnolang/gno/tm2/pkg/bptree"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// --- Store pruning correctness with KeepRecent ---

func TestCoverage_StorePruningKeepRecent(t *testing.T) {
	db := memdb.NewMemDB()
	opts := types.StoreOptions{}
	opts.KeepRecent = 3
	opts.KeepEvery = 0
	st := StoreConstructor(db, opts).(*Store)

	// Commit 10 versions
	for i := 0; i < 10; i++ {
		st.Set(nil, []byte("k"), []byte{byte(i)})
		st.Commit()
	}

	// With KeepRecent=3, after version 10:
	// Versions 1-6 should be pruned, 7-10 should exist
	// (toRelease = previous - KeepRecent = 9-3 = 6 at version 10)
	for v := int64(1); v <= 6; v++ {
		require.False(t, st.VersionExists(v), "version %d should be pruned", v)
	}
	for v := int64(7); v <= 10; v++ {
		require.True(t, st.VersionExists(v), "version %d should exist", v)
	}
}

func TestCoverage_StorePruneEverythingDeletesOld(t *testing.T) {
	db := memdb.NewMemDB()
	opts := types.StoreOptions{} // KeepRecent=0, KeepEvery=0
	st := StoreConstructor(db, opts).(*Store)

	for i := 0; i < 5; i++ {
		st.Set(nil, []byte("k"), []byte{byte(i)})
		st.Commit()
	}

	// With PruneEverything, only the latest version should survive
	for v := int64(1); v < 5; v++ {
		require.False(t, st.VersionExists(v), "version %d should be pruned", v)
	}
	require.True(t, st.VersionExists(5), "latest version should exist")
}

// --- Query with proof ---

func TestCoverage_QueryWithProof(t *testing.T) {
	db := memdb.NewMemDB()
	opts := types.StoreOptions{}
	opts.KeepRecent = 100
	st := StoreConstructor(db, opts).(*Store)

	st.Set(nil, []byte("proofkey"), []byte("proofval"))
	cid := st.Commit()
	require.NotNil(t, cid.Hash)

	// Query with proof
	res := st.Query(makeQuery("/key", []byte("proofkey"), cid.Version, true))
	require.Equal(t, []byte("proofval"), res.Value)
	require.NotNil(t, res.Proof, "proof should not be nil")

	// Query non-existent key with proof
	res = st.Query(makeQuery("/key", []byte("missing"), cid.Version, true))
	require.Nil(t, res.Value)
	require.NotNil(t, res.Proof, "non-existence proof should not be nil")
}

// --- Query with height=0 ---

func TestCoverage_QueryDefaultHeight(t *testing.T) {
	db := memdb.NewMemDB()
	opts := types.StoreOptions{}
	opts.KeepRecent = 100
	st := StoreConstructor(db, opts).(*Store)

	st.Set(nil, []byte("h0key"), []byte("val1"))
	st.Commit() // v1

	st.Set(nil, []byte("h0key"), []byte("val2"))
	st.Commit() // v2

	// Query with height=0 should use latest-1 = v1
	res := st.Query(makeQuery("/key", []byte("h0key"), 0, false))
	// Note: height=0 picks latest-1 if it exists, else latest
	// With 2 versions, it should pick v1 which has "val1"
	require.Equal(t, []byte("val1"), res.Value)
}

// --- Immutable store operations ---

func TestCoverage_LoadLatestVersionImmutable(t *testing.T) {
	db := memdb.NewMemDB()
	opts := types.StoreOptions{}
	opts.KeepRecent = 100
	st := StoreConstructor(db, opts).(*Store)

	st.Set(nil, []byte("imm"), []byte("val"))
	st.Commit()

	// Reload as immutable
	opts2 := types.StoreOptions{}
	opts2.KeepRecent = 100
	opts2.Immutable = true
	st2 := UnsafeNewStore(bp.NewMutableTreeWithDB(db, 100, bp.NewNopLogger()), opts2)
	err := st2.LoadLatestVersion()
	require.NoError(t, err)

	// Should be able to read
	val := st2.Get(nil, []byte("imm"))
	require.Equal(t, []byte("val"), val)

	// Should panic on write
	require.Panics(t, func() { st2.Set(nil, []byte("x"), []byte("y")) })
}

func TestCoverage_LoadVersionImmutable(t *testing.T) {
	db := memdb.NewMemDB()
	opts := types.StoreOptions{}
	opts.KeepRecent = 100
	st := StoreConstructor(db, opts).(*Store)

	st.Set(nil, []byte("lv"), []byte("v1"))
	st.Commit()
	st.Set(nil, []byte("lv"), []byte("v2"))
	st.Commit()

	// Load version 1 as immutable
	opts2 := types.StoreOptions{}
	opts2.Immutable = true
	opts2.KeepRecent = 100
	st2 := UnsafeNewStore(bp.NewMutableTreeWithDB(db, 100, bp.NewNopLogger()), opts2)
	err := st2.LoadVersion(1)
	require.NoError(t, err)

	val := st2.Get(nil, []byte("lv"))
	require.Equal(t, []byte("v1"), val)
}

func TestCoverage_ImmutableStoreCommitPanics(t *testing.T) {
	db := memdb.NewMemDB()
	opts := types.StoreOptions{}
	opts.KeepRecent = 100
	st := StoreConstructor(db, opts).(*Store)

	st.Set(nil, []byte("x"), []byte("y"))
	st.Commit()

	immSt, err := st.GetImmutable(1)
	require.NoError(t, err)
	defer immSt.Close()

	require.Panics(t, func() { immSt.Commit() })
}

// --- Iterator on immutable store ---

func TestCoverage_ImmutableStoreIterator(t *testing.T) {
	db := memdb.NewMemDB()
	opts := types.StoreOptions{}
	opts.KeepRecent = 100
	st := StoreConstructor(db, opts).(*Store)

	st.Set(nil, []byte("a"), []byte("1"))
	st.Set(nil, []byte("b"), []byte("2"))
	st.Set(nil, []byte("c"), []byte("3"))
	st.Commit()

	immSt, err := st.GetImmutable(1)
	require.NoError(t, err)
	defer immSt.Close()

	itr := immSt.Iterator(nil, nil, nil)
	defer itr.Close()
	count := 0
	for itr.Valid() {
		require.NotNil(t, itr.Key())
		require.NotNil(t, itr.Value())
		count++
		itr.Next()
	}
	require.Equal(t, 3, count)
}

// --- Query before commit ---

func TestCoverage_QueryBeforeCommit(t *testing.T) {
	db := memdb.NewMemDB()
	opts := types.StoreOptions{}
	opts.KeepRecent = 100
	st := StoreConstructor(db, opts).(*Store)

	st.Set(nil, []byte("uncommitted"), []byte("val"))
	// Query before commit — version 0 should return empty
	res := st.Query(makeQuery("/key", []byte("uncommitted"), 0, false))
	require.Nil(t, res.Value, "uncommitted data should not appear in queries")
}

// --- Helper ---

func makeQuery(path string, data []byte, height int64, prove bool) abci.RequestQuery {
	return abci.RequestQuery{Path: path, Data: data, Height: height, Prove: prove}
}
