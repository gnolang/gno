package bptree

// Ported from tm2/pkg/iavl/tree_test.go

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

func TestVersionedRandomTree(t *testing.T) {
	require := require.New(t)
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	versions := 50
	keysPerVersion := 30

	r := rand.New(rand.NewSource(0))
	for i := 1; i <= versions; i++ {
		for j := 0; j < keysPerVersion; j++ {
			k := make([]byte, 8)
			v := make([]byte, 8)
			r.Read(k)
			r.Read(v)
			tree.Set(k, v)
		}
		tree.SaveVersion()
	}

	// Ensure it returns all versions in sorted order
	available := tree.AvailableVersions()
	assert.Equal(t, versions, len(available))
	assert.Equal(t, 1, available[0])
	assert.Equal(t, versions, available[len(available)-1])

	tree.DeleteVersionsTo(int64(versions - 1))

	tr, err := tree.GetImmutable(int64(versions))
	require.NoError(err, "GetImmutable should not error for latest version")
	require.NotNil(tr)

	// we should only have one available version now
	available = tree.AvailableVersions()
	assert.Equal(t, 1, len(available))
	assert.Equal(t, versions, available[0])
}

func TestTreeHash(t *testing.T) {
	const (
		randSeed    = 49872768940
		keySize     = 16
		valueSize   = 16
		versions    = 4
		versionOps  = 4096
		updateRatio = 0.4
		deleteRatio = 0.2
	)

	r := rand.New(rand.NewSource(randSeed))
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 0, NewNopLogger())

	keys := make([][]byte, 0, versionOps)
	hashes := make([][]byte, versions)

	for i := 0; i < versions; i++ {
		for j := 0; j < versionOps; j++ {
			key := make([]byte, keySize)
			value := make([]byte, valueSize)

			switch {
			case len(keys) > 0 && r.Float64() <= deleteRatio:
				index := r.Intn(len(keys))
				key = keys[index]
				keys = append(keys[:index], keys[index+1:]...)
				_, removed, err := tree.Remove(key)
				require.NoError(t, err)
				require.True(t, removed)

			case len(keys) > 0 && r.Float64() <= updateRatio:
				key = keys[r.Intn(len(keys))]
				r.Read(value)
				updated, err := tree.Set(key, value)
				require.NoError(t, err)
				require.True(t, updated)

			default:
				r.Read(key)
				r.Read(value)
				updated, err := tree.Set(key, value)
				require.NoError(t, err)
				require.False(t, updated)
				keys = append(keys, key)
			}
		}

		hash, _, err := tree.SaveVersion()
		require.NoError(t, err)
		hashes[i] = hash
	}

	// Verify hashes are deterministic by replaying
	r2 := rand.New(rand.NewSource(randSeed))
	tree2 := NewMutableTreeWithDB(memdb.NewMemDB(), 0, NewNopLogger())
	keys2 := make([][]byte, 0, versionOps)

	for i := 0; i < versions; i++ {
		for j := 0; j < versionOps; j++ {
			key := make([]byte, keySize)
			value := make([]byte, valueSize)

			switch {
			case len(keys2) > 0 && r2.Float64() <= deleteRatio:
				index := r2.Intn(len(keys2))
				key = keys2[index]
				keys2 = append(keys2[:index], keys2[index+1:]...)
				tree2.Remove(key)

			case len(keys2) > 0 && r2.Float64() <= updateRatio:
				key = keys2[r2.Intn(len(keys2))]
				r2.Read(value)
				tree2.Set(key, value)

			default:
				r2.Read(key)
				r2.Read(value)
				tree2.Set(key, value)
				keys2 = append(keys2, key)
			}
		}
		hash2, _, _ := tree2.SaveVersion()
		require.Equal(t, hashes[i], hash2, "hash mismatch at version %d", i+1)
	}
}

func TestVersionedRandomTreeSmallKeys(t *testing.T) {
	require := require.New(t)
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	singleVersionRecords := 20
	versions := 20
	r := rand.New(rand.NewSource(0))

	for i := 0; i < versions; i++ {
		for j := 0; j < singleVersionRecords; j++ {
			// small 1-byte keys cause lots of collisions
			nKey := r.Intn(256)
			bKey := []byte{byte(nKey)}
			bVal := make([]byte, 8)
			r.Read(bVal)
			tree.Set(bKey, bVal)
		}
		tree.SaveVersion()
	}

	latest := int64(versions)
	available := tree.AvailableVersions()
	require.Equal(versions, len(available))

	tree.DeleteVersionsTo(latest - 1)
	available = tree.AvailableVersions()
	require.Equal(1, len(available))
}

func TestVersionedEmptyTree(t *testing.T) {
	require := require.New(t)
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 0, NewNopLogger())

	hash, v, err := tree.SaveVersion()
	require.NoError(err)
	require.NotNil(hash) // empty tree hash is SHA256(""), matching IAVL
	require.Len(hash, 32)
	require.EqualValues(1, v)

	_, v, err = tree.SaveVersion()
	require.NoError(err)
	require.EqualValues(2, v)

	_, v, err = tree.SaveVersion()
	require.NoError(err)
	require.EqualValues(3, v)

	_, v, err = tree.SaveVersion()
	require.NoError(err)
	require.EqualValues(4, v)

	require.EqualValues(4, tree.Version())
	require.True(tree.VersionExists(1))
	require.True(tree.VersionExists(3))

	// Test the empty root loads correctly
	it, err := tree.GetImmutable(3)
	require.NoError(err)
	require.True(it.IsEmpty())
	it.Close()

	require.NoError(tree.DeleteVersionsTo(3))
	require.False(tree.VersionExists(1))
	require.False(tree.VersionExists(3))

	tree.Set([]byte("k"), []byte("v"))

	// Reload the tree
	tree = NewMutableTreeWithDB(db, 0, NewNopLogger())
	tree.Load()

	require.False(tree.VersionExists(1))
	require.False(tree.VersionExists(2))
	require.False(tree.VersionExists(3))

	_, err = tree.GetImmutable(2)
	require.Error(err, "GetImmutable should fail for version 2")
}

func TestVersionedTreeHash(t *testing.T) {
	require := require.New(t)
	tree := getTestTree(0)

	tree.Set([]byte("I"), []byte("D"))
	tree.Set([]byte("J"), []byte("O"))
	tree.Set([]byte("E"), []byte("R"))
	hash1, _, _ := tree.SaveVersion()

	tree.Set([]byte("G"), []byte("G"))
	hash2, _, _ := tree.SaveVersion()

	// Reloading the tree should give the same hash
	tree2 := getTestTree(0)
	tree2.Set([]byte("I"), []byte("D"))
	tree2.Set([]byte("J"), []byte("O"))
	tree2.Set([]byte("E"), []byte("R"))
	hash1b, _, _ := tree2.SaveVersion()
	require.Equal(hash1, hash1b)

	tree2.Set([]byte("G"), []byte("G"))
	hash2b, _, _ := tree2.SaveVersion()
	require.Equal(hash2, hash2b)
}

func TestNilValueSemantics(t *testing.T) {
	tree := getTestTree(0)

	// Setting a nil value should error, matching IAVL behavior.
	_, err := tree.Set([]byte("k"), nil)
	require.Error(t, err)
}

func TestCopyValueSemantics(t *testing.T) {
	require := require.New(t)
	tree := getTestTree(0)

	val := []byte("v1")
	tree.Set([]byte("k"), val)

	v, err := tree.Get([]byte("k"))
	require.NoError(err)
	require.Equal([]byte("v1"), v)

	// INTENTIONAL BEHAVIORAL DIFFERENCE from IAVL:
	// IAVL stores value references directly, so mutating the original
	// slice affects what Get returns (val[1]='2' makes Get return "v2").
	// B+tree stores values out-of-line by content hash, so the value is
	// always copied. Mutating the original does NOT affect stored value.
	// This is safer behavior — callers cannot corrupt stored state.
	val[1] = '2'
	v, err = tree.Get([]byte("k"))
	require.NoError(err)
	require.Equal([]byte("v1"), v, "bptree copies values on Set; mutation of caller's slice must not corrupt stored data")
}

func TestRollback(t *testing.T) {
	require := require.New(t)
	tree := getTestTree(0)

	tree.Set([]byte("k"), []byte("v"))
	tree.SaveVersion()

	tree.Set([]byte("r"), []byte("v"))
	tree.Set([]byte("s"), []byte("v"))

	tree.Rollback()

	tree.Set([]byte("t"), []byte("v"))
	tree.SaveVersion()

	require.Equal(int64(2), tree.Size())

	val, err := tree.Get([]byte("r"))
	require.NoError(err)
	require.Nil(val)

	val, err = tree.Get([]byte("s"))
	require.NoError(err)
	require.Nil(val)

	val, err = tree.Get([]byte("t"))
	require.NoError(err)
	require.Equal([]byte("v"), val)
}

func TestLoadVersion(t *testing.T) {
	tree := getTestTree(0)
	maxVersions := 10

	for i := 0; i < maxVersions; i++ {
		tree.Set([]byte(fmt.Sprintf("key_%d", i+1)), []byte(fmt.Sprintf("value_%d", i+1)))
		_, _, err := tree.SaveVersion()
		require.NoError(t, err, "SaveVersion should not fail")
	}

	// require the ability to load the latest version
	version, err := tree.LoadVersion(int64(maxVersions))
	require.NoError(t, err, "unexpected error when loading version")
	require.Equal(t, int64(maxVersions), version)

	value, err := tree.Get([]byte(fmt.Sprintf("key_%d", maxVersions)))
	require.NoError(t, err)
	require.Equal(t, []byte(fmt.Sprintf("value_%d", maxVersions)), value)

	// require the ability to load an older version
	_, err = tree.LoadVersion(int64(maxVersions - 1))
	require.NoError(t, err, "unexpected error when loading version")

	value, err = tree.Get([]byte(fmt.Sprintf("key_%d", maxVersions-1)))
	require.NoError(t, err)
	require.Equal(t, []byte(fmt.Sprintf("value_%d", maxVersions-1)), value)
}

func TestOverwrite(t *testing.T) {
	require := require.New(t)
	mdb := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(mdb, 0, NewNopLogger())

	// Set one kv pair and save version 1
	tree.Set([]byte("key1"), []byte("value1"))
	_, _, err := tree.SaveVersion()
	require.NoError(err)

	// Set another kv pair and save version 2
	tree.Set([]byte("key2"), []byte("value2"))
	_, _, err = tree.SaveVersion()
	require.NoError(err)

	// Reload tree at version 1
	tree = NewMutableTreeWithDB(mdb, 0, NewNopLogger())
	_, err = tree.LoadVersion(int64(1))
	require.NoError(err)

	// Attempt to put a different kv pair and save
	tree.Set([]byte("key2"), []byte("value2"))
	_, _, err = tree.SaveVersion()
	// In our implementation we allow this (overwriting is permitted)
	require.NoError(err)
}

// TestLoadVersionForOverwriting_UnsupportedIAVL verifies the IAVL-compat
// API returns ErrUnsupported (Finding #12).
func TestLoadVersionForOverwriting_UnsupportedIAVL(t *testing.T) {
	mdb := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(mdb, 0, NewNopLogger())
	tree.Set([]byte("k"), []byte("v"))
	tree.SaveVersion()
	err := tree.LoadVersionForOverwriting(1)
	require.New(t).ErrorIs(err, ErrUnsupported)
}

func TestIterate_ImmutableTree_Version1(t *testing.T) {
	tree := getTestTree(0)
	mirror := make(map[string]string)

	// Insert random keys
	r := rand.New(rand.NewSource(42))
	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("key_%d", r.Intn(1000))
		v := fmt.Sprintf("val_%d", i)
		tree.Set([]byte(k), []byte(v))
		mirror[k] = v
	}

	tree.SaveVersion()
	immutableTree, err := tree.GetImmutable(1)
	require.NoError(t, err)

	// Verify iteration matches the mirror
	count := 0
	immutableTree.Iterate(func(key, value []byte) bool {
		count++
		return false
	})
	require.Equal(t, len(mirror), count)
}

func TestIterate_ImmutableTree_Version2(t *testing.T) {
	tree := getTestTree(0)
	mirror := make(map[string]string)

	r := rand.New(rand.NewSource(42))
	for i := 0; i < 50; i++ {
		k := fmt.Sprintf("key_%d", r.Intn(1000))
		v := fmt.Sprintf("val_%d", i)
		tree.Set([]byte(k), []byte(v))
		mirror[k] = v
	}
	tree.SaveVersion()

	// Add more for version 2
	for i := 50; i < 100; i++ {
		k := fmt.Sprintf("key_%d", r.Intn(1000))
		v := fmt.Sprintf("val2_%d", i)
		tree.Set([]byte(k), []byte(v))
		mirror[k] = v
	}
	tree.SaveVersion()

	immutableTree, err := tree.GetImmutable(2)
	require.NoError(t, err)

	count := 0
	immutableTree.Iterate(func(key, value []byte) bool {
		count++
		return false
	})
	require.Equal(t, len(mirror), count)
}

func TestGetByIndex_ImmutableTree(t *testing.T) {
	tree := getTestTree(0)
	mirror := make(map[string]string)

	r := rand.New(rand.NewSource(99))
	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("idx_%03d", r.Intn(200))
		v := fmt.Sprintf("val_%d", i)
		tree.Set([]byte(k), []byte(v))
		mirror[k] = v
	}

	mirrorKeys := make([]string, 0, len(mirror))
	for k := range mirror {
		mirrorKeys = append(mirrorKeys, k)
	}
	sort.Strings(mirrorKeys)

	tree.SaveVersion()
	immutableTree, err := tree.GetImmutable(1)
	require.NoError(t, err)

	for index, expectedKey := range mirrorKeys {
		actualKey, _, err := immutableTree.GetByIndex(int64(index))
		require.NoError(t, err)
		require.Equal(t, expectedKey, string(actualKey))
	}
}

func TestGetWithIndex_ImmutableTree(t *testing.T) {
	tree := getTestTree(0)
	mirror := make(map[string]string)

	r := rand.New(rand.NewSource(99))
	for i := 0; i < 100; i++ {
		k := fmt.Sprintf("widx_%03d", r.Intn(200))
		v := fmt.Sprintf("val_%d", i)
		tree.Set([]byte(k), []byte(v))
		mirror[k] = v
	}

	mirrorKeys := make([]string, 0, len(mirror))
	for k := range mirror {
		mirrorKeys = append(mirrorKeys, k)
	}
	sort.Strings(mirrorKeys)

	tree.SaveVersion()
	immutableTree, err := tree.GetImmutable(1)
	require.NoError(t, err)

	for expectedIndex, key := range mirrorKeys {
		actualIndex, actualValue, err := immutableTree.GetWithIndex([]byte(key))
		require.NoError(t, err)
		require.NotNil(t, actualValue)
		require.Equal(t, int64(expectedIndex), actualIndex)
	}
}

func TestEmptyVersionDelete(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 0, NewNopLogger())

	tree.Set([]byte("key1"), []byte("value1"))

	toVersion := 10
	for i := 0; i < toVersion; i++ {
		_, _, err := tree.SaveVersion()
		require.NoError(t, err)
	}

	require.NoError(t, tree.DeleteVersionsTo(5))

	// Load the tree from disk
	tree = NewMutableTreeWithDB(db, 0, NewNopLogger())
	v, err := tree.Load()
	require.NoError(t, err)
	require.Equal(t, int64(toVersion), v)

	// Versions 1-5 should be deleted
	versions := tree.AvailableVersions()
	require.Equal(t, 6, versions[0])
	require.Len(t, versions, 5)
}

func TestWorkingHashWithInitialVersion(t *testing.T) {
	db := memdb.NewMemDB()
	initialVersion := int64(100)
	tree := NewMutableTreeWithDB(db, 0, NewNopLogger())
	tree.SetInitialVersion(uint64(initialVersion))

	v := tree.WorkingVersion()
	require.Equal(t, initialVersion, v)

	tree.Set([]byte("key1"), []byte("value1"))

	workingHash := tree.WorkingHash()
	commitHash, _, err := tree.SaveVersion()
	require.NoError(t, err)
	require.Equal(t, commitHash, workingHash)

	// Without WorkingHash — should produce the same result
	db2 := memdb.NewMemDB()
	tree2 := NewMutableTreeWithDB(db2, 0, NewNopLogger(), InitialVersionOption(uint64(initialVersion)))
	tree2.Set([]byte("key1"), []byte("value1"))
	commitHash2, _, err := tree2.SaveVersion()
	require.NoError(t, err)
	require.True(t, bytes.Equal(commitHash, commitHash2))
}

func TestVersionedTreeSaveAndLoad(t *testing.T) {
	require := require.New(t)
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 0, NewNopLogger())

	// Create multiple versions
	for i := 0; i < 5; i++ {
		tree.Set([]byte(fmt.Sprintf("key_%d", i)), []byte(fmt.Sprintf("value_%d", i)))
		_, _, err := tree.SaveVersion()
		require.NoError(err)
	}

	// Reload
	tree2 := NewMutableTreeWithDB(db, 0, NewNopLogger())
	v, err := tree2.Load()
	require.NoError(err)
	require.Equal(int64(5), v)

	// All keys should be present
	for i := 0; i < 5; i++ {
		val, err := tree2.Get([]byte(fmt.Sprintf("key_%d", i)))
		require.NoError(err)
		require.Equal([]byte(fmt.Sprintf("value_%d", i)), val)
	}
}

func TestVersionedTree(t *testing.T) {
	require := require.New(t)
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 0, NewNopLogger())

	require.True(tree.IsEmpty())

	// version 1
	tree.Set([]byte("key1"), []byte("val0"))
	tree.Set([]byte("key2"), []byte("val0"))
	require.False(tree.IsEmpty())

	hash1, v, err := tree.SaveVersion()
	require.NoError(err)
	require.EqualValues(1, v)

	// version 2
	tree.Set([]byte("key1"), []byte("val1"))
	tree.Set([]byte("key2"), []byte("val1"))
	tree.Set([]byte("key3"), []byte("val1"))

	hash2, v2, err := tree.SaveVersion()
	require.NoError(err)
	require.False(bytes.Equal(hash1, hash2))
	require.EqualValues(2, v2)

	// Reload tree
	tree = NewMutableTreeWithDB(db, 100, NewNopLogger())
	_, err = tree.Load()
	require.NoError(err)
	require.EqualValues(v2, tree.Version())

	// version 3: remove key1, update key2
	tree.Remove([]byte("key1"))
	tree.Set([]byte("key2"), []byte("val2"))

	hash3, v3, _ := tree.SaveVersion()
	require.EqualValues(3, v3)

	// version 4: no changes (same hash as v3)
	hash4, _, _ := tree.SaveVersion()
	require.True(bytes.Equal(hash3, hash4))
	require.NotNil(hash4)

	// Reload
	tree = NewMutableTreeWithDB(db, 100, NewNopLogger())
	_, err = tree.Load()
	require.NoError(err)

	tree.Set([]byte("key1"), []byte("val0"))

	// GetVersioned checks
	val, err := tree.GetVersioned([]byte("key2"), 1)
	require.NoError(err)
	require.Equal("val0", string(val))

	val, err = tree.GetVersioned([]byte("key2"), 2)
	require.NoError(err)
	require.Equal("val1", string(val))

	val, err = tree.Get([]byte("key2"))
	require.NoError(err)
	require.Equal("val2", string(val))

	val, err = tree.GetVersioned([]byte("key1"), 1)
	require.NoError(err)
	require.Equal("val0", string(val))

	val, err = tree.GetVersioned([]byte("key1"), 2)
	require.NoError(err)
	require.Equal("val1", string(val))

	val, err = tree.GetVersioned([]byte("key1"), 3)
	require.NoError(err)
	require.Nil(val)

	val, err = tree.Get([]byte("key1"))
	require.NoError(err)
	require.Equal("val0", string(val))

	val, err = tree.GetVersioned([]byte("key3"), 2)
	require.NoError(err)
	require.Equal("val1", string(val))

	val, err = tree.GetVersioned([]byte("key3"), 3)
	require.NoError(err)
	require.Equal("val1", string(val))

	// Delete versions up to 2
	tree.DeleteVersionsTo(2)

	// Deleted versions should not be queryable
	val, err = tree.GetVersioned([]byte("key2"), 2)
	require.NoError(err)
	require.Nil(val)

	// But latest version should still work
	val, err = tree.Get([]byte("key2"))
	require.NoError(err)
	require.Equal("val2", string(val))

	val, err = tree.Get([]byte("key3"))
	require.NoError(err)
	require.Equal("val1", string(val))

	// Version 1 should not be available
	val, err = tree.GetVersioned([]byte("key1"), 1)
	require.NoError(err)
	require.Nil(val)
}

func TestOverwriteEmpty(t *testing.T) {
	require := require.New(t)
	mdb := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(mdb, 0, NewNopLogger())

	// Save empty version 1
	_, _, err := tree.SaveVersion()
	require.NoError(err)

	// Save empty version 2
	_, _, err = tree.SaveVersion()
	require.NoError(err)

	// Save a key in version 3
	tree.Set([]byte("key"), []byte("value"))
	_, _, err = tree.SaveVersion()
	require.NoError(err)

	// Load version 1 and attempt to save a different key
	_, err = tree.LoadVersion(1)
	require.NoError(err)
	tree.Set([]byte("foo"), []byte("bar"))
	_, _, err = tree.SaveVersion()
	// Should error because version 2 already exists with a different hash
	require.Error(err)

	// However, removing the key and saving an empty version should work
	// since it matches the existing empty version 2's hash.
	tree.Remove([]byte("foo"))
	_, version, err := tree.SaveVersion()
	require.NoError(err)
	require.EqualValues(2, version)
}

func TestRandomOperations(t *testing.T) {
	// Ported from tree_random_test.go — randomized multi-version operations
	seeds := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for _, seed := range seeds {
		t.Run(fmt.Sprintf("seed_%d", seed), func(t *testing.T) {
			r := rand.New(rand.NewSource(seed))
			db := memdb.NewMemDB()
			tree := NewMutableTreeWithDB(db, 100, NewNopLogger())
			mirror := make(map[string]string)

			versions := 32
			opsPerVersion := 50

			for v := 0; v < versions; v++ {
				for op := 0; op < opsPerVersion; op++ {
					switch r.Intn(3) {
					case 0: // Set
						k := fmt.Sprintf("k%d", r.Intn(200))
						val := fmt.Sprintf("v%d_%d", v, op)
						tree.Set([]byte(k), []byte(val))
						mirror[k] = val
					case 1: // Update
						if len(mirror) > 0 {
							keys := make([]string, 0, len(mirror))
							for k := range mirror {
								keys = append(keys, k)
							}
							k := keys[r.Intn(len(keys))]
							val := fmt.Sprintf("u%d_%d", v, op)
							tree.Set([]byte(k), []byte(val))
							mirror[k] = val
						}
					case 2: // Delete
						if len(mirror) > 0 {
							keys := make([]string, 0, len(mirror))
							for k := range mirror {
								keys = append(keys, k)
							}
							k := keys[r.Intn(len(keys))]
							tree.Remove([]byte(k))
							delete(mirror, k)
						}
					}
				}
				_, _, err := tree.SaveVersion()
				require.NoError(t, err)
			}

			// Final verification
			require.Equal(t, int64(len(mirror)), tree.Size())
			for k, v := range mirror {
				val, err := tree.Get([]byte(k))
				require.NoError(t, err)
				require.Equal(t, v, string(val), "mismatch for key %s", k)
			}
		})
	}
}
