package bptree

// Coverage tests for gaps identified by review.
// These target specific untested code paths that could hide real bugs.

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// --- GetByIndex / GetWithIndex after DB reload ---

func TestCoverage_GetByIndexAfterReload(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	keys := make([]string, 100)
	for i := 0; i < 100; i++ {
		keys[i] = fmt.Sprintf("idx%05d", i)
		tree.Set([]byte(keys[i]), fmt.Appendf(nil, "val%05d", i))
	}
	tree.SaveVersion()

	// Reload from DB
	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	tree2.LoadVersion(1)

	// GetByIndex for every valid index
	for i := int64(0); i < tree2.Size(); i++ {
		k, v, err := tree2.GetByIndex(i)
		if err != nil {
			t.Fatalf("GetByIndex(%d): %v", i, err)
		}
		if k == nil || v == nil {
			t.Fatalf("GetByIndex(%d): nil key or value", i)
		}
	}
}

func TestCoverage_GetWithIndexAfterReload(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	for i := 0; i < 100; i++ {
		tree.Set(fmt.Appendf(nil, "wi%05d", i), fmt.Appendf(nil, "val%05d", i))
	}
	tree.SaveVersion()

	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	tree2.LoadVersion(1)

	for i := 0; i < 100; i++ {
		key := fmt.Appendf(nil, "wi%05d", i)
		idx, val, err := tree2.GetWithIndex(key)
		if err != nil {
			t.Fatalf("GetWithIndex(%q): %v", key, err)
		}
		if val == nil {
			t.Fatalf("GetWithIndex(%q): nil value", key)
		}
		if idx < 0 || idx >= tree2.Size() {
			t.Fatalf("GetWithIndex(%q): index %d out of range [0,%d)", key, idx, tree2.Size())
		}
	}
}

// --- Full Rollback → SaveVersion → Set → SaveVersion cycle ---

func TestCoverage_RollbackSaveSetSaveCycle(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	// V1
	tree.Set([]byte("a"), []byte("v1"))
	tree.SaveVersion()

	// Mutate, then rollback
	tree.Set([]byte("a"), []byte("rolled_back"))
	tree.Set([]byte("b"), []byte("rolled_back"))
	tree.Rollback()

	// V2: save the rolled-back state (same as V1)
	hash2, v2, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v2 != 2 {
		t.Fatalf("expected version 2, got %d", v2)
	}

	// V3: new mutations after rollback+save
	tree.Set([]byte("a"), []byte("v3"))
	tree.Set([]byte("c"), []byte("v3"))
	hash3, v3, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v3 != 3 {
		t.Fatalf("expected version 3, got %d", v3)
	}

	// Reload and verify
	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	tree2.LoadVersion(3)
	val, _ := tree2.Get([]byte("a"))
	if string(val) != "v3" {
		t.Fatalf("Get(a) = %q, want v3", val)
	}
	val, _ = tree2.Get([]byte("c"))
	if string(val) != "v3" {
		t.Fatalf("Get(c) = %q, want v3", val)
	}
	// "b" should not exist (was in rolled-back session)
	val, _ = tree2.Get([]byte("b"))
	if val != nil {
		t.Fatalf("Get(b) = %q, should not exist", val)
	}

	// Prune V1 and V2, verify value cleanup
	tree.DeleteVersionsTo(2)
	valCount := countDBValues(db)
	// V3 has keys a,c = 2 values. V1's "a"="v1" should be cleaned.
	if valCount != 2 {
		t.Fatalf("after prune: %d values, want 2", valCount)
	}

	_ = hash2
	_ = hash3
}

func TestCoverage_RollbackWithTier2Orphans(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	// V1: set a key
	tree.Set([]byte("k"), []byte("v1"))
	tree.SaveVersion()

	// Working session for V2: overwrite (creates Tier 2 orphan)
	tree.Set([]byte("k"), []byte("v2_never_saved"))

	// Rollback — Tier 2 orphan (V1's valueKey) should be discarded,
	// V1's value should remain intact
	tree.Rollback()

	// V1's value should still be accessible
	val, _ := tree.Get([]byte("k"))
	if string(val) != "v1" {
		t.Fatalf("after rollback: Get(k) = %q, want v1", val)
	}

	// DB should have exactly 1 value (v1's, not v2's)
	valCount := countDBValues(db)
	if valCount != 1 {
		t.Fatalf("after rollback: %d values, want 1", valCount)
	}
}

func TestCoverage_MultipleRollbacks(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	tree.Set([]byte("a"), []byte("v1"))
	tree.SaveVersion()

	// First rollback
	tree.Set([]byte("b"), []byte("gone1"))
	tree.Rollback()

	// Second rollback after more mutations
	tree.Set([]byte("c"), []byte("gone2"))
	tree.Rollback()

	if tree.Size() != 1 {
		t.Fatalf("after 2 rollbacks: size=%d, want 1", tree.Size())
	}
	val, _ := tree.Get([]byte("a"))
	if string(val) != "v1" {
		t.Fatalf("Get(a) = %q, want v1", val)
	}
	// Only V1's value should be in DB
	if countDBValues(db) != 1 {
		t.Fatalf("after 2 rollbacks: %d values, want 1", countDBValues(db))
	}
}

// --- Import error paths ---

func TestCoverage_ImportErrors(t *testing.T) {
	db := memdb.NewMemDB()

	t.Run("leaf_numkeys_too_large", func(t *testing.T) {
		tree := NewMutableTreeWithDB(db, 100, NewNopLogger())
		imp, _ := tree.Import(1)
		imp.Add(&ExportNode{Key: []byte("a"), Value: []byte("1"), Height: 0})
		err := imp.Add(&ExportNode{Height: -1, NumKeys: B + 1})
		if err == nil {
			t.Fatal("expected error for NumKeys > B")
		}
	})

	t.Run("leaf_buffer_underflow", func(t *testing.T) {
		tree := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger())
		imp, _ := tree.Import(1)
		// Send boundary marker with 5 keys but only 1 entry buffered
		imp.Add(&ExportNode{Key: []byte("a"), Value: []byte("1"), Height: 0})
		err := imp.Add(&ExportNode{Height: -1, NumKeys: 5})
		if err == nil {
			t.Fatal("expected error for buffer underflow")
		}
	})

	t.Run("inner_numkeys_too_large", func(t *testing.T) {
		tree := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger())
		imp, _ := tree.Import(1)
		// Build two leaves
		imp.Add(&ExportNode{Key: []byte("a"), Value: []byte("1"), Height: 0})
		imp.Add(&ExportNode{Height: -1, NumKeys: 1})
		imp.Add(&ExportNode{Key: []byte("b"), Value: []byte("2"), Height: 0})
		imp.Add(&ExportNode{Height: -1, NumKeys: 1})
		// Inner marker with too many keys
		err := imp.Add(&ExportNode{Height: 1, NumKeys: B, SeparatorKeys: make([][]byte, B)})
		if err == nil {
			t.Fatal("expected error for inner NumKeys >= B")
		}
	})

	t.Run("inner_stack_underflow", func(t *testing.T) {
		tree := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger())
		imp, _ := tree.Import(1)
		// Inner marker with 1 key (needs 2 children) but stack is empty
		err := imp.Add(&ExportNode{Height: 1, NumKeys: 1, SeparatorKeys: [][]byte{[]byte("b")}})
		if err == nil {
			t.Fatal("expected error for stack underflow")
		}
	})

	t.Run("separator_count_mismatch", func(t *testing.T) {
		tree := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger())
		imp, _ := tree.Import(1)
		imp.Add(&ExportNode{Key: []byte("a"), Value: []byte("1"), Height: 0})
		imp.Add(&ExportNode{Height: -1, NumKeys: 1})
		imp.Add(&ExportNode{Key: []byte("b"), Value: []byte("2"), Height: 0})
		imp.Add(&ExportNode{Height: -1, NumKeys: 1})
		// Inner with NumKeys=1 but 0 separator keys
		err := imp.Add(&ExportNode{Height: 1, NumKeys: 1, SeparatorKeys: [][]byte{}})
		if err == nil {
			t.Fatal("expected error for separator count mismatch")
		}
	})

	t.Run("commit_with_leftover_entries", func(t *testing.T) {
		tree := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger())
		imp, _ := tree.Import(1)
		imp.Add(&ExportNode{Key: []byte("a"), Value: []byte("1"), Height: 0})
		// Commit without a leaf boundary marker
		err := imp.Commit()
		if err == nil {
			t.Fatal("expected error for unbounded leaf entries")
		}
	})

	t.Run("commit_with_multiple_roots", func(t *testing.T) {
		tree := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger())
		imp, _ := tree.Import(1)
		// Two separate leaves (no inner to combine them)
		imp.Add(&ExportNode{Key: []byte("a"), Value: []byte("1"), Height: 0})
		imp.Add(&ExportNode{Height: -1, NumKeys: 1})
		imp.Add(&ExportNode{Key: []byte("b"), Value: []byte("2"), Height: 0})
		imp.Add(&ExportNode{Height: -1, NumKeys: 1})
		err := imp.Commit()
		if err == nil {
			t.Fatal("expected error for multiple roots on stack")
		}
	})
}

// --- Descending iterator on DB-backed tree ---

func TestCoverage_DescendingIteratorDBBacked(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	for i := 0; i < 100; i++ {
		tree.Set(fmt.Appendf(nil, "di%05d", i), fmt.Appendf(nil, "val%05d", i))
	}
	tree.SaveVersion()

	// Reload
	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	tree2.LoadVersion(1)

	itr, err := tree2.Iterator(nil, nil, false) // descending
	if err != nil {
		t.Fatal(err)
	}
	defer itr.Close()

	var prev string
	count := 0
	for itr.Valid() {
		k := string(itr.Key())
		v := itr.Value()
		if v == nil {
			t.Fatalf("nil value for key %q", k)
		}
		if prev != "" && k >= prev {
			t.Fatalf("order broken: %q >= %q", k, prev)
		}
		prev = k
		count++
		itr.Next()
	}
	if count != 100 {
		t.Fatalf("iterated %d, want 100", count)
	}
}

// --- Non-membership proof on DB-loaded tree ---

func TestCoverage_NonMembershipProofDBBacked(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "np%05d", i), []byte("v"))
	}
	tree.SaveVersion()

	// Reload and get immutable
	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	tree2.LoadVersion(1)
	imm, err := tree2.GetImmutable(1)
	if err != nil {
		t.Fatal(err)
	}

	// Non-membership for key before all
	proof, err := imm.GetNonMembershipProof([]byte("aaa"))
	if err != nil {
		t.Fatalf("non-membership proof (before all): %v", err)
	}
	if proof == nil {
		t.Fatal("nil proof")
	}

	// Non-membership for key after all
	proof, err = imm.GetNonMembershipProof([]byte("zzz"))
	if err != nil {
		t.Fatalf("non-membership proof (after all): %v", err)
	}
	if proof == nil {
		t.Fatal("nil proof")
	}

	// Non-membership for key between existing keys
	proof, err = imm.GetNonMembershipProof([]byte("np00005_missing"))
	if err != nil {
		t.Fatalf("non-membership proof (between): %v", err)
	}
	if proof == nil {
		t.Fatal("nil proof")
	}
}

// --- LoadVersion for a pruned version ---

func TestCoverage_LoadVersionPruned(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	tree.Set([]byte("a"), []byte("1"))
	tree.SaveVersion()
	tree.Set([]byte("b"), []byte("2"))
	tree.SaveVersion()

	tree.DeleteVersionsTo(1)

	// Try to load pruned version 1 — should error
	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	_, err := tree2.LoadVersion(1)
	if err == nil {
		t.Fatal("expected error loading pruned version")
	}
}

// --- Export of empty tree ---

func TestCoverage_ExportEmptyTree(t *testing.T) {
	// For in-memory trees, GetImmutable on an empty tree (nil root) returns
	// ErrVersionDoesNotExist because lastSaved is nil. This is correct —
	// there's nothing to export from an empty tree.
	tree := NewMutableTreeMem()
	tree.SaveVersion()

	_, err := tree.GetImmutable(1)
	if err == nil {
		t.Fatal("expected error getting immutable of empty tree")
	}

	// For DB-backed trees, same behavior
	db := memdb.NewMemDB()
	tree2 := NewMutableTreeWithDB(db, 100, NewNopLogger())
	tree2.SaveVersion()

	imm2, err := tree2.GetImmutable(1)
	if err != nil {
		// GetImmutable for empty DB tree returns an ImmutableTree with nil root
		t.Logf("GetImmutable returned error (expected for empty): %v", err)
		return
	}
	// If we got an ImmutableTree, Export should fail with ErrNotInitializedTree
	_, err = imm2.Export(tree2.ndb)
	if err == nil {
		t.Fatal("expected error exporting empty tree")
	}
}

// --- SaveVersion after Rollback (no mutations) ---

func TestCoverage_SaveVersionAfterRollback(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	tree.Set([]byte("a"), []byte("v1"))
	hash1, _, _ := tree.SaveVersion()

	tree.Set([]byte("a"), []byte("v2"))
	tree.Rollback()

	// SaveVersion without any mutations — should produce same content as V1
	hash2, v2, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v2 != 2 {
		t.Fatalf("version = %d, want 2", v2)
	}
	if !bytes.Equal(hash1, hash2) {
		t.Fatalf("hash changed after rollback+save: %x != %x", hash1, hash2)
	}
}

// --- GetImmutable on in-memory tree for wrong version ---

func TestCoverage_GetImmutableWrongVersion(t *testing.T) {
	tree := NewMutableTreeMem()
	tree.Set([]byte("a"), []byte("1"))
	tree.SaveVersion()

	_, err := tree.GetImmutable(99)
	if err == nil {
		t.Fatal("expected error for non-existent version on in-memory tree")
	}
}
