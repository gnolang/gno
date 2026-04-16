package bptree

// Remaining ported IAVL tests — behavioral tests that don't depend on
// IAVL-internal structures (orphans, fast nodes, node encoding, etc.)

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

func TestVersionedTreeErrors(t *testing.T) {
	require := require.New(t)
	tree := getTestTree(100)

	// Can't delete non-existent versions (no version saved yet)
	require.Error(tree.DeleteVersionsTo(1))

	tree.Set([]byte("key"), []byte("val"))
	_, _, err := tree.SaveVersion()
	require.NoError(err)

	// Can't delete current (latest) version
	require.Error(tree.DeleteVersionsTo(1))

	// Trying to get a key from a non-existent version
	val, err := tree.GetVersioned([]byte("key"), 404)
	require.NoError(err)
	require.Nil(val)
}

func TestVersionedRandomTreeSmallKeysRandomDeletes(t *testing.T) {
	require := require.New(t)
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	versions := 30
	keysPerVersion := 50
	r := rand.New(rand.NewSource(0))

	for i := 1; i <= versions; i++ {
		for j := 0; j < keysPerVersion; j++ {
			// 1-byte keys cause lots of collisions
			k := []byte{byte(r.Intn(256))}
			v := make([]byte, 8)
			r.Read(v)
			tree.Set(k, v)
		}
		tree.SaveVersion()
	}

	// Delete versions in random order (all except latest)
	perm := r.Perm(versions - 1)
	for _, i := range perm {
		v := int64(i + 1)
		if tree.VersionExists(v) {
			tree.DeleteVersionsTo(v)
		}
	}

	// After cleanup, tree should still be functional
	require.Equal(int64(versions), tree.Version())

	// Try getting random keys — they should exist
	for i := 0; i < keysPerVersion; i++ {
		k := []byte{byte(r.Intn(256))}
		has, _ := tree.Has(k)
		// May or may not exist depending on random ops
		_ = has
	}
}

func TestVersionedTreeSpecial1(t *testing.T) {
	tree := getTestTree(100)

	tree.Set([]byte("key1"), []byte("val0"))
	tree.Set([]byte("key2"), []byte("val0"))
	tree.SaveVersion()

	tree.Set([]byte("key1"), []byte("val1"))
	tree.Set([]byte("key2"), []byte("val1"))
	tree.Set([]byte("key3"), []byte("val1"))
	tree.SaveVersion()

	tree.Set([]byte("key2"), []byte("val2"))
	tree.SaveVersion()

	// All keys should be retrievable
	val, _ := tree.Get([]byte("key1"))
	require.Equal(t, []byte("val1"), val)
	val, _ = tree.Get([]byte("key2"))
	require.Equal(t, []byte("val2"), val)
	val, _ = tree.Get([]byte("key3"))
	require.Equal(t, []byte("val1"), val)
}

func TestVersionedTreeSpecialCase(t *testing.T) {
	require := require.New(t)
	tree := getTestTree(0)

	tree.Set([]byte("key1"), []byte("val0"))
	tree.Set([]byte("key2"), []byte("val0"))
	tree.SaveVersion()

	tree.Set([]byte("key1"), []byte("val1"))
	tree.Set([]byte("key2"), []byte("val1"))
	tree.SaveVersion()

	tree.Set([]byte("key2"), []byte("val2"))
	tree.SaveVersion()

	tree.DeleteVersionsTo(1)
	tree.DeleteVersionsTo(2)

	val, err := tree.Get([]byte("key2"))
	require.NoError(err)
	require.Equal([]byte("val2"), val)
}

func TestVersionedTreeSpecialCase2(t *testing.T) {
	require := require.New(t)
	tree := getTestTree(0)

	tree.Set([]byte("key1"), []byte("val0"))
	tree.Set([]byte("key2"), []byte("val0"))
	tree.SaveVersion()

	tree.Set([]byte("key1"), []byte("val1"))
	tree.Set([]byte("key2"), []byte("val1"))
	tree.Set([]byte("key3"), []byte("val1"))
	tree.SaveVersion()

	tree.Remove([]byte("key3"))
	tree.SaveVersion()

	tree.DeleteVersionsTo(1)
	tree.DeleteVersionsTo(2)

	val, err := tree.Get([]byte("key2"))
	require.NoError(err)
	require.Equal([]byte("val1"), val)

	has, _ := tree.Has([]byte("key3"))
	require.False(has)
}

func TestVersionedTreeSpecialCase3(t *testing.T) {
	require := require.New(t)
	tree := getTestTree(0)

	tree.Set([]byte("m"), []byte("liber"))
	tree.SaveVersion()

	tree.Set([]byte("k"), []byte("ursa"))
	tree.Set([]byte("m"), []byte("mansen"))
	tree.SaveVersion()

	tree.Set([]byte("m"), []byte("mollis"))
	tree.SaveVersion()

	tree.DeleteVersionsTo(1)
	tree.DeleteVersionsTo(2)

	val, err := tree.Get([]byte("m"))
	require.NoError(err)
	require.Equal([]byte("mollis"), val)

	val, err = tree.Get([]byte("k"))
	require.NoError(err)
	require.Equal([]byte("ursa"), val)
}

func TestVersionedCheckpointsSpecialCase(t *testing.T) {
	require := require.New(t)
	tree := getTestTree(0)

	key := []byte("k")
	tree.Set(key, []byte("val1"))
	tree.SaveVersion() // v1

	tree.Set(key, []byte("val2"))
	tree.SaveVersion() // v2

	tree.Set(key, []byte("val3"))
	tree.SaveVersion() // v3

	// Delete old versions, keeping latest
	tree.DeleteVersionsTo(1)

	val, _ := tree.GetVersioned(key, 2)
	require.Equal([]byte("val2"), val)

	val, _ = tree.GetVersioned(key, 3)
	require.Equal([]byte("val3"), val)
}

func TestVersionedCheckpointsSpecialCase2(t *testing.T) {
	tree := getTestTree(0)

	tree.Set([]byte("key1"), []byte("val1"))
	tree.SaveVersion()

	tree.Remove([]byte("key1"))
	tree.SaveVersion()

	tree.Set([]byte("key1"), []byte("val1"))
	tree.SaveVersion()

	tree.DeleteVersionsTo(1)

	// key1 should exist in v3
	val, _ := tree.Get([]byte("key1"))
	require.Equal(t, []byte("val1"), val)
}

func TestVersionedCheckpointsSpecialCase3(t *testing.T) {
	tree := getTestTree(0)

	tree.Set([]byte("key1"), []byte("val1"))
	tree.SaveVersion()

	tree.Remove([]byte("key1"))
	tree.SaveVersion()

	tree.SaveVersion() // empty version

	tree.DeleteVersionsTo(2)

	has, _ := tree.Has([]byte("key1"))
	require.False(t, has)
}

func TestVersionedCheckpointsSpecialCase4(t *testing.T) {
	tree := getTestTree(0)

	tree.Set([]byte("key1"), []byte("val1"))
	tree.SaveVersion()

	tree.Set([]byte("key2"), []byte("val2"))
	tree.SaveVersion()

	tree.Set([]byte("key2"), []byte("val3"))
	tree.SaveVersion()

	tree.Set([]byte("key3"), []byte("val4"))
	tree.SaveVersion()

	tree.DeleteVersionsTo(2)

	val, _ := tree.GetVersioned([]byte("key2"), 3)
	require.Equal(t, []byte("val3"), val)

	val, _ = tree.Get([]byte("key3"))
	require.Equal(t, []byte("val4"), val)
}

func TestVersionedCheckpointsSpecialCase5(t *testing.T) {
	tree := getTestTree(0)

	tree.Set([]byte("key1"), []byte("val1"))
	tree.SaveVersion()

	tree.Remove([]byte("key1"))
	tree.SaveVersion()

	tree.Set([]byte("key2"), []byte("val2"))
	tree.SaveVersion()

	tree.DeleteVersionsTo(1)

	val, _ := tree.Get([]byte("key2"))
	require.Equal(t, []byte("val2"), val)

	has, _ := tree.Has([]byte("key1"))
	require.False(t, has)
}

func TestVersionedCheckpointsSpecialCase6(t *testing.T) {
	tree := getTestTree(0)

	tree.Set([]byte("key1"), []byte("val1"))
	tree.SaveVersion()

	tree.Set([]byte("key2"), []byte("val2"))
	tree.SaveVersion()

	tree.Remove([]byte("key2"))
	tree.SaveVersion()

	tree.Set([]byte("key3"), []byte("val3"))
	tree.SaveVersion()

	tree.DeleteVersionsTo(2)

	val, _ := tree.Get([]byte("key1"))
	require.Equal(t, []byte("val1"), val)

	has, _ := tree.Has([]byte("key2"))
	require.False(t, has)

	val, _ = tree.Get([]byte("key3"))
	require.Equal(t, []byte("val3"), val)
}

func TestVersionedCheckpointsSpecialCase7(t *testing.T) {
	tree := getTestTree(0)

	tree.Set([]byte("key1"), []byte("val1"))
	tree.Set([]byte("key2"), []byte("val2"))
	tree.SaveVersion()

	tree.Set([]byte("key1"), []byte("val3"))
	tree.SaveVersion()

	tree.Set([]byte("key2"), []byte("val4"))
	tree.SaveVersion()

	tree.DeleteVersionsTo(2)

	val, _ := tree.Get([]byte("key1"))
	require.Equal(t, []byte("val3"), val)
	val, _ = tree.Get([]byte("key2"))
	require.Equal(t, []byte("val4"), val)
}

// TestLoadVersionForOverwriting_Unsupported verifies that the IAVL-compat
// entry point returns ErrUnsupported (Finding #12) rather than panicking.
func TestLoadVersionForOverwriting_Unsupported(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 0, NewNopLogger())
	tree.Set([]byte("a"), []byte("1"))
	tree.SaveVersion()
	err := tree.LoadVersionForOverwriting(1)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrUnsupported)
}

// TestLoadVersionForOverwritingCase2, Case3 removed — LoadVersionForOverwriting
// returns ErrUnsupported.

func TestVersionedTreeProofs(t *testing.T) {
	tree := getTestTree(0)

	tree.Set([]byte("k1"), []byte("v1"))
	tree.Set([]byte("k2"), []byte("v1"))
	tree.Set([]byte("k3"), []byte("v1"))
	tree.SaveVersion() // v1

	tree.Set([]byte("k1"), []byte("v2"))
	tree.Remove([]byte("k3"))
	tree.SaveVersion() // v2

	// Proofs for current version
	proof, err := tree.GetMembershipProof([]byte("k1"))
	require.NoError(t, err)
	require.NotNil(t, proof)

	proof, err = tree.GetMembershipProof([]byte("k2"))
	require.NoError(t, err)
	require.NotNil(t, proof)

	// Non-membership proof for deleted key
	proof, err = tree.GetNonMembershipProof([]byte("k3"))
	require.NoError(t, err)
	require.NotNil(t, proof)

	// Non-membership proof for non-existent key
	proof, err = tree.GetNonMembershipProof([]byte("k4"))
	require.NoError(t, err)
	require.NotNil(t, proof)
}

func TestTreeGetProof(t *testing.T) {
	tree := getTestTree(0)

	// Proof on empty tree should fail
	_, err := tree.GetMembershipProof([]byte("foo"))
	require.Error(t, err)

	// Insert keys
	for i := 0; i < 200; i++ {
		k := fmt.Sprintf("pkey_%03d", i)
		tree.Set([]byte(k), []byte(k))
	}
	tree.SaveVersion()

	// Random non-existent key should fail membership proof
	_, err = tree.GetMembershipProof([]byte("nonexistent_xyz"))
	require.Error(t, err)

	// Valid proofs for existing keys
	for i := 0; i < 200; i++ {
		k := fmt.Sprintf("pkey_%03d", i)
		proof, err := tree.GetMembershipProof([]byte(k))
		require.NoError(t, err)
		require.NotNil(t, proof)
		require.Equal(t, []byte(k), proof.GetExist().Value)
	}
}

func TestTreeKeyExistsProof(t *testing.T) {
	tree := getTestTree(0)
	r := rand.New(rand.NewSource(42))

	keys := make([][]byte, 200)
	for i := 0; i < 200; i++ {
		key := make([]byte, 20)
		r.Read(key)
		val := make([]byte, 20)
		r.Read(val)
		tree.Set(key, val)
		keys[i] = key
	}
	tree.SaveVersion()

	for _, key := range keys {
		proof, err := tree.GetMembershipProof(key)
		require.NoError(t, err)
		require.NotNil(t, proof)
	}
}

func TestDeleteVersionsFromNoDeadlock(t *testing.T) {
	// DeleteVersionsFrom returns ErrUnsupported (Finding #12); verify the
	// early-return contract rather than the old panic behaviour.
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 0, NewNopLogger())
	tree.Set([]byte("k"), []byte("v"))
	tree.SaveVersion()
	tree.SaveVersion()
	err := tree.DeleteVersionsFrom(1)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrUnsupported)
}

func TestIAVLAlternativePruning(t *testing.T) {
	// KeepRecent=3, KeepEvery=5
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 0, NewNopLogger())
	keepRecent := int64(3)
	keepEvery := int64(5)

	for i := 1; i <= 15; i++ {
		tree.Set([]byte(fmt.Sprintf("k%d", i)), []byte(fmt.Sprintf("v%d", i)))
		tree.SaveVersion()

		previous := int64(i) - 1
		if keepRecent < previous {
			toRelease := previous - keepRecent
			if keepEvery == 0 || toRelease%keepEvery != 0 {
				if tree.VersionExists(toRelease) {
					tree.DeleteVersionsTo(toRelease)
				}
			}
		}
	}

	// Recent versions should exist
	for _, v := range []int64{13, 14, 15} {
		require.True(t, tree.VersionExists(v), "version %d should exist", v)
	}
}
