package bptree

// Ported from tm2/pkg/iavl/iterator_test.go
// Skips fast-iterator-specific and unsaved-fast-iterator tests.

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIterator_NewIterator_NilTree_Failure(t *testing.T) {
	tree := NewMutableTreeMem()
	// Empty tree — iterator should be invalid
	itr, err := tree.Iterator(nil, nil, true)
	require.NoError(t, err)
	require.False(t, itr.Valid())
	itr.Close()
}

func TestIterator_Empty_Invalid(t *testing.T) {
	tree := getTestTree(0)
	// Insert a-z
	for c := byte('a'); c <= byte('z'); c++ {
		tree.Set([]byte{c}, []byte{c})
	}
	// Range [a, a) = empty
	itr, err := tree.Iterator([]byte("a"), []byte("a"), true)
	require.NoError(t, err)
	require.False(t, itr.Valid())
	itr.Close()
}

func TestIterator_Basic_Ranged_Ascending_Success(t *testing.T) {
	tree := getTestTree(0)
	for c := byte('a'); c <= byte('z'); c++ {
		tree.Set([]byte{c}, []byte{c})
	}
	itr, _ := tree.Iterator([]byte("e"), []byte("w"), true)
	defer itr.Close()

	var keys []string
	for itr.Valid() {
		keys = append(keys, string(itr.Key()))
		itr.Next()
	}
	// e through v (w exclusive)
	require.Equal(t, 18, len(keys)) // e,f,g,...,v
	require.Equal(t, "e", keys[0])
	require.Equal(t, "v", keys[len(keys)-1])
	require.True(t, sort.StringsAreSorted(keys))
}

func TestIterator_Basic_Ranged_Descending_Success(t *testing.T) {
	tree := getTestTree(0)
	for c := byte('a'); c <= byte('z'); c++ {
		tree.Set([]byte{c}, []byte{c})
	}
	itr, _ := tree.Iterator([]byte("e"), []byte("w"), false)
	defer itr.Close()

	var keys []string
	for itr.Valid() {
		keys = append(keys, string(itr.Key()))
		itr.Next()
	}
	require.Equal(t, 18, len(keys))
	require.Equal(t, "v", keys[0])
	require.Equal(t, "e", keys[len(keys)-1])
}

func TestIterator_Basic_Full_Ascending_Success(t *testing.T) {
	tree := getTestTree(0)
	for c := byte('a'); c <= byte('z'); c++ {
		tree.Set([]byte{c}, []byte{c})
	}
	itr, _ := tree.Iterator(nil, nil, true)
	defer itr.Close()

	var keys []string
	for itr.Valid() {
		keys = append(keys, string(itr.Key()))
		itr.Next()
	}
	require.Equal(t, 26, len(keys))
	require.Equal(t, "a", keys[0])
	require.Equal(t, "z", keys[25])
	require.True(t, sort.StringsAreSorted(keys))
}

func TestIterator_Basic_Full_Descending_Success(t *testing.T) {
	tree := getTestTree(0)
	for c := byte('a'); c <= byte('z'); c++ {
		tree.Set([]byte{c}, []byte{c})
	}
	itr, _ := tree.Iterator(nil, nil, false)
	defer itr.Close()

	var keys []string
	for itr.Valid() {
		keys = append(keys, string(itr.Key()))
		itr.Next()
	}
	require.Equal(t, 26, len(keys))
	require.Equal(t, "z", keys[0])
	require.Equal(t, "a", keys[25])
}

func TestIterator_WithDelete_Full_Ascending_Success(t *testing.T) {
	tree := getTestTree(0)
	for c := byte('a'); c <= byte('z'); c++ {
		tree.Set([]byte{c}, []byte{c})
	}
	tree.SaveVersion()

	// Delete every other key
	for c := byte('b'); c <= byte('z'); c += 2 {
		tree.Remove([]byte{c})
	}
	tree.SaveVersion()

	itr, _ := tree.Iterator(nil, nil, true)
	defer itr.Close()

	var keys []string
	for itr.Valid() {
		keys = append(keys, string(itr.Key()))
		itr.Next()
	}
	require.Equal(t, 13, len(keys)) // a,c,e,g,i,k,m,o,q,s,u,w,y
	require.Equal(t, "a", keys[0])
	require.Equal(t, "y", keys[len(keys)-1])
	require.True(t, sort.StringsAreSorted(keys))
}
