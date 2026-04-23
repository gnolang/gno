package bptree

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"testing"
)

func TestEdge_RollbackAfterMerge_COWIntegrity(t *testing.T) {
	// The bug: fixUnderflow didn't clone the merge-left sibling for in-memory
	// trees, corrupting lastSaved. This test verifies the fix.
	tree := NewMutableTreeMem()

	// Build a tree with 3+ leaves
	for i := 0; i < 49; i++ {
		tree.Set(fmt.Appendf(nil, "cow%03d", i), []byte("v"))
	}
	tree.SaveVersion()

	savedSize := tree.Size()
	var savedKeys []string
	tree.Iterate(func(k, v []byte) bool {
		savedKeys = append(savedKeys, string(k))
		return false
	})

	// Remove keys to trigger merge-left
	for i := 10; i < 30; i++ {
		tree.Remove(fmt.Appendf(nil, "cow%03d", i))
	}

	// Rollback should restore exactly the saved state
	tree.Rollback()

	if tree.Size() != savedSize {
		t.Fatalf("after rollback: size=%d, want %d", tree.Size(), savedSize)
	}

	var afterKeys []string
	tree.Iterate(func(k, v []byte) bool {
		afterKeys = append(afterKeys, string(k))
		return false
	})

	if len(afterKeys) != len(savedKeys) {
		t.Fatalf("after rollback: %d keys, want %d", len(afterKeys), len(savedKeys))
	}
	for i := range savedKeys {
		if savedKeys[i] != afterKeys[i] {
			t.Fatalf("key mismatch at %d: %s != %s", i, savedKeys[i], afterKeys[i])
		}
	}
}

func TestEdge_RollbackThenMutate(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "rm%03d", i), []byte("v1"))
	}
	tree.SaveVersion()

	// Mutate
	tree.Set([]byte("rm999"), []byte("new"))
	tree.Remove([]byte("rm000"))

	// Rollback
	tree.Rollback()

	// Mutate again after rollback — COW must clone the shared root
	tree.Set([]byte("rm888"), []byte("post-rollback"))
	if tree.Size() != 51 {
		t.Fatalf("size after rollback+set = %d, want 51", tree.Size())
	}
	has, _ := tree.Has([]byte("rm000"))
	if !has {
		t.Fatalf("rm000 should exist (rollback restored it)")
	}
	has, _ = tree.Has([]byte("rm888"))
	if !has {
		t.Fatalf("rm888 should exist (set after rollback)")
	}
}

func TestEdge_90_10_RightChildUnderflow(t *testing.T) {
	tree := NewMutableTreeMem()
	// Sequential inserts trigger 90/10 splits: left=31, right=2
	for i := 0; i < B+1; i++ {
		tree.Set(fmt.Appendf(nil, "ru%04d", i), []byte("v"))
	}
	if tree.Height() < 1 {
		t.Fatalf("need at least 2 leaves")
	}

	// The right child has only 2 keys. Remove one → underflow (1 < 16).
	tree.Remove(fmt.Appendf(nil, "ru%04d", B))
	if tree.Size() != int64(B) {
		t.Fatalf("size = %d, want %d", tree.Size(), B)
	}

	// Verify all remaining keys
	for i := 0; i < B; i++ {
		has, _ := tree.Has(fmt.Appendf(nil, "ru%04d", i))
		if !has {
			t.Fatalf("ru%04d not found after right-child underflow", i)
		}
	}

	var keys []string
	tree.Iterate(func(k, v []byte) bool {
		keys = append(keys, string(k))
		return false
	})
	if !sort.StringsAreSorted(keys) {
		t.Fatalf("keys not sorted after 90/10 underflow")
	}
}

func TestEdge_ExactMinKeys_AllSiblingsCantSpare(t *testing.T) {
	tree := NewMutableTreeMem()
	// Insert enough keys to create multiple leaves, then remove to
	// bring siblings to exactly MinKeys each.
	n := B * 3
	for i := 0; i < n; i++ {
		tree.Set(fmt.Appendf(nil, "ms%04d", i), []byte("v"))
	}

	// Remove keys from the middle until merges cascade
	removed := 0
	for i := n / 3; i < 2*n/3; i++ {
		_, found, _ := tree.Remove(fmt.Appendf(nil, "ms%04d", i))
		if found {
			removed++
		}
	}

	// Verify integrity
	var keys []string
	tree.Iterate(func(k, v []byte) bool {
		keys = append(keys, string(k))
		return false
	})
	if int64(len(keys)) != tree.Size() {
		t.Fatalf("iterate count %d != size %d", len(keys), tree.Size())
	}
	if !sort.StringsAreSorted(keys) {
		t.Fatalf("keys not sorted")
	}
}

func TestEdge_NilValue(t *testing.T) {
	tree := NewMutableTreeMem()
	// Set with nil value — should error (matching IAVL behavior)
	_, err := tree.Set([]byte("k"), nil)
	if err == nil {
		t.Fatalf("Set nil value should error")
	}
}

func TestEdge_InsertRemoveInsert_FullCycle(t *testing.T) {
	tree := NewMutableTreeMem()
	n := B * 4 // 128

	// Insert all
	for i := 0; i < n; i++ {
		tree.Set(fmt.Appendf(nil, "cy%04d", i), []byte("v"))
	}
	if tree.Size() != int64(n) {
		t.Fatalf("after insert: size=%d", tree.Size())
	}

	// Remove all
	for i := 0; i < n; i++ {
		tree.Remove(fmt.Appendf(nil, "cy%04d", i))
	}
	if !tree.IsEmpty() {
		t.Fatalf("should be empty after removing all")
	}

	// Insert again — tree should work from scratch
	for i := 0; i < n; i++ {
		tree.Set(fmt.Appendf(nil, "cy%04d", i), []byte("v2"))
	}
	if tree.Size() != int64(n) {
		t.Fatalf("after re-insert: size=%d", tree.Size())
	}

	var keys []string
	tree.Iterate(func(k, v []byte) bool {
		keys = append(keys, string(k))
		return false
	})
	if !sort.StringsAreSorted(keys) || len(keys) != n {
		t.Fatalf("after re-insert: %d keys, sorted=%v", len(keys), sort.StringsAreSorted(keys))
	}
}

func TestEdge_GetByIndex_LastKey(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 0; i < 100; i++ {
		tree.Set(fmt.Appendf(nil, "lk%04d", i), []byte("v"))
	}

	k, _, err := tree.GetByIndex(tree.Size() - 1)
	if err != nil {
		t.Fatalf("GetByIndex(last): %v", err)
	}
	if string(k) != "lk0099" {
		t.Fatalf("last key = %s, want lk0099", k)
	}
}

func TestEdge_Iterator_DescendingEndBeforeFirstKey(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 10; i < 20; i++ {
		tree.Set(fmt.Appendf(nil, "de%04d", i), []byte("v"))
	}

	// end = "de0005" which is before all keys
	itr, _ := tree.Iterator(nil, []byte("de0005"), false)
	defer itr.Close()
	if itr.Valid() {
		t.Fatalf("descending with end before all keys should be empty")
	}
}

func TestEdge_ConcurrentImmutableReads(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 0; i < 200; i++ {
		tree.Set(fmt.Appendf(nil, "cc%04d", i), []byte("v"))
	}
	imm := tree.Snapshot(1)

	var wg sync.WaitGroup
	errs := make(chan error, 10)

	// Multiple goroutines reading concurrently
	for g := 0; g < 5; g++ {
		wg.Add(1)
		go func(gid int) {
			defer wg.Done()
			// Get
			for i := 0; i < 200; i++ {
				_, err := imm.Has(fmt.Appendf(nil, "cc%04d", i))
				if err != nil {
					errs <- fmt.Errorf("goroutine %d Has error: %w", gid, err)
					return
				}
			}
			// Iterate
			count := 0
			imm.Iterate(func(k, v []byte) bool {
				count++
				return false
			})
			if count != 200 {
				errs <- fmt.Errorf("goroutine %d iterate count=%d", gid, count)
			}
			// GetByIndex
			for i := int64(0); i < 10; i++ {
				_, _, err := imm.GetByIndex(i)
				if err != nil {
					errs <- fmt.Errorf("goroutine %d GetByIndex error: %w", gid, err)
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

func TestEdge_GetByIndex_AfterInnerMerge_SizeConsistency(t *testing.T) {
	tree := NewMutableTreeMem()
	// Build a height-2 tree
	n := 1100
	for i := 0; i < n; i++ {
		tree.Set(fmt.Appendf(nil, "gm%05d", i), []byte("v"))
	}
	if tree.Height() < 2 {
		t.Skipf("need height >= 2, got %d", tree.Height())
	}

	// Remove enough to trigger inner merges
	for i := 0; i < 800; i++ {
		tree.Remove(fmt.Appendf(nil, "gm%05d", i))
	}

	// GetByIndex for every valid index must work without panic
	remaining := tree.Size()
	var prev string
	for i := int64(0); i < remaining; i++ {
		k, _, err := tree.GetByIndex(i)
		if err != nil {
			t.Fatalf("GetByIndex(%d) after inner merge: %v", i, err)
		}
		if string(k) <= prev && prev != "" {
			t.Fatalf("order broken at %d: %q <= %q", i, k, prev)
		}
		prev = string(k)
	}
}

func TestEdge_SaveVersion_ThenMerge_ThenRollback(t *testing.T) {
	// DB-backed version of the COW merge rollback test
	tree := newTestTree(t) // uses memdb

	for i := 0; i < 60; i++ {
		tree.Set(fmt.Appendf(nil, "sr%03d", i), []byte("v"))
	}
	hash1, _, _ := tree.SaveVersion()

	// Remove to trigger merges
	for i := 0; i < 30; i++ {
		tree.Remove(fmt.Appendf(nil, "sr%03d", i))
	}

	tree.Rollback()

	hash2 := tree.WorkingHash()
	if !bytes.Equal(hash1, hash2) {
		t.Fatalf("hash changed after rollback: %x != %x", hash1, hash2)
	}
	if tree.Size() != 60 {
		t.Fatalf("size after rollback = %d, want 60", tree.Size())
	}
}
