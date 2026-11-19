package bptree

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

func TestConcurrent_LazyLoadFromDB(t *testing.T) {
	// Save a DB-backed tree, reload it (childNodes nil), then
	// have multiple goroutines do concurrent Gets.
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	for i := range 200 {
		tree.Set(fmt.Appendf(nil, "clz%04d", i), fmt.Appendf(nil, "val%04d", i))
	}
	tree.SaveVersion()

	// Reload — root loaded, children are nil (lazy)
	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	tree2.Load()
	imm, err := tree2.GetImmutable(1)
	require.NoError(t, err)

	// Concurrent reads — exercises the getChild mutex
	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for g := range 10 {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			for i := range 200 {
				key := fmt.Appendf(nil, "clz%04d", i)
				has, err := imm.Has(key)
				if err != nil {
					errs <- fmt.Errorf("g%d Has(%s): %w", gid, key, err)
					return
				}
				if !has {
					errs <- fmt.Errorf("g%d: key %s not found", gid, key)
					return
				}
			}
		}(g)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
}

func TestGetChild_ErrorsOnDBError(t *testing.T) {
	// Create a tree, save it, then corrupt the DB so getChild fails
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	for i := range 50 {
		tree.Set(fmt.Appendf(nil, "panic%03d", i), []byte("v"))
	}
	tree.SaveVersion()

	// Reload to get lazy-loaded nodes
	tree2 := NewMutableTreeWithDB(db, 0, NewNopLogger()) // cache=0 so nothing cached
	tree2.Load()

	// Delete all node entries from the DB to simulate corruption
	prefix := []byte{PrefixNode}
	end := []byte{PrefixNode + 1}
	itr, _ := db.Iterator(prefix, end)
	var nodeKeys [][]byte
	for ; itr.Valid(); itr.Next() {
		nodeKeys = append(nodeKeys, append([]byte(nil), itr.Key()...))
	}
	itr.Close()
	for _, k := range nodeKeys {
		db.Delete(k)
	}

	// Any traversal returns an error (a single corrupt/missing record fails
	// the one operation; it must not kill the process).
	if _, err := tree2.Get([]byte("panic025")); err == nil {
		t.Fatal("expected error from getChild on DB error")
	}
}

func TestCreateExistenceProof_NilValueResolver(t *testing.T) {
	tree := newMemTree()
	tree.Set([]byte("a"), []byte("1"))

	// Create ImmutableTree directly without a valueResolver
	imm := NewImmutableTree(tree.root, 0)
	// valueResolver is nil

	_, err := imm.GetMembershipProof([]byte("a"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "value resolver")
}

func TestMutableTreeMem_GetReturnsActualValues(t *testing.T) {
	tree := newMemTree()
	tree.Set([]byte("key"), []byte("actual_value"))

	val, err := tree.Get([]byte("key"))
	require.NoError(t, err)
	require.Equal(t, []byte("actual_value"), val,
		"in-memory Get should return actual value, not 32-byte hash")

	// Also verify GetByIndex
	k, v, err := tree.GetByIndex(0)
	require.NoError(t, err)
	require.Equal(t, []byte("key"), k)
	require.Equal(t, []byte("actual_value"), v)

	// And GetWithIndex
	idx, v, err := tree.GetWithIndex([]byte("key"))
	require.NoError(t, err)
	require.Equal(t, int64(0), idx)
	require.Equal(t, []byte("actual_value"), v)

	// And Iterate
	var iterVal []byte
	tree.Iterate(func(key, value []byte) bool {
		iterVal = value
		return true
	})
	require.Equal(t, []byte("actual_value"), iterVal,
		"in-memory Iterate should return actual value")

	// And Remove returns the old value
	old, found, err := tree.Remove([]byte("key"))
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, []byte("actual_value"), old)
}
