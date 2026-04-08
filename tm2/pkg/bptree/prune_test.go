package bptree

import (
	"bytes"
	"fmt"
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

	// Close exporter and retry
	exporter.Close()
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
				t.Fatalf("after prune v%d, v%d hash changed", pruneV, checkV)
			}
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

	// Prune V1 — this is where the bug would trigger: walkAndPrune would
	// search for old children under only one of the new inner children,
	// miss the ones under the sibling, and incorrectly delete them.
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
