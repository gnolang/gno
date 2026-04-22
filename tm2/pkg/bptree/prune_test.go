package bptree

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

func newPruneTree(t *testing.T) *MutableTree {
	t.Helper()
	return NewMutableTreeWithDB(memdb.NewMemDB(), 1000, NewNopLogger())
}

func TestPrune_BasicPrune(t *testing.T) {
	tree := newPruneTree(t)

	// V1: 50 keys
	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "p%03d", i), []byte("v1"))
	}
	tree.SaveVersion()

	// V2: add 20 more
	for i := 50; i < 70; i++ {
		tree.Set(fmt.Appendf(nil, "p%03d", i), []byte("v2"))
	}
	tree.SaveVersion()

	// V3: update some
	for i := 0; i < 10; i++ {
		tree.Set(fmt.Appendf(nil, "p%03d", i), []byte("v3"))
	}
	tree.SaveVersion()

	// Prune v1 and v2
	err := tree.DeleteVersionsTo(2)
	if err != nil {
		t.Fatalf("DeleteVersionsTo(2): %v", err)
	}

	// V1 and V2 should be gone
	if tree.VersionExists(1) || tree.VersionExists(2) {
		t.Fatalf("versions 1-2 should be pruned")
	}

	// V3 should still work
	if !tree.VersionExists(3) {
		t.Fatalf("version 3 should exist")
	}
	imm, err := tree.GetImmutable(3)
	if err != nil {
		t.Fatalf("GetImmutable(3): %v", err)
	}
	if imm.Size() != 70 {
		t.Fatalf("v3 size = %d, want 70", imm.Size())
	}
}

func TestPrune_PruneAndContinue(t *testing.T) {
	tree := newPruneTree(t)

	for i := 0; i < 30; i++ {
		tree.Set(fmt.Appendf(nil, "c%03d", i), []byte("v"))
	}
	tree.SaveVersion()

	tree.Set([]byte("c_new"), []byte("added"))
	tree.SaveVersion()

	// Prune v1
	tree.DeleteVersionsTo(1)

	// V2 should work
	tree2 := NewMutableTreeWithDB(tree.ndb.db, 1000, NewNopLogger())
	tree2.LoadVersion(2)
	if tree2.Size() != 31 {
		t.Fatalf("v2 size = %d, want 31", tree2.Size())
	}

	// Can still make new versions after pruning
	tree.Set([]byte("c_another"), []byte("more"))
	_, v, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion after prune: %v", err)
	}
	if v != 3 {
		t.Fatalf("version = %d, want 3", v)
	}
}

func TestPrune_CannotPruneLatest(t *testing.T) {
	tree := newPruneTree(t)
	tree.Set([]byte("a"), []byte("b"))
	tree.SaveVersion()

	err := tree.DeleteVersionsTo(1)
	if err == nil {
		t.Fatalf("should not be able to prune latest version")
	}
}

func TestPrune_VersionReaders(t *testing.T) {
	tree := newPruneTree(t)
	for i := 0; i < 30; i++ {
		tree.Set(fmt.Appendf(nil, "vr%03d", i), []byte("v"))
	}
	tree.SaveVersion()
	tree.Set([]byte("vr_extra"), []byte("v"))
	tree.SaveVersion()

	// Open an exporter on v1
	imm, _ := tree.GetImmutable(1)
	exporter, _ := imm.Export(tree.ndb)

	// Pruning should fail
	err := tree.DeleteVersionsTo(1)
	if err == nil {
		t.Fatalf("should fail with active reader")
	}

	// Close exporter AND the immutable snapshot, then retry
	exporter.Close()
	imm.Close()
	err = tree.DeleteVersionsTo(1)
	if err != nil {
		t.Fatalf("prune after close: %v", err)
	}
}

func TestPrune_PreservesLatestState(t *testing.T) {
	tree := newPruneTree(t)

	// V1
	for i := 0; i < 100; i++ {
		tree.Set(fmt.Appendf(nil, "ps%04d", i), fmt.Appendf(nil, "val%04d", i))
	}
	tree.SaveVersion()

	// V2: remove some, update some, add some
	for i := 0; i < 20; i++ {
		tree.Remove(fmt.Appendf(nil, "ps%04d", i))
	}
	for i := 20; i < 40; i++ {
		tree.Set(fmt.Appendf(nil, "ps%04d", i), []byte("updated"))
	}
	for i := 100; i < 120; i++ {
		tree.Set(fmt.Appendf(nil, "ps%04d", i), []byte("new"))
	}
	hash2, _, _ := tree.SaveVersion()

	// Prune v1
	tree.DeleteVersionsTo(1)

	// Reload v2 from DB
	tree2 := NewMutableTreeWithDB(tree.ndb.db, 1000, NewNopLogger())
	tree2.LoadVersion(2)

	hash2b := tree2.WorkingHash()
	if !bytes.Equal(hash2, hash2b) {
		t.Fatalf("hash changed after pruning")
	}
	if tree2.Size() != 100 { // 100 - 20 + 20 = 100
		t.Fatalf("size = %d, want 100", tree2.Size())
	}

	// Verify specific keys
	val, _ := tree2.Get([]byte("ps0030"))
	if !bytes.Equal(val, []byte("updated")) {
		t.Fatalf("ps0030 = %q, want 'updated'", val)
	}
	val, _ = tree2.Get([]byte("ps0110"))
	if !bytes.Equal(val, []byte("new")) {
		t.Fatalf("ps0110 = %q, want 'new'", val)
	}
	val, _ = tree2.Get([]byte("ps0010"))
	if val != nil {
		t.Fatalf("ps0010 should be deleted")
	}
}

func TestPrune_MultiplePrunes(t *testing.T) {
	tree := newPruneTree(t)

	// Create 5 versions
	for v := 0; v < 5; v++ {
		for i := 0; i < 20; i++ {
			tree.Set(fmt.Appendf(nil, "mp%03d", i), fmt.Appendf(nil, "v%d", v))
		}
		tree.SaveVersion()
	}

	// Prune v1
	tree.DeleteVersionsTo(1)
	if tree.VersionExists(1) {
		t.Fatalf("v1 should be pruned")
	}

	// Prune v2-v3
	tree.DeleteVersionsTo(3)
	if tree.VersionExists(2) || tree.VersionExists(3) {
		t.Fatalf("v2-v3 should be pruned")
	}

	// V4 and V5 should still work
	for _, v := range []int64{4, 5} {
		imm, err := tree.GetImmutable(v)
		if err != nil {
			t.Fatalf("GetImmutable(%d): %v", v, err)
		}
		if imm.Size() != 20 {
			t.Fatalf("v%d size = %d, want 20", v, imm.Size())
		}
	}
}

func TestPrune_AfterSplitsAndMerges(t *testing.T) {
	tree := newPruneTree(t)

	// V1: sequential inserts causing splits
	for i := 0; i < 200; i++ {
		tree.Set(fmt.Appendf(nil, "sm%04d", i), []byte("v1"))
	}
	tree.SaveVersion()

	// V2: remove many, causing merges
	for i := 0; i < 100; i++ {
		tree.Remove(fmt.Appendf(nil, "sm%04d", i))
	}
	tree.SaveVersion()

	// V3: add more, causing more splits
	for i := 200; i < 300; i++ {
		tree.Set(fmt.Appendf(nil, "sm%04d", i), []byte("v3"))
	}
	hash3, _, _ := tree.SaveVersion()

	// Prune v1 and v2
	err := tree.DeleteVersionsTo(2)
	if err != nil {
		t.Fatalf("DeleteVersionsTo(2): %v", err)
	}

	// V3 should be intact
	tree2 := NewMutableTreeWithDB(tree.ndb.db, 1000, NewNopLogger())
	tree2.LoadVersion(3)

	hash3b := tree2.WorkingHash()
	if !bytes.Equal(hash3, hash3b) {
		t.Fatalf("hash changed after pruning splits/merges")
	}
	if tree2.Size() != 200 { // 200 - 100 + 100
		t.Fatalf("size = %d, want 200", tree2.Size())
	}
}

func TestPrune_IncrementalPreservesAll(t *testing.T) {
	tree := newPruneTree(t)

	// Create 5 versions with different mutations
	hashes := make([][]byte, 6) // hashes[1..5]
	for v := 1; v <= 5; v++ {
		for i := 0; i < 20; i++ {
			tree.Set(
				fmt.Appendf(nil, "ip%03d", i+(v-1)*5),
				fmt.Appendf(nil, "v%d_%d", v, i),
			)
		}
		h, _, _ := tree.SaveVersion()
		hashes[v] = h
	}

	// Prune one at a time, verifying all remaining versions after each
	for pruneV := int64(1); pruneV <= 4; pruneV++ {
		err := tree.DeleteVersionsTo(pruneV)
		if err != nil {
			t.Fatalf("prune v%d: %v", pruneV, err)
		}

		// Check all remaining versions
		for checkV := pruneV + 1; checkV <= 5; checkV++ {
			imm, err := tree.GetImmutable(checkV)
			if err != nil {
				t.Fatalf("after prune v%d, GetImmutable(%d): %v", pruneV, checkV, err)
			}
			h := imm.Hash()
			if !bytes.Equal(h, hashes[checkV]) {
				imm.Close()
				t.Fatalf("after prune v%d, v%d hash changed", pruneV, checkV)
			}
			imm.Close()
		}
	}
}

func TestPrune_DBNodeCountDecreases(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	// V1: 200 keys
	for i := 0; i < 200; i++ {
		tree.Set(fmt.Appendf(nil, "nc%04d", i), []byte("v1"))
	}
	tree.SaveVersion()

	// V2: update 100 keys (creates ~100 new leaf nodes)
	for i := 0; i < 100; i++ {
		tree.Set(fmt.Appendf(nil, "nc%04d", i), []byte("v2"))
	}
	tree.SaveVersion()

	// Count nodes before pruning
	countBefore := countDBNodes(db)

	// Prune v1
	tree.DeleteVersionsTo(1)

	// Count nodes after pruning
	countAfter := countDBNodes(db)

	if countAfter >= countBefore {
		t.Fatalf("node count did not decrease: before=%d after=%d", countBefore, countAfter)
	}
	t.Logf("node count: %d -> %d (deleted %d)", countBefore, countAfter, countBefore-countAfter)
}

func countDBNodes(db *memdb.MemDB) int {
	count := 0
	prefix := []byte{PrefixNode}
	end := []byte{PrefixNode + 1}
	itr, _ := db.Iterator(prefix, end)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		count++
	}
	return count
}

func countDBValues(db *memdb.MemDB) int {
	count := 0
	itr, _ := db.Iterator([]byte{PrefixVal}, []byte{PrefixVal + 1})
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		count++
	}
	return count
}

// --- Value cleanup tests ---

func TestPrune_ValueCountDecreases(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	// V1: 100 unique keys
	for i := 0; i < 100; i++ {
		tree.Set(fmt.Appendf(nil, "vc%05d", i), fmt.Appendf(nil, "val1_%05d", i))
	}
	tree.SaveVersion()

	// V2: update 50 keys with new values
	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "vc%05d", i), fmt.Appendf(nil, "val2_%05d", i))
	}
	tree.SaveVersion()

	before := countDBValues(db)
	if before != 150 {
		t.Fatalf("before prune: %d values, want 150", before)
	}

	// Prune V1
	if err := tree.DeleteVersionsTo(1); err != nil {
		t.Fatal(err)
	}

	after := countDBValues(db)
	if after != 100 {
		t.Fatalf("after prune: %d values, want 100", after)
	}
}

func TestPrune_ValueCountBounded(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	// Initial: 100 keys
	for i := 0; i < 100; i++ {
		tree.Set(fmt.Appendf(nil, "vb%05d", i), fmt.Appendf(nil, "v0_k%d", i))
	}
	tree.SaveVersion()

	// 20 iterations: overwrite all 100 keys, save, prune oldest
	for iter := 1; iter <= 20; iter++ {
		for i := 0; i < 100; i++ {
			tree.Set(fmt.Appendf(nil, "vb%05d", i), fmt.Appendf(nil, "v%d_k%d", iter, i))
		}
		tree.SaveVersion()
		if err := tree.DeleteVersionsTo(int64(iter)); err != nil {
			t.Fatalf("iter %d prune: %v", iter, err)
		}
		count := countDBValues(db)
		if count != 100 {
			t.Fatalf("iter %d: %d values, want 100", iter, count)
		}
	}
}

func TestPrune_OverwrittenValueCleaned(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	tree.Set([]byte("k"), []byte("v1"))
	tree.SaveVersion()
	tree.Set([]byte("k"), []byte("v2"))
	tree.SaveVersion()

	tree.DeleteVersionsTo(1)
	count := countDBValues(db)
	if count != 1 {
		t.Fatalf("after prune: %d values, want 1", count)
	}
}

func TestRemove_OrphanedValueCleaned(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	tree.Set([]byte("k"), []byte("v"))
	tree.SaveVersion()
	tree.Remove([]byte("k"))
	tree.SaveVersion()

	tree.DeleteVersionsTo(1)
	count := countDBValues(db)
	if count != 0 {
		t.Fatalf("after prune: %d values, want 0", count)
	}
}

func TestSet_IntraVersionOverwrite(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	tree.Set([]byte("k"), []byte("v1"))
	tree.Set([]byte("k"), []byte("v2"))
	tree.Set([]byte("k"), []byte("v3"))
	tree.SaveVersion()

	count := countDBValues(db)
	if count != 1 {
		t.Fatalf("after save: %d values, want 1 (v1,v2 should be Tier 1 cleaned)", count)
	}
}

func TestRollback_CleansUpValues(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	tree.Set([]byte("k"), []byte("v1"))
	if countDBValues(db) != 1 {
		t.Fatal("value should be eagerly written")
	}

	tree.Rollback()
	if countDBValues(db) != 0 {
		t.Fatal("rollback should delete eagerly-written values")
	}

	// Normal operation after rollback
	tree.Set([]byte("k"), []byte("v2"))
	tree.SaveVersion()
	if countDBValues(db) != 1 {
		t.Fatal("after rollback+save: should have 1 value")
	}
}

func TestPrune_DisjointKeysPreservesValues(t *testing.T) {
	// Regression: pruning should NOT delete shared values
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "dj_a%05d", i), []byte("va"))
	}
	tree.SaveVersion()

	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "dj_b%05d", i), []byte("vb"))
	}
	tree.SaveVersion()

	tree.DeleteVersionsTo(1)
	count := countDBValues(db)
	if count != 100 {
		t.Fatalf("after prune: %d values, want 100 (all live in V2)", count)
	}
}

func TestRemoveThenReSet_SameKey(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	tree.Set([]byte("k"), []byte("v1"))
	tree.SaveVersion()

	tree.Remove([]byte("k"))
	tree.Set([]byte("k"), []byte("v2"))
	tree.SaveVersion()

	tree.DeleteVersionsTo(1)
	count := countDBValues(db)
	if count != 1 {
		t.Fatalf("after prune: %d values, want 1", count)
	}
	val, _ := tree.Get([]byte("k"))
	if string(val) != "v2" {
		t.Fatalf("Get(k) = %q, want v2", val)
	}
}

func TestExportImport_ValueKeysCorrect(t *testing.T) {
	db1 := memdb.NewMemDB()
	tree1 := NewMutableTreeWithDB(db1, 1000, NewNopLogger())

	for i := 0; i < 100; i++ {
		tree1.Set(fmt.Appendf(nil, "ei%05d", i), fmt.Appendf(nil, "val%05d", i))
	}
	tree1.SaveVersion()

	// Export
	imm, _ := tree1.GetImmutable(1)
	exporter, _ := imm.Export(tree1.ndb)

	// Import
	db2 := memdb.NewMemDB()
	tree2 := NewMutableTreeWithDB(db2, 1000, NewNopLogger())
	imp, _ := tree2.Import(1)
	for {
		node, err := exporter.Next()
		if err == ErrExportDone {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		imp.Add(node)
	}
	exporter.Close()
	imp.Commit()

	// Verify values readable
	for i := 0; i < 100; i++ {
		val, _ := tree2.Get(fmt.Appendf(nil, "ei%05d", i))
		expected := fmt.Appendf(nil, "val%05d", i)
		if string(val) != string(expected) {
			t.Fatalf("Get(ei%05d) = %q, want %q", i, val, expected)
		}
	}

	// Overwrite 30 keys, save V2, prune V1
	for i := 0; i < 30; i++ {
		tree2.Set(fmt.Appendf(nil, "ei%05d", i), fmt.Appendf(nil, "new%05d", i))
	}
	tree2.SaveVersion()
	tree2.DeleteVersionsTo(1)

	count := countDBValues(db2)
	if count != 100 {
		t.Fatalf("after import+prune: %d values, want 100", count)
	}
}

func TestPrune_MultiVersion(t *testing.T) {
	db := memdb.NewMemDB()
	// countDBValues asserts equal PrefixVal entries, which requires
	// external storage; disable inlining for this check.
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger(), InlineValueThresholdOption(-1))

	// V1: keys a,b,c
	tree.Set([]byte("a"), []byte("a1"))
	tree.Set([]byte("b"), []byte("b1"))
	tree.Set([]byte("c"), []byte("c1"))
	tree.SaveVersion()

	// V2: update a
	tree.Set([]byte("a"), []byte("a2"))
	tree.SaveVersion()

	// V3: update b, remove c
	tree.Set([]byte("b"), []byte("b3"))
	tree.Remove([]byte("c"))
	tree.SaveVersion()

	// V4: update a again, add d
	tree.Set([]byte("a"), []byte("a4"))
	tree.Set([]byte("d"), []byte("d4"))
	tree.SaveVersion()

	// V5: update d
	tree.Set([]byte("d"), []byte("d5"))
	tree.SaveVersion()

	// Prune V1-V3
	tree.DeleteVersionsTo(3)

	// V4 has: a=a4, b=b3(shared from V3), d=d4
	// V5 has: a=a4(shared from V4), b=b3(shared), d=d5
	// Live values across V4+V5: a4, b3, d4, d5 = 4
	count := countDBValues(db)
	if count != 4 {
		t.Fatalf("after multi-prune: %d values, want 4", count)
	}
}

func TestSet_EmptyValueCleanup(t *testing.T) {
	db := memdb.NewMemDB()
	// Disable inlining so we exercise the eager SaveValue + DeleteValue
	// path under test.
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger(), InlineValueThresholdOption(-1))

	tree.Set([]byte("k"), []byte{})
	tree.SaveVersion()

	val, _ := tree.Get([]byte("k"))
	if val == nil {
		t.Fatal("Get should return []byte{}, not nil")
	}
	if len(val) != 0 {
		t.Fatalf("Get = %q, want empty", val)
	}

	tree.Set([]byte("k"), []byte("notempty"))
	tree.SaveVersion()
	tree.DeleteVersionsTo(1)

	count := countDBValues(db)
	if count != 1 {
		t.Fatalf("after prune: %d values, want 1", count)
	}
}

func TestPrune_ValueIntegrityAfterOverwrite(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	// V1: 50 keys
	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "vi%03d", i), fmt.Appendf(nil, "val_v1_%03d", i))
	}
	tree.SaveVersion()

	// V2: overwrite 30
	for i := 0; i < 30; i++ {
		tree.Set(fmt.Appendf(nil, "vi%03d", i), fmt.Appendf(nil, "val_v2_%03d", i))
	}
	tree.SaveVersion()

	// V3: overwrite 10
	for i := 0; i < 10; i++ {
		tree.Set(fmt.Appendf(nil, "vi%03d", i), fmt.Appendf(nil, "val_v3_%03d", i))
	}
	tree.SaveVersion()

	// Prune V1 and V2
	tree.DeleteVersionsTo(2)

	// Verify every key returns correct value
	for i := 0; i < 50; i++ {
		key := fmt.Appendf(nil, "vi%03d", i)
		val, err := tree.Get(key)
		if err != nil {
			t.Fatalf("Get(%q): %v", key, err)
		}
		var expected []byte
		switch {
		case i < 10:
			expected = fmt.Appendf(nil, "val_v3_%03d", i)
		case i < 30:
			expected = fmt.Appendf(nil, "val_v2_%03d", i)
		default:
			expected = fmt.Appendf(nil, "val_v1_%03d", i)
		}
		if !bytes.Equal(val, expected) {
			t.Fatalf("Get(%q) = %q, want %q", key, val, expected)
		}
	}

	// Reload from DB and verify again
	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	tree2.LoadVersion(3)
	for i := 0; i < 50; i++ {
		key := fmt.Appendf(nil, "vi%03d", i)
		val, _ := tree2.Get(key)
		var expected []byte
		switch {
		case i < 10:
			expected = fmt.Appendf(nil, "val_v3_%03d", i)
		case i < 30:
			expected = fmt.Appendf(nil, "val_v2_%03d", i)
		default:
			expected = fmt.Appendf(nil, "val_v1_%03d", i)
		}
		if !bytes.Equal(val, expected) {
			t.Fatalf("reloaded Get(%q) = %q, want %q", key, val, expected)
		}
	}
}

// TestPrune_SeparatorKeyRouting verifies that pruning is correct when
// inner-node splits have restructured the tree between versions. Under
// the replaced positional-descent algorithm, separator-key routing could
// pick the wrong peer inner node and incorrectly delete shared nodes;
// the current mark-and-sweep implementation is content-addressed via
// NodeKey and is immune to that class of bug.
//
// This test builds a height-3 tree, makes structural changes at height 2
// (inner node splits), and verifies that pruning doesn't delete shared nodes.
func TestPrune_SeparatorKeyRouting(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger(), InlineValueThresholdOption(-1))

	// V1: Insert enough keys to create a height-2 tree (~1100 keys).
	// With B=32, this gives ~34 leaves under a single root inner node.
	for i := 0; i < 1100; i++ {
		tree.Set(fmt.Appendf(nil, "sk%06d", i), fmt.Appendf(nil, "v1_%06d", i))
	}
	hash1, _, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}
	if tree.Height() < 2 {
		t.Skipf("need height >= 2, got %d", tree.Height())
	}

	// V2: Insert 500 more keys to trigger inner node split at height 1,
	// creating a height-3 tree. This forces mark-and-sweep to traverse
	// both V1's and V2's inner-node layers when computing reachability.
	for i := 1100; i < 1600; i++ {
		tree.Set(fmt.Appendf(nil, "sk%06d", i), fmt.Appendf(nil, "v2_%06d", i))
	}
	// Also update some V1 keys to create cross-version orphans
	for i := 0; i < 200; i++ {
		tree.Set(fmt.Appendf(nil, "sk%06d", i), fmt.Appendf(nil, "v2_upd_%06d", i))
	}
	hash2, _, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}

	// V3: More mutations to ensure V2 isn't the latest when pruning V1
	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "sk%06d", i), fmt.Appendf(nil, "v3_%06d", i))
	}
	hash3, _, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}

	// Prune V1 — this exercises mark-and-sweep across the V1/V2 inner-node
	// restructure.
	if err := tree.DeleteVersionsTo(1); err != nil {
		t.Fatalf("prune V1: %v", err)
	}

	// Verify V2 integrity: reload from DB, check hash and all keys
	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	tree2.LoadVersion(2)
	if !bytes.Equal(hash2, tree2.WorkingHash()) {
		t.Fatalf("V2 hash changed after pruning V1")
	}
	for i := 0; i < 1600; i++ {
		key := fmt.Appendf(nil, "sk%06d", i)
		val, _ := tree2.Get(key)
		if val == nil {
			t.Fatalf("V2: key %q missing after prune", key)
		}
	}

	// Prune V2 — exercises inner node routing again with V3
	if err := tree.DeleteVersionsTo(2); err != nil {
		t.Fatalf("prune V2: %v", err)
	}

	// Verify V3 integrity
	tree3 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	tree3.LoadVersion(3)
	if !bytes.Equal(hash3, tree3.WorkingHash()) {
		t.Fatalf("V3 hash changed after pruning V2")
	}
	for i := 0; i < 1600; i++ {
		key := fmt.Appendf(nil, "sk%06d", i)
		val, _ := tree3.Get(key)
		if val == nil {
			t.Fatalf("V3: key %q missing after prune", key)
		}
	}

	// Also verify value count is correct (no leaked orphans)
	expectedValues := 1600 // all keys live in V3
	actualValues := countDBValues(db)
	if actualValues != expectedValues {
		t.Fatalf("after pruning V1+V2: %d values, want %d", actualValues, expectedValues)
	}

	_ = hash1
}

func TestPrune_InnerNodeSplit(t *testing.T) {
	tree := newPruneTree(t)

	// V1: Insert enough keys to create a height-1 tree (root inner node
	// with ~30+ leaf children). With fan-out 32, ~1100 keys fills the root.
	for i := 0; i < 1100; i++ {
		tree.Set(fmt.Appendf(nil, "split%05d", i), []byte("v1"))
	}
	hash1, v1, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion v1: %v", err)
	}
	if v1 != 1 {
		t.Fatalf("v1 = %d, want 1", v1)
	}
	_ = hash1

	// V2: Insert 300+ more keys to trigger root inner node split
	// (height-1 → height-2). The old root's children are now distributed
	// across two or more new inner nodes.
	for i := 1100; i < 1400; i++ {
		tree.Set(fmt.Appendf(nil, "split%05d", i), []byte("v2"))
	}
	hash2, v2, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion v2: %v", err)
	}
	if v2 != 2 {
		t.Fatalf("v2 = %d, want 2", v2)
	}

	// Prune V1 — this is where the replaced positional-descent algorithm
	// would trigger: it searched for old children under only one of the
	// new inner children, missed the ones under the sibling, and
	// incorrectly deleted them. Mark-and-sweep avoids this by recording
	// every NodeKey reachable from the retained version before deleting.
	err = tree.DeleteVersionsTo(1)
	if err != nil {
		t.Fatalf("DeleteVersionsTo(1): %v", err)
	}

	// Reload V2 from DB and verify integrity
	tree2 := NewMutableTreeWithDB(tree.ndb.db, 1000, NewNopLogger())
	if _, err := tree2.LoadVersion(2); err != nil {
		t.Fatalf("LoadVersion(2) after prune: %v", err)
	}

	hash2b := tree2.WorkingHash()
	if !bytes.Equal(hash2, hash2b) {
		t.Fatalf("V2 hash changed after pruning V1: %x != %x", hash2, hash2b)
	}
	if tree2.Size() != 1400 {
		t.Fatalf("V2 size = %d, want 1400", tree2.Size())
	}

	// Verify all keys are accessible
	for i := 0; i < 1400; i++ {
		key := fmt.Appendf(nil, "split%05d", i)
		val, _ := tree2.Get(key)
		if val == nil {
			t.Fatalf("key %q missing after prune", key)
		}
	}

	// Continue: V3 with more mutations, then prune V2
	for i := 0; i < 200; i++ {
		tree.Set(fmt.Appendf(nil, "split%05d", i), []byte("v3"))
	}
	hash3, _, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion v3: %v", err)
	}

	err = tree.DeleteVersionsTo(2)
	if err != nil {
		t.Fatalf("DeleteVersionsTo(2): %v", err)
	}

	tree3 := NewMutableTreeWithDB(tree.ndb.db, 1000, NewNopLogger())
	if _, err := tree3.LoadVersion(3); err != nil {
		t.Fatalf("LoadVersion(3) after prune: %v", err)
	}
	hash3b := tree3.WorkingHash()
	if !bytes.Equal(hash3, hash3b) {
		t.Fatalf("V3 hash changed after pruning V2: %x != %x", hash3, hash3b)
	}
}

func TestPrune_SustainedInsertPrune(t *testing.T) {
	tree := newPruneTree(t)

	// Bootstrap: insert initial keys so the tree has some structure
	for i := 0; i < 500; i++ {
		tree.Set(fmt.Appendf(nil, "sus%06d", i), []byte("init"))
	}
	_, v, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion initial: %v", err)
	}
	nextKey := 500

	// 20 iterations: insert 200 random keys, save, prune oldest, verify
	for iter := 0; iter < 20; iter++ {
		for i := 0; i < 200; i++ {
			tree.Set(fmt.Appendf(nil, "sus%06d", nextKey), fmt.Appendf(nil, "iter%d", iter))
			nextKey++
		}
		latestHash, newV, err := tree.SaveVersion()
		if err != nil {
			t.Fatalf("iter %d: SaveVersion: %v", iter, err)
		}
		v = newV

		// Prune the oldest surviving version
		oldestToPrune := v - 1
		if oldestToPrune < 1 {
			continue
		}
		err = tree.DeleteVersionsTo(oldestToPrune)
		if err != nil {
			t.Fatalf("iter %d: DeleteVersionsTo(%d): %v", iter, oldestToPrune, err)
		}

		// Verify latest version integrity by reloading from DB
		tree2 := NewMutableTreeWithDB(tree.ndb.db, 1000, NewNopLogger())
		if _, err := tree2.LoadVersion(v); err != nil {
			t.Fatalf("iter %d: LoadVersion(%d): %v", iter, v, err)
		}
		hash2 := tree2.WorkingHash()
		if !bytes.Equal(latestHash, hash2) {
			t.Fatalf("iter %d: hash mismatch after prune: %x != %x", iter, latestHash, hash2)
		}
		expectedSize := int64(500 + (iter+1)*200)
		if tree2.Size() != expectedSize {
			t.Fatalf("iter %d: size = %d, want %d", iter, tree2.Size(), expectedSize)
		}
	}
}

func TestPrune_EmptyVersions(t *testing.T) {
	tree := newPruneTree(t)

	// V1: empty
	tree.SaveVersion()

	// V2: add some keys
	tree.Set([]byte("a"), []byte("b"))
	tree.SaveVersion()

	// Prune v1 (empty)
	err := tree.DeleteVersionsTo(1)
	if err != nil {
		t.Fatalf("prune empty version: %v", err)
	}

	// V2 should work
	if !tree.VersionExists(2) {
		t.Fatalf("v2 should exist")
	}
}

// TestPrune_Height3ChurnAndPrune stresses the mark-and-sweep pruning
// algorithm against a height-3 tree (≥ 32768 keys at B=32) with a
// high-churn delete + insert + prune workload. This is the specific
// regime where the replaced positional-descent algorithm historically
// misidentified subtrees during inner-node splits and merges (see
// POTENTIAL_IMPROVEMENTS.md Finding #3).
//
// The workload per iteration:
//  1. Delete ~20% of live keys (forces merges across inner-node boundaries).
//  2. Insert ~25% new keys (forces splits; some of the new keys fall
//     into ranges that were just merged, causing re-split at different
//     positions than the original split).
//  3. SaveVersion, then prune everything up to latest-1.
//  4. Reload latest from a fresh MutableTree against the same DB and
//     verify every live key, the root hash, and tree size.
//
// Shrinks deterministically with a fixed seed so crashes are
// reproducible. Marked `t.Short()`-skippable because the full run
// pushes ~40k keys through the tree and takes several seconds.
func TestPrune_Height3ChurnAndPrune(t *testing.T) {
	if testing.Short() {
		t.Skip("height-3 stress; skip under -short")
	}

	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 2000, NewNopLogger())

	mirror := make(map[string][]byte)
	rng := rand.New(rand.NewSource(0xbadf00d))

	// Bootstrap: insert 40k keys to reach height 3 (32^3 = 32768).
	// Using 8-byte keys drawn from a 200k-key namespace so subsequent
	// random deletes + inserts land across the full range.
	const namespaceSize = 200_000
	keyBytes := func(id uint64) []byte {
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], id)
		return b[:]
	}
	valBytes := func(rng *rand.Rand) []byte {
		var b [16]byte
		binary.BigEndian.PutUint64(b[:8], rng.Uint64())
		binary.BigEndian.PutUint64(b[8:], rng.Uint64())
		return b[:]
	}

	// Bootstrap.
	for i := 0; i < 40_000; i++ {
		id := uint64(rng.Intn(namespaceSize))
		k := keyBytes(id)
		v := valBytes(rng)
		if _, err := tree.Set(k, v); err != nil {
			t.Fatalf("bootstrap Set: %v", err)
		}
		mirror[string(k)] = v
	}
	hash1, v1, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion bootstrap: %v", err)
	}
	if tree.Height() < 3 {
		t.Fatalf("need height >= 3 for this reproducer, got %d", tree.Height())
	}
	t.Logf("bootstrap: v=%d, size=%d, height=%d, hash=%x", v1, tree.Size(), tree.Height(), hash1[:8])

	// Churn loop: 12 iterations, each a delete+insert+prune cycle.
	const iterations = 12
	currentVer := v1
	for iter := 0; iter < iterations; iter++ {
		// --- 1. Delete ~20% of live keys ---
		liveKeys := make([][]byte, 0, len(mirror))
		for k := range mirror {
			liveKeys = append(liveKeys, []byte(k))
		}
		rng.Shuffle(len(liveKeys), func(i, j int) { liveKeys[i], liveKeys[j] = liveKeys[j], liveKeys[i] })
		toDelete := len(liveKeys) / 5
		for i := 0; i < toDelete; i++ {
			k := liveKeys[i]
			if _, _, err := tree.Remove(k); err != nil {
				t.Fatalf("iter %d: Remove(%x): %v", iter, k, err)
			}
			delete(mirror, string(k))
		}

		// --- 2. Insert ~25% new keys ---
		toInsert := (len(mirror) + toDelete) / 4
		for i := 0; i < toInsert; i++ {
			id := uint64(rng.Intn(namespaceSize))
			k := keyBytes(id)
			v := valBytes(rng)
			if _, err := tree.Set(k, v); err != nil {
				t.Fatalf("iter %d: Set: %v", iter, err)
			}
			mirror[string(k)] = v
		}

		// --- 3. Save and prune ---
		latestHash, newVer, err := tree.SaveVersion()
		if err != nil {
			t.Fatalf("iter %d: SaveVersion: %v", iter, err)
		}
		currentVer = newVer
		if newVer >= 2 {
			if err := tree.DeleteVersionsTo(newVer - 1); err != nil {
				t.Fatalf("iter %d: DeleteVersionsTo(%d): %v", iter, newVer-1, err)
			}
		}

		// --- 4. Verify: reload latest from fresh tree, check hash + every key ---
		tree2 := NewMutableTreeWithDB(db, 2000, NewNopLogger())
		if _, err := tree2.LoadVersion(newVer); err != nil {
			t.Fatalf("iter %d: LoadVersion(%d): %v", iter, newVer, err)
		}
		if !bytes.Equal(latestHash, tree2.WorkingHash()) {
			t.Fatalf("iter %d: hash mismatch after prune: got %x want %x",
				iter, tree2.WorkingHash()[:8], latestHash[:8])
		}
		if int(tree2.Size()) != len(mirror) {
			t.Fatalf("iter %d: size %d != mirror %d", iter, tree2.Size(), len(mirror))
		}
		// Spot-check every live key resolves to the expected value.
		// Iterating the mirror directly so we catch nodes that are
		// silently missing from the DB post-prune.
		for k, want := range mirror {
			got, err := tree2.Get([]byte(k))
			if err != nil {
				t.Fatalf("iter %d: Get(%x) after prune: %v", iter, []byte(k), err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("iter %d: Get(%x) after prune: got %x, want %x",
					iter, []byte(k), got, want)
			}
		}

		if iter%3 == 0 {
			t.Logf("iter %d: v=%d size=%d height=%d", iter, newVer, tree2.Size(), tree2.Height())
		}
	}

	_ = currentVer
}

// TestPrune_CascadingMultiVersionPrune creates many retained versions
// with heavy churn between each, then prunes them all in a single
// DeleteVersionsTo call. This exercises the cascading prune path
// (sweep is invoked per intermediate version), which is where a
// silently corrupted DB from pass N would manifest as an error or
// missing-key failure during pass N+1.
//
// If mark-and-sweep ever incorrectly deleted a node still reachable
// from a retained version, a subsequent pass on that later version
// would panic (via getChild returning an error) or silently produce an
// incorrect hash. Both are caught by the reload/verify loop.
func TestPrune_CascadingMultiVersionPrune(t *testing.T) {
	if testing.Short() {
		t.Skip("cascading multi-version prune; skip under -short")
	}

	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 2000, NewNopLogger())
	rng := rand.New(rand.NewSource(0xbadcafe))

	const namespaceSize = 200_000
	keyBytes := func(id uint64) []byte {
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], id)
		return b[:]
	}
	valBytes := func(rng *rand.Rand) []byte {
		var b [16]byte
		binary.BigEndian.PutUint64(b[:8], rng.Uint64())
		binary.BigEndian.PutUint64(b[8:], rng.Uint64())
		return b[:]
	}

	mirror := make(map[string][]byte)

	// Bootstrap to height 3 with heavy churn.
	for i := 0; i < 40_000; i++ {
		id := uint64(rng.Intn(namespaceSize))
		k := keyBytes(id)
		v := valBytes(rng)
		if _, err := tree.Set(k, v); err != nil {
			t.Fatalf("bootstrap Set: %v", err)
		}
		mirror[string(k)] = v
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("bootstrap SaveVersion: %v", err)
	}
	if tree.Height() < 3 {
		t.Fatalf("need height >= 3, got %d", tree.Height())
	}

	// Create 15 additional versions with a mix of inserts and removes
	// between each SaveVersion. 15 levels of separation between the
	// bootstrap and the final version is enough to exercise all the
	// dual-tree-walk paths: pass 1 (v1 vs v2), pass 2 (v2 vs v3), ...,
	// pass 14 (v14 vs v15).
	const extraVersions = 15
	for v := 0; v < extraVersions; v++ {
		// ~15% churn per version.
		liveKeys := make([][]byte, 0, len(mirror))
		for k := range mirror {
			liveKeys = append(liveKeys, []byte(k))
		}
		rng.Shuffle(len(liveKeys), func(i, j int) { liveKeys[i], liveKeys[j] = liveKeys[j], liveKeys[i] })
		toDelete := len(liveKeys) * 15 / 100
		for i := 0; i < toDelete; i++ {
			k := liveKeys[i]
			if _, _, err := tree.Remove(k); err != nil {
				t.Fatalf("v%d: Remove: %v", v, err)
			}
			delete(mirror, string(k))
		}
		toInsert := toDelete + toDelete/3
		for i := 0; i < toInsert; i++ {
			id := uint64(rng.Intn(namespaceSize))
			k := keyBytes(id)
			v := valBytes(rng)
			if _, err := tree.Set(k, v); err != nil {
				t.Fatalf("Set: %v", err)
			}
			mirror[string(k)] = v
		}
		if _, _, err := tree.SaveVersion(); err != nil {
			t.Fatalf("SaveVersion: %v", err)
		}
	}

	latestHash := tree.WorkingHash()
	latestVer := tree.Version()
	t.Logf("pre-prune: versions=%d, size=%d, height=%d", latestVer, tree.Size(), tree.Height())

	// CRITICAL: prune all versions except the latest in a single call.
	// This drives the per-version sweep for v=1,2,...,latestVer-1
	// consecutively, which is exactly the cascading path documented in
	// Finding #3.
	if err := tree.DeleteVersionsTo(latestVer - 1); err != nil {
		t.Fatalf("DeleteVersionsTo(%d): %v", latestVer-1, err)
	}

	// Reload the latest from a fresh tree handle.
	tree2 := NewMutableTreeWithDB(db, 2000, NewNopLogger())
	if _, err := tree2.LoadVersion(latestVer); err != nil {
		t.Fatalf("LoadVersion(%d) after cascading prune: %v", latestVer, err)
	}
	if !bytes.Equal(latestHash, tree2.WorkingHash()) {
		t.Fatalf("hash mismatch after cascading prune: got %x want %x",
			tree2.WorkingHash()[:8], latestHash[:8])
	}
	if int(tree2.Size()) != len(mirror) {
		t.Fatalf("size mismatch: tree=%d mirror=%d", tree2.Size(), len(mirror))
	}
	for k, want := range mirror {
		got, err := tree2.Get([]byte(k))
		if err != nil {
			t.Fatalf("Get(%x) after cascading prune: %v", []byte(k), err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("Get(%x): got %x want %x", []byte(k), got, want)
		}
	}
	t.Logf("post-prune: size=%d, height=%d, nodes=%d, values=%d",
		tree2.Size(), tree2.Height(), countDBNodes(db), countDBValues(db))
}

// TestPrune_HeightOscillation drives a workload that repeatedly grows
// and shrinks the tree through heights 2 → 3 → 2 → 3, pruning every
// intermediate version. If any prune pass incorrectly deletes a node
// shared with a later version, a subsequent cycle will fail either:
//   - at LoadVersion time ("failed to load child node"),
//   - at SaveVersion time (rebuild reads corrupted children),
//   - at Get-after-prune time (value missing from mirror).
//
// The shrink-then-grow pattern stresses mark-and-sweep by making the
// same key range appear at different tree heights across versions: the
// reachable set must record every NodeKey reachable from the retained
// version regardless of the structural transition between passes.
func TestPrune_HeightOscillation(t *testing.T) {
	if testing.Short() {
		t.Skip("height oscillation stress; skip under -short")
	}

	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 2000, NewNopLogger())
	rng := rand.New(rand.NewSource(0xfeedface))

	keyBytes := func(id uint64) []byte {
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], id)
		return b[:]
	}
	valBytes := func(rng *rand.Rand) []byte {
		var b [16]byte
		binary.BigEndian.PutUint64(b[:8], rng.Uint64())
		binary.BigEndian.PutUint64(b[8:], rng.Uint64())
		return b[:]
	}

	mirror := make(map[string][]byte)

	// Phase 1: Grow to height 3.
	for i := 0; i < 40_000; i++ {
		k := keyBytes(uint64(rng.Intn(200_000)))
		v := valBytes(rng)
		tree.Set(k, v)
		mirror[string(k)] = v
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("bootstrap save: %v", err)
	}
	if tree.Height() < 3 {
		t.Fatalf("want height >= 3, got %d", tree.Height())
	}

	// 4 oscillation cycles: shrink to height 2, grow back to height 3.
	// After each phase, prune ALL prior versions to force the
	// dual-tree-walk across the structural transition.
	for cycle := 0; cycle < 4; cycle++ {
		// --- Shrink: delete 80% of keys, dropping to height 2. ---
		liveKeys := make([][]byte, 0, len(mirror))
		for k := range mirror {
			liveKeys = append(liveKeys, []byte(k))
		}
		rng.Shuffle(len(liveKeys), func(i, j int) { liveKeys[i], liveKeys[j] = liveKeys[j], liveKeys[i] })
		toDelete := len(liveKeys) * 4 / 5
		for i := 0; i < toDelete; i++ {
			k := liveKeys[i]
			if _, _, err := tree.Remove(k); err != nil {
				t.Fatalf("cycle %d shrink Remove: %v", cycle, err)
			}
			delete(mirror, string(k))
		}
		latestHash, shrunkVer, err := tree.SaveVersion()
		if err != nil {
			t.Fatalf("cycle %d shrink save: %v", cycle, err)
		}
		// Prune all prior versions.
		if shrunkVer >= 2 {
			if err := tree.DeleteVersionsTo(shrunkVer - 1); err != nil {
				t.Fatalf("cycle %d shrink prune: %v", cycle, err)
			}
		}
		// Verify.
		t2 := NewMutableTreeWithDB(db, 2000, NewNopLogger())
		if _, err := t2.LoadVersion(shrunkVer); err != nil {
			t.Fatalf("cycle %d shrink reload: %v", cycle, err)
		}
		if !bytes.Equal(latestHash, t2.WorkingHash()) {
			t.Fatalf("cycle %d shrink: hash mismatch", cycle)
		}
		if int(t2.Size()) != len(mirror) {
			t.Fatalf("cycle %d shrink: size %d != mirror %d", cycle, t2.Size(), len(mirror))
		}
		for k, want := range mirror {
			got, _ := t2.Get([]byte(k))
			if !bytes.Equal(got, want) {
				t.Fatalf("cycle %d shrink: Get(%x) got %x want %x", cycle, []byte(k), got, want)
			}
		}
		t.Logf("cycle %d shrunk: v=%d size=%d height=%d", cycle, shrunkVer, t2.Size(), t2.Height())

		// --- Grow: re-insert 40k keys with fresh values, back to height 3. ---
		for i := 0; i < 40_000; i++ {
			k := keyBytes(uint64(rng.Intn(200_000)))
			v := valBytes(rng)
			tree.Set(k, v)
			mirror[string(k)] = v
		}
		latestHash, grewVer, err := tree.SaveVersion()
		if err != nil {
			t.Fatalf("cycle %d grow save: %v", cycle, err)
		}
		if err := tree.DeleteVersionsTo(grewVer - 1); err != nil {
			t.Fatalf("cycle %d grow prune: %v", cycle, err)
		}
		t3 := NewMutableTreeWithDB(db, 2000, NewNopLogger())
		if _, err := t3.LoadVersion(grewVer); err != nil {
			t.Fatalf("cycle %d grow reload: %v", cycle, err)
		}
		if !bytes.Equal(latestHash, t3.WorkingHash()) {
			t.Fatalf("cycle %d grow: hash mismatch", cycle)
		}
		if int(t3.Size()) != len(mirror) {
			t.Fatalf("cycle %d grow: size %d != mirror %d", cycle, t3.Size(), len(mirror))
		}
		for k, want := range mirror {
			got, _ := t3.Get([]byte(k))
			if !bytes.Equal(got, want) {
				t.Fatalf("cycle %d grow: Get(%x) got %x want %x", cycle, []byte(k), got, want)
			}
		}
		t.Logf("cycle %d grew: v=%d size=%d height=%d", cycle, grewVer, t3.Size(), t3.Height())
	}
}

// TestPrune_EmptyVersionOrphansCleaned verifies that when a version's
// root is nil (empty tree), pruning still processes the orphan record
// at the next version. Pre-fix, an early `return nil` when vRootNK == nil
// skipped orphan cleanup, leaking values and the orphan record itself.
// See Finding #2 in POTENTIAL_IMPROVEMENTS.md.
func TestPrune_EmptyVersionOrphansCleaned(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	// V1: empty tree. The pruner will run the sweep for (v=1, nextV=2)
	// with vRootNK == nil — this is the code path we care about.
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("V1 SaveVersion: %v", err)
	}

	// V2: non-empty. We want a next version to exist (cannot prune latest).
	if _, err := tree.Set([]byte("real-key"), []byte("real-val")); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("V2 SaveVersion: %v", err)
	}

	// Inject an orphan record at V2 pointing to a fabricated ValueKey,
	// and store a value under that ValueKey. This simulates the state
	// where V2's creation displaced a value from an earlier version —
	// the scenario the pre-fix prune failed to clean up.
	fakeVK := (&NodeKey{Version: 1, Nonce: 99}).GetKey()
	fakeValue := []byte("leaked-if-prune-skips-orphan-block")
	if err := tree.ndb.SaveValue(fakeValue, fakeVK); err != nil {
		t.Fatalf("SaveValue: %v", err)
	}
	if err := tree.ndb.SaveOrphans(2, [][]byte{fakeVK}); err != nil {
		t.Fatalf("SaveOrphans: %v", err)
	}
	if err := tree.ndb.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	// Sanity: value and orphan record are present pre-prune.
	if got, err := tree.ndb.GetValue(fakeVK); err != nil {
		t.Fatalf("pre-prune GetValue: %v", err)
	} else if !bytes.Equal(got, fakeValue) {
		t.Fatalf("pre-prune GetValue: got %q, want %q", got, fakeValue)
	}
	if orphans, err := tree.ndb.LoadOrphans(2); err != nil {
		t.Fatalf("pre-prune LoadOrphans(2): %v", err)
	} else if len(orphans) != 1 {
		t.Fatalf("pre-prune orphan count = %d, want 1", len(orphans))
	}

	// V3: another save so V2 isn't the latest (cannot prune latest).
	if _, err := tree.Set([]byte("another"), []byte("val")); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("V3 SaveVersion: %v", err)
	}

	// Prune V1. This triggers the sweep for (v=1, nextV=2) where V1's
	// root is nil. With the fix, the orphan block still runs and cleans up.
	if err := tree.DeleteVersionsTo(1); err != nil {
		t.Fatalf("DeleteVersionsTo(1): %v", err)
	}

	// The fake orphan value must be deleted. Post-fix, GetValue returns
	// ErrValueMissing (not (nil, nil)) for an absent record so chain
	// divergence surfaces as a typed error rather than passing through
	// silently. See ajnavarro PR #5571 review.
	got, err := tree.ndb.GetValue(fakeVK)
	if !errors.Is(err, ErrValueMissing) {
		t.Fatalf("post-prune GetValue: want ErrValueMissing, got val=%q err=%v", got, err)
	}

	// V2's orphan record must be deleted.
	orphans, err := tree.ndb.LoadOrphans(2)
	if err != nil {
		t.Fatalf("post-prune LoadOrphans(2): %v", err)
	}
	if len(orphans) != 0 {
		t.Fatalf("V2 orphan record leaked: %v", orphans)
	}
}

// TestPrune_BothVersionsEmptyOrphansCleaned covers the degenerate case
// where BOTH v and nextV have empty-tree roots but an orphan record
// still exists at nextV. The fix must process orphans even when both
// trees are empty.
func TestPrune_BothVersionsEmptyOrphansCleaned(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	// V1: empty
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("V1: %v", err)
	}
	// V2: empty
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("V2: %v", err)
	}

	// Inject orphan record + value at V2.
	fakeVK := (&NodeKey{Version: 1, Nonce: 42}).GetKey()
	fakeValue := []byte("orphan-across-two-empty-versions")
	if err := tree.ndb.SaveValue(fakeValue, fakeVK); err != nil {
		t.Fatalf("SaveValue: %v", err)
	}
	if err := tree.ndb.SaveOrphans(2, [][]byte{fakeVK}); err != nil {
		t.Fatalf("SaveOrphans: %v", err)
	}
	if err := tree.ndb.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	// V3: non-empty, so V2 isn't latest.
	if _, err := tree.Set([]byte("k"), []byte("v")); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("V3: %v", err)
	}

	// Prune V1. The sweep for (v=1, nextV=2) runs with both roots nil.
	if err := tree.DeleteVersionsTo(1); err != nil {
		t.Fatalf("DeleteVersionsTo(1): %v", err)
	}

	if got, _ := tree.ndb.GetValue(fakeVK); got != nil {
		t.Fatalf("orphan value leaked: %q", got)
	}
	if orphans, _ := tree.ndb.LoadOrphans(2); len(orphans) != 0 {
		t.Fatalf("V2 orphan record leaked: %v", orphans)
	}
}
