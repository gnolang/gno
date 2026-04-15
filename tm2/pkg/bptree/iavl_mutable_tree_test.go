package bptree

// Ported from tm2/pkg/iavl/mutable_tree_test.go
// Skips fast-node-specific tests (TestUpgradeStorageToFast*, etc.)
// since B+ tree doesn't have a fast node index.

import (
	"bytes"
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

var (
	tKey1 = []byte("k1")
	tKey2 = []byte("k2")
	tVal1 = []byte("v1")
	tVal2 = []byte("v2")
)

// setupMutableTree creates a DB-backed tree for testing.
func setupMutableTree(_ bool) *MutableTree {
	db := memdb.NewMemDB()
	return NewMutableTreeWithDB(db, 1000, NewNopLogger())
}

// prepareTree creates a tree with 2 saved versions.
// v1: key {1} = "a"
// v2: key {1} = "b" (updated)
func prepareTree(t *testing.T) *MutableTree {
	t.Helper()
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 0, NewNopLogger())

	tree.Set([]byte{1}, []byte("a"))
	_, _, err := tree.SaveVersion()
	require.NoError(t, err)

	tree.Set([]byte{1}, []byte("b"))
	_, _, err = tree.SaveVersion()
	require.NoError(t, err)

	return tree
}

func TestDeleteVersionsTo(t *testing.T) {
	tree := setupMutableTree(false)

	tree.Set([]byte("k1"), []byte("Fred"))
	_, version, err := tree.SaveVersion()
	require.NoError(t, err)
	_, _, err = tree.SaveVersion()
	require.NoError(t, err)

	require.NoError(t, tree.DeleteVersionsTo(version))

	// Version 1 should be gone, version 2 should remain
	require.False(t, tree.VersionExists(version))
	require.True(t, tree.VersionExists(version+1))
}

func TestDeleteVersionsFrom(t *testing.T) {
	tree := setupMutableTree(false)

	tree.Set([]byte("k1"), []byte("Wilma"))
	_, version, err := tree.SaveVersion()
	require.NoError(t, err)
	_, _, err = tree.SaveVersion()
	require.NoError(t, err)
	_, _, err = tree.SaveVersion()
	require.NoError(t, err)

	require.NoError(t, tree.DeleteVersionsFrom(version+1))

	require.True(t, tree.VersionExists(version))
	require.False(t, tree.VersionExists(version+1))
	require.False(t, tree.VersionExists(version+2))
}

func TestDeleteVersionsFrom_ResetsWorkingTree(t *testing.T) {
	tree := setupMutableTree(false)

	tree.Set([]byte("a"), []byte("v1"))
	tree.SaveVersion() // v1
	tree.Set([]byte("b"), []byte("v2"))
	tree.SaveVersion() // v2
	tree.Set([]byte("c"), []byte("v3"))
	tree.SaveVersion() // v3

	// Working tree is at version 3
	require.Equal(t, int64(3), tree.Version())
	require.Equal(t, int64(3), tree.Size())

	// Delete versions >= 2 — working tree must reset to v1
	require.NoError(t, tree.DeleteVersionsFrom(2))

	require.Equal(t, int64(1), tree.Version())
	require.Equal(t, int64(1), tree.Size())
	require.True(t, tree.VersionExists(1))
	require.False(t, tree.VersionExists(2))
	require.False(t, tree.VersionExists(3))

	// The tree should only have key "a"
	val, _ := tree.Get([]byte("a"))
	require.Equal(t, []byte("v1"), val)
	val, _ = tree.Get([]byte("b"))
	require.Nil(t, val)
	val, _ = tree.Get([]byte("c"))
	require.Nil(t, val)

	// Can save a new version 2
	tree.Set([]byte("d"), []byte("v2new"))
	_, v, err := tree.SaveVersion()
	require.NoError(t, err)
	require.Equal(t, int64(2), v)
}

func TestDeleteVersionsFrom_DeleteAll(t *testing.T) {
	tree := setupMutableTree(false)

	tree.Set([]byte("a"), []byte("v1"))
	tree.SaveVersion()
	tree.Set([]byte("b"), []byte("v2"))
	tree.SaveVersion()

	// Delete ALL versions (fromVersion=1)
	require.NoError(t, tree.DeleteVersionsFrom(1))

	require.Equal(t, int64(0), tree.Version())
	require.Equal(t, int64(0), tree.Size())
	require.True(t, tree.IsEmpty())

	// Can start fresh with version 1
	tree.Set([]byte("x"), []byte("fresh"))
	_, v, err := tree.SaveVersion()
	require.NoError(t, err)
	require.Equal(t, int64(1), v)
	val, _ := tree.Get([]byte("x"))
	require.Equal(t, []byte("fresh"), val)
}

func TestGetRemove(t *testing.T) {
	require := require.New(t)
	tree := setupMutableTree(false)

	testGet := func(exists bool) {
		v, err := tree.Get(tKey1)
		require.NoError(err)
		if exists {
			require.Equal(tVal1, v, "key should exist")
		} else {
			require.Nil(v, "key should not exist")
		}
	}

	testGet(false)

	ok, err := tree.Set(tKey1, tVal1)
	require.NoError(err)
	require.False(ok, "new key set: nothing to update")

	ok, err = tree.Set(tKey2, tVal2)
	require.NoError(err)
	require.False(ok, "new key set: nothing to update")

	testGet(true)

	_, version, err := tree.SaveVersion()
	require.NoError(err)
	require.Equal(int64(1), version)

	testGet(true)

	v, ok, err := tree.Remove(tKey1)
	require.NoError(err)
	require.True(ok, "key should be removed")
	require.Equal(tVal1, v, "key should exist")

	testGet(false)
}

func TestTraverse(t *testing.T) {
	tree := setupMutableTree(false)

	for i := 0; i < 6; i++ {
		tree.Set([]byte(fmt.Sprintf("k%d", i)), []byte(fmt.Sprintf("v%d", i)))
	}

	require.Equal(t, int64(6), tree.Size(), "Size of tree unexpected")
}

func TestMutableTree_DeleteVersionsTo(t *testing.T) {
	tree := setupMutableTree(false)

	type entry struct {
		key   []byte
		value []byte
	}

	versionEntries := make(map[int64][]entry)
	r := rand.New(rand.NewSource(42))

	// create 10 tree versions, each with 1000 random key/value entries
	for i := 0; i < 10; i++ {
		entries := make([]entry, 1000)

		for j := 0; j < 1000; j++ {
			k := make([]byte, 10)
			v := make([]byte, 10)
			r.Read(k)
			r.Read(v)

			entries[j] = entry{k, v}
			tree.Set(k, v)
		}

		_, ver, err := tree.SaveVersion()
		require.NoError(t, err)
		versionEntries[ver] = entries
	}

	// delete versions up to 8
	versionToDelete := int64(8)
	require.NoError(t, tree.DeleteVersionsTo(versionToDelete))

	// ensure deleted versions cannot be loaded
	for v := int64(1); v <= versionToDelete; v++ {
		require.False(t, tree.VersionExists(v))
	}

	// ensure remaining versions exist and data is queryable
	for _, v := range []int64{9, 10} {
		_, err := tree.LoadVersion(v)
		require.NoError(t, err)

		for _, e := range versionEntries[v] {
			val, err := tree.Get(e.key)
			require.NoError(t, err)
			if val != nil {
				require.True(t, bytes.Equal(e.value, val))
			}
		}
	}
}

func TestMutableTree_LoadVersion_Empty(t *testing.T) {
	tree := setupMutableTree(false)

	version, err := tree.LoadVersion(0)
	require.NoError(t, err)
	assert.EqualValues(t, 0, version)
}

func TestMutableTree_InitialVersion(t *testing.T) {
	memDB := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(memDB, 0, NewNopLogger(), InitialVersionOption(9))

	tree.Set([]byte("a"), []byte{0x01})
	_, version, err := tree.SaveVersion()
	require.NoError(t, err)
	assert.EqualValues(t, 9, version)

	tree.Set([]byte("b"), []byte{0x02})
	_, version, err = tree.SaveVersion()
	require.NoError(t, err)
	assert.EqualValues(t, 10, version)

	// Reloading the tree with the same initial version is fine
	tree = NewMutableTreeWithDB(memDB, 0, NewNopLogger(), InitialVersionOption(9))
	version, err = tree.Load()
	require.NoError(t, err)
	assert.EqualValues(t, 10, version)
}

func TestMutableTree_SetInitialVersion(t *testing.T) {
	tree := setupMutableTree(false)
	tree.SetInitialVersion(9)

	tree.Set([]byte("a"), []byte{0x01})
	_, version, err := tree.SaveVersion()
	require.NoError(t, err)
	assert.EqualValues(t, 9, version)
}

func TestMutableTree_Version(t *testing.T) {
	tree := prepareTree(t)
	require.True(t, tree.VersionExists(1))
	require.True(t, tree.VersionExists(2))
	require.False(t, tree.VersionExists(3))
}

func TestMutableTree_GetVersioned(t *testing.T) {
	tree := prepareTree(t)

	// Check versioned values
	val, err := tree.GetVersioned([]byte{1}, 1)
	require.NoError(t, err)
	require.Equal(t, []byte("a"), val)

	val, err = tree.GetVersioned([]byte{1}, 2)
	require.NoError(t, err)
	require.Equal(t, []byte("b"), val)
}

func TestMutableTree_DeleteVersion(t *testing.T) {
	tree := prepareTree(t)

	require.NoError(t, tree.DeleteVersionsTo(1))

	require.False(t, tree.VersionExists(1))
	require.True(t, tree.VersionExists(2))
	require.False(t, tree.VersionExists(3))

	// cannot delete latest version
	require.Error(t, tree.DeleteVersionsTo(2))
}

func TestMutableTree_LazyLoadVersionWithEmptyTree(t *testing.T) {
	mdb := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(mdb, 1000, NewNopLogger())
	_, v1, err := tree.SaveVersion()
	require.NoError(t, err)

	newTree1 := NewMutableTreeWithDB(mdb, 1000, NewNopLogger())
	v2, err := newTree1.LoadVersion(1)
	require.NoError(t, err)
	require.True(t, v1 == v2)
}

func TestMutableTree_SetSimple(t *testing.T) {
	mdb := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(mdb, 0, NewNopLogger())

	isUpdated, err := tree.Set([]byte("a"), []byte("test"))
	require.NoError(t, err)
	require.False(t, isUpdated)

	fastValue, err := tree.Get([]byte("a"))
	require.NoError(t, err)
	_, regularValue, err := tree.GetWithIndex([]byte("a"))
	require.NoError(t, err)

	require.Equal(t, []byte("test"), fastValue)
	require.Equal(t, []byte("test"), regularValue)
}

func TestMutableTree_SetTwoKeys(t *testing.T) {
	tree := setupMutableTree(false)

	isUpdated, err := tree.Set([]byte("a"), []byte("test"))
	require.NoError(t, err)
	require.False(t, isUpdated)

	isUpdated, err = tree.Set([]byte("b"), []byte("test2"))
	require.NoError(t, err)
	require.False(t, isUpdated)

	fastValue, err := tree.Get([]byte("a"))
	require.NoError(t, err)
	_, regularValue, err := tree.GetWithIndex([]byte("a"))
	require.NoError(t, err)
	require.Equal(t, []byte("test"), fastValue)
	require.Equal(t, []byte("test"), regularValue)

	fastValue2, err := tree.Get([]byte("b"))
	require.NoError(t, err)
	_, regularValue2, err := tree.GetWithIndex([]byte("b"))
	require.NoError(t, err)
	require.Equal(t, []byte("test2"), fastValue2)
	require.Equal(t, []byte("test2"), regularValue2)
}

func TestMutableTree_SetOverwrite(t *testing.T) {
	tree := setupMutableTree(false)

	isUpdated, err := tree.Set([]byte("a"), []byte("test"))
	require.NoError(t, err)
	require.False(t, isUpdated)

	isUpdated, err = tree.Set([]byte("a"), []byte("test2"))
	require.NoError(t, err)
	require.True(t, isUpdated)

	fastValue, err := tree.Get([]byte("a"))
	require.NoError(t, err)
	_, regularValue, err := tree.GetWithIndex([]byte("a"))
	require.NoError(t, err)
	require.Equal(t, []byte("test2"), fastValue)
	require.Equal(t, []byte("test2"), regularValue)
}

func TestMutableTree_SetRemoveSet(t *testing.T) {
	tree := setupMutableTree(false)

	// Set
	isUpdated, err := tree.Set([]byte("a"), []byte("test"))
	require.NoError(t, err)
	require.False(t, isUpdated)

	fastValue, err := tree.Get([]byte("a"))
	require.NoError(t, err)
	require.Equal(t, []byte("test"), fastValue)

	// Remove
	removedVal, removed, err := tree.Remove([]byte("a"))
	require.NoError(t, err)
	require.True(t, removed)
	require.Equal(t, []byte("test"), removedVal)

	fastValue, err = tree.Get([]byte("a"))
	require.NoError(t, err)
	require.Nil(t, fastValue)

	// Set again
	isUpdated, err = tree.Set([]byte("a"), []byte("test2"))
	require.NoError(t, err)
	require.False(t, isUpdated)

	fastValue, err = tree.Get([]byte("a"))
	require.NoError(t, err)
	require.Equal(t, []byte("test2"), fastValue)
}

func TestIterate_MutableTree_Unsaved(t *testing.T) {
	tree, mirror := getRandomizedTreeAndMirrorForMutable(t)
	assertMutableMirrorIterate(t, tree, mirror)
}

func TestIterate_MutableTree_Saved(t *testing.T) {
	tree, mirror := getRandomizedTreeAndMirrorForMutable(t)
	_, _, err := tree.SaveVersion()
	require.NoError(t, err)
	assertMutableMirrorIterate(t, tree, mirror)
}

func TestIterate_MutableTree_Unsaved_NextVersion(t *testing.T) {
	tree, mirror := getRandomizedTreeAndMirrorForMutable(t)
	_, _, err := tree.SaveVersion()
	require.NoError(t, err)

	// Add more random entries for the next version
	randomizeMutableTreeAndMirror(t, tree, mirror)
	assertMutableMirrorIterate(t, tree, mirror)
}

func TestIterator_MutableTree_Invalid(t *testing.T) {
	tree := setupMutableTree(false)

	itr, err := tree.Iterator([]byte("a"), []byte("b"), true)
	require.NoError(t, err)
	require.False(t, itr.Valid())
	itr.Close()
}

func TestMutableTree_InitialVersion_FirstVersion(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 0, NewNopLogger(), InitialVersionOption(1))

	tree.Set([]byte("a"), []byte("1"))
	_, v, err := tree.SaveVersion()
	require.NoError(t, err)
	require.Equal(t, int64(1), v)

	tree.Set([]byte("b"), []byte("2"))
	_, v, err = tree.SaveVersion()
	require.NoError(t, err)
	require.Equal(t, int64(2), v)
}

func TestMutableTreeClose(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 0, NewNopLogger())
	tree.Set([]byte("a"), []byte("1"))
	_, _, err := tree.SaveVersion()
	require.NoError(t, err)

	require.NoError(t, tree.Close())
}

func TestMutableTree_InitialVersionZero(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 0, NewNopLogger(), InitialVersionOption(0))

	tree.Set([]byte("a"), []byte("1"))
	_, v, err := tree.SaveVersion()
	require.NoError(t, err)
	require.Equal(t, int64(1), v)
}

// --- Helpers for mutable tree iteration tests ---

func getRandomizedTreeAndMirrorForMutable(t *testing.T) (*MutableTree, map[string]string) {
	t.Helper()
	tree := setupMutableTree(false)
	mirror := make(map[string]string)
	r := rand.New(rand.NewSource(99))
	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("mkey_%03d", r.Intn(200))
		v := fmt.Sprintf("mval_%d", i)
		tree.Set([]byte(k), []byte(v))
		mirror[k] = v
	}
	return tree, mirror
}

func randomizeMutableTreeAndMirror(t *testing.T, tree *MutableTree, mirror map[string]string) {
	t.Helper()
	r := rand.New(rand.NewSource(123))
	for i := 0; i < 50; i++ {
		k := fmt.Sprintf("mkey_%03d", r.Intn(200))
		v := fmt.Sprintf("mval2_%d", i)
		tree.Set([]byte(k), []byte(v))
		mirror[k] = v
	}
}

func assertMutableMirrorIterate(t *testing.T, tree *MutableTree, mirror map[string]string) {
	t.Helper()
	mirrorKeys := make([]string, 0, len(mirror))
	for k := range mirror {
		mirrorKeys = append(mirrorKeys, k)
	}
	sort.Strings(mirrorKeys)

	i := 0
	tree.Iterate(func(key, value []byte) bool {
		require.Less(t, i, len(mirrorKeys), "too many keys in tree")
		require.Equal(t, mirrorKeys[i], string(key))
		require.Equal(t, mirror[mirrorKeys[i]], string(value))
		i++
		return false
	})
	require.Equal(t, len(mirrorKeys), i, "not enough keys in tree")
}
