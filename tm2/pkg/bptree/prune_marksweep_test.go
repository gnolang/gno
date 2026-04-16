package bptree

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// The tests in this file exercise workloads that historically broke the
// earlier positional-descent pruning algorithm documented in
// POTENTIAL_IMPROVEMENTS.md Finding #3. They now serve as regression
// tests for the mark-and-sweep implementation that replaced it.

// TestMarkSweepPrune_Basic validates the happy path: a few versions,
// a prune, and a reload check.
func TestMarkSweepPrune_Basic(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 1000, NewNopLogger())

	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "p%03d", i), []byte("v1"))
	}
	tree.SaveVersion()

	for i := 50; i < 70; i++ {
		tree.Set(fmt.Appendf(nil, "p%03d", i), []byte("v2"))
	}
	tree.SaveVersion()

	for i := 0; i < 10; i++ {
		tree.Set(fmt.Appendf(nil, "p%03d", i), []byte("v3"))
	}
	tree.SaveVersion()

	if err := tree.DeleteVersionsTo(2); err != nil {
		t.Fatalf("DeleteVersionsTo(2): %v", err)
	}
	if tree.VersionExists(1) || tree.VersionExists(2) {
		t.Fatalf("versions 1-2 should be pruned")
	}
	imm, err := tree.GetImmutable(3)
	if err != nil {
		t.Fatalf("GetImmutable(3): %v", err)
	}
	defer imm.Close()
	if imm.Size() != 70 {
		t.Fatalf("v3 size = %d, want 70", imm.Size())
	}
}

// TestMarkSweepPrune_ReloadAfterPrune verifies that after a prune run,
// a fresh MutableTree loading the surviving version sees exactly the
// expected key-value content.
func TestMarkSweepPrune_ReloadAfterPrune(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	// V1: 100 keys.
	for i := 0; i < 100; i++ {
		tree.Set(fmt.Appendf(nil, "ms%04d", i), fmt.Appendf(nil, "val%04d", i))
	}
	tree.SaveVersion()
	// V2: remove some, update some, add some.
	for i := 0; i < 20; i++ {
		tree.Remove(fmt.Appendf(nil, "ms%04d", i))
	}
	for i := 20; i < 40; i++ {
		tree.Set(fmt.Appendf(nil, "ms%04d", i), []byte("updated"))
	}
	for i := 100; i < 120; i++ {
		tree.Set(fmt.Appendf(nil, "ms%04d", i), []byte("new"))
	}
	hash2, _, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion v2: %v", err)
	}

	if err := tree.DeleteVersionsTo(1); err != nil {
		t.Fatalf("DeleteVersionsTo(1): %v", err)
	}

	// Reload v2 in a fresh tree and check content.
	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	if _, err := tree2.LoadVersion(2); err != nil {
		t.Fatalf("LoadVersion(2): %v", err)
	}
	if !bytes.Equal(tree2.Hash(), hash2) {
		t.Fatalf("reloaded v2 hash mismatch: got %x, want %x", tree2.Hash(), hash2)
	}
	// Removed keys should be absent.
	for i := 0; i < 20; i++ {
		got, err := tree2.Get(fmt.Appendf(nil, "ms%04d", i))
		if err != nil {
			t.Fatalf("Get(removed): %v", err)
		}
		if got != nil {
			t.Fatalf("removed key ms%04d still present: %q", i, got)
		}
	}
	// Updated keys should have the new value.
	for i := 20; i < 40; i++ {
		got, err := tree2.Get(fmt.Appendf(nil, "ms%04d", i))
		if err != nil {
			t.Fatalf("Get(updated): %v", err)
		}
		if !bytes.Equal(got, []byte("updated")) {
			t.Fatalf("updated key ms%04d got %q, want 'updated'", i, got)
		}
	}
}

// TestMarkSweepPrune_SustainedInsertPrune runs a seeded random-insert /
// prune workload. With the earlier positional-descent algorithm this
// historically surfaced "bptree: failed to load child node ..." panics;
// under mark-and-sweep it must run cleanly to completion and preserve
// observable content at each step.
func TestMarkSweepPrune_SustainedInsertPrune(t *testing.T) {
	if testing.Short() {
		t.Skip("long running; skip under -short")
	}
	seeds := []int64{1, 42, 0xdeadbeef, 0xbadc0ffee}
	const opBytes = 50_000

	for _, seed := range seeds {
		t.Run(fmt.Sprintf("seed=%x", uint64(seed)), func(t *testing.T) {
			data := make([]byte, 8+opBytes)
			binary.LittleEndian.PutUint64(data, uint64(seed))
			rng := rand.New(rand.NewSource(seed))
			for i := 8; i < len(data); i++ {
				data[i] = byte(rng.Intn(256))
			}
			driveRandomOpsMarkSweep(t, data)
		})
	}
}

func driveRandomOpsMarkSweep(t *testing.T, data []byte) {
	t.Helper()

	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 64, NewNopLogger())
	mirror := make(map[string][]byte)
	committed := make(map[string][]byte)
	var lastVer int64

	seed := binary.LittleEndian.Uint64(padTo8(data[:min(len(data), 8)]))
	rng := rand.New(rand.NewSource(int64(seed)))

	const keySpace = 256

	pos := 8
	for pos < len(data) && !t.Failed() {
		op := data[pos] & 0x7
		pos++

		switch op {
		case 0, 1, 2, 3:
			k := deriveKey(rng, keySpace)
			v := deriveValue(rng)
			if _, err := tree.Set(k, v); err != nil {
				t.Fatalf("Set: %v", err)
			}
			mirror[string(k)] = v

		case 4:
			if len(mirror) == 0 {
				continue
			}
			k := pickMirrorKey(rng, mirror)
			if _, _, err := tree.Remove(k); err != nil {
				t.Fatalf("Remove: %v", err)
			}
			delete(mirror, string(k))

		case 5:
			_, v, err := tree.SaveVersion()
			if err != nil {
				t.Fatalf("SaveVersion: %v", err)
			}
			lastVer = v
			committed = copyMirror(mirror)
			assertMirrorMatchesTree(t, tree, mirror)

		case 6:
			if lastVer < 2 {
				continue
			}
			target := int64(1) + rng.Int63n(lastVer-1)
			if err := tree.PruneVersionsTo(target); err != nil {
				t.Fatalf("PruneVersionsTo(%d): %v", target, err)
			}
			assertMirrorMatchesTree(t, tree, mirror)

		case 7:
			if lastVer == 0 {
				continue
			}
			t2 := NewMutableTreeWithDB(db, 64, NewNopLogger())
			if _, err := t2.LoadVersion(lastVer); err != nil {
				t.Fatalf("LoadVersion(%d): %v", lastVer, err)
			}
			assertMirrorMatchesTree(t, t2, committed)
		}
	}
}

// TestMarkSweepPrune_NodeCountDecreases confirms the sweep phase is
// doing real work: after pruning an old version whose tree was largely
// replaced, the DB node count must strictly decrease.
func TestMarkSweepPrune_NodeCountDecreases(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	for i := 0; i < 500; i++ {
		tree.Set(fmt.Appendf(nil, "msN%05d", i), []byte("v1"))
	}
	tree.SaveVersion()
	// Overwrite every key so most inner nodes are replaced in v2.
	for i := 0; i < 500; i++ {
		tree.Set(fmt.Appendf(nil, "msN%05d", i), []byte("v2-changed-significantly"))
	}
	tree.SaveVersion()

	before := countDBNodes(db)
	if err := tree.DeleteVersionsTo(1); err != nil {
		t.Fatalf("DeleteVersionsTo(1): %v", err)
	}
	after := countDBNodes(db)
	if after >= before {
		t.Fatalf("node count did not decrease: before=%d after=%d", before, after)
	}
	t.Logf("node count: %d -> %d (deleted %d)", before, after, before-after)
}

// TestMarkSweepPrune_PreservesSharedSubtrees verifies the shared-subtree
// optimisation: when the old tree shares a subtree with the new tree,
// the sweep skips descending and leaves those nodes alive.
//
// Construction: V1 has 200 keys. V2 changes only a single key. Almost
// every leaf/inner in V1 is still shared with V2. After pruning V1,
// the DB must still be able to reload V2 with the full mirror.
func TestMarkSweepPrune_PreservesSharedSubtrees(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	mirror := make(map[string][]byte)
	for i := 0; i < 200; i++ {
		k := fmt.Appendf(nil, "ss%04d", i)
		v := fmt.Appendf(nil, "val%04d", i)
		tree.Set(k, v)
		mirror[string(k)] = v
	}
	tree.SaveVersion()

	// Update a single key.
	k := []byte("ss0042")
	v := []byte("touched")
	tree.Set(k, v)
	mirror[string(k)] = v
	hash2, _, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion v2: %v", err)
	}

	if err := tree.DeleteVersionsTo(1); err != nil {
		t.Fatalf("DeleteVersionsTo(1): %v", err)
	}

	// Reload v2 fresh and assert mirror match.
	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	if _, err := tree2.LoadVersion(2); err != nil {
		t.Fatalf("LoadVersion(2): %v", err)
	}
	if !bytes.Equal(tree2.Hash(), hash2) {
		t.Fatalf("hash mismatch after prune: got %x, want %x", tree2.Hash(), hash2)
	}
	assertMirrorMatchesTree(t, tree2, mirror)
}

// TestMarkSweepPrune_EmptyOldVersion covers the interaction between
// mark-and-sweep pruning and Finding #2 (empty-tree branch orphan
// handling). When the old version is empty but the next version has
// orphans associated with it, those orphans must still be deleted.
func TestMarkSweepPrune_EmptyOldVersion(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	// V1 is empty.
	tree.SaveVersion()
	// V2 populates and V3 overwrites, producing orphan value records.
	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "eo%04d", i), []byte("v1"))
	}
	tree.SaveVersion()
	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "eo%04d", i), []byte("v2-overwrite"))
	}
	tree.SaveVersion()

	if err := tree.DeleteVersionsTo(2); err != nil {
		t.Fatalf("DeleteVersionsTo(2): %v", err)
	}
	// v3 must still load and carry the current state.
	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	if _, err := tree2.LoadVersion(3); err != nil {
		t.Fatalf("LoadVersion(3): %v", err)
	}
	for i := 0; i < 50; i++ {
		got, err := tree2.Get(fmt.Appendf(nil, "eo%04d", i))
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if !bytes.Equal(got, []byte("v2-overwrite")) {
			t.Fatalf("key eo%04d got %q, want v2-overwrite", i, got)
		}
	}
}
