package rootmulti

import (
	"strings"
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
	store1.Set(k, v)

	cID := ms.Commit()
	require.Equal(t, int64(1), cID.Version)

	// require failure when given an invalid or pruned version
	_, err = ms.MultiImmutableCacheWrapWithVersion(cID.Version + 1)
	require.Error(t, err)

	// require a valid version can be cache-loaded
	cms, err := ms.MultiImmutableCacheWrapWithVersion(cID.Version)
	require.NoError(t, err)

	// require a valid key lookup yields the correct value
	kvStore := cms.GetStore(ms.keysByName["store1"])
	require.NotNil(t, kvStore)
	require.Equal(t, kvStore.Get(k), v)

	// require we cannot commit (write) to a cache-versioned multi-store
	require.Panics(t, func() {
		kvStore.Set(k, []byte("newValue"))
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
	store1.Set(k, v)

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
	store1.Set(k, v)

	// ... and another.
	store2 := multi.getStoreByName("store2")
	store2.Set(k2, v2)

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
