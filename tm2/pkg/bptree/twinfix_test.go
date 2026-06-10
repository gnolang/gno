package bptree

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// collectReachable walks every retained version's tree from its persisted root
// and returns the sets of reachable node records (by NodeKey bytes) and
// referenced value records (by valueKey bytes). A reachable-but-missing node
// fails the test immediately — that is the over-deletion detector.
func collectReachable(t testing.TB, tree *MutableTree) (nodes, values map[string]bool) {
	t.Helper()
	nodes, values = map[string]bool{}, map[string]bool{}
	// Read the RAW DB, bypassing the node cache: an over-deleting prune
	// re-reads (and re-caches) its own deleted records before the batch
	// commits, so a cache-first read would mask exactly the over-deletion
	// this walk exists to detect.
	loadRaw := func(ref []byte) (Node, error) {
		data, err := tree.ndb.db.Get(nodeDBKey(ref))
		if err != nil {
			return nil, err
		}
		if data == nil {
			return nil, fmt.Errorf("node record %x not in DB", ref)
		}
		return ReadNode(GetNodeKey(ref), data)
	}
	var walk func(ref []byte)
	walk = func(ref []byte) {
		if nodes[string(ref)] {
			return
		}
		nodes[string(ref)] = true
		n, err := loadRaw(ref)
		if err != nil {
			t.Fatalf("OVER-DELETION: retained version references missing node %x: %v", ref, err)
		}
		switch nn := n.(type) {
		case *InnerNode:
			for i := 0; i < nn.NumChildren(); i++ {
				walk(nn.children[i])
			}
		case *LeafNode:
			for i := 0; i < int(nn.numKeys); i++ {
				values[string(nn.valueKeys[i])] = true
			}
		}
	}
	for _, v := range tree.AvailableVersions() {
		nk, _, err := tree.ndb.GetRoot(int64(v))
		if err != nil {
			t.Fatal(err)
		}
		if nk != nil {
			walk(nk)
		}
	}
	return nodes, values
}

// assertNoGarbage is the exact garbage oracle: every node record in the DB
// must be reachable from some retained version's root, and (when checkValues)
// every value record must be referenced by some retained leaf. This is the
// leak detector; collectReachable inside it is the over-deletion detector.
func assertNoGarbage(t testing.TB, tree *MutableTree, checkValues bool) {
	t.Helper()
	nodes, values := collectReachable(t, tree)
	it, err := tree.ndb.db.Iterator(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer it.Close()
	for ; it.Valid(); it.Next() {
		k := it.Key()
		if len(k) == 0 {
			continue
		}
		switch k[0] {
		case PrefixNode:
			if !nodes[string(k[1:])] {
				t.Fatalf("LEAK: node record %x unreachable from any retained version", k[1:])
			}
		case PrefixVal:
			if checkValues && !values[string(k[1:])] {
				t.Fatalf("LEAK: value record %x referenced by no retained leaf", k[1:])
			}
		}
	}
}

// TestTwinFix_NetZeroTwin_RootLeaf (M19, comparison site 1): a session that
// nets back to identical content re-saves the root leaf under a new NodeKey —
// a content-identical "twin". Hash-based sharing skipped the old record as
// shared, leaking it.
func TestTwinFix_NetZeroTwin_RootLeaf(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 1000, NewNopLogger())
	for i := 0; i < 10; i++ {
		tree.Set(fmt.Appendf(nil, "tz%03d", i), []byte("v"))
	}
	if _, _, err := tree.SaveVersion(); err != nil { // v1
		t.Fatal(err)
	}
	tree.Set([]byte("tz_tmp"), []byte("x"))
	if _, _, err := tree.Remove([]byte("tz_tmp")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := tree.SaveVersion(); err != nil { // v2: twin of v1's root
		t.Fatal(err)
	}
	tree.Set([]byte("tz_real"), []byte("y"))
	if _, _, err := tree.SaveVersion(); err != nil { // v3
		t.Fatal(err)
	}
	if err := tree.DeleteVersionsTo(2); err != nil {
		t.Fatal(err)
	}
	assertNoGarbage(t, tree, true)
}

// TestTwinFix_NetZeroTwin_UnderInner (M19, sites 2/3): the twin sits under an
// inner node, so the whole root→leaf path twins. Sequential keys give a
// [31, N] split; the net-zero targets the 31-key left leaf (stays well above
// MinKeys after the remove, so no redistribute destroys the twin).
func TestTwinFix_NetZeroTwin_UnderInner(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 1000, NewNopLogger())
	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "tw%03d", i), []byte("v"))
	}
	if _, _, err := tree.SaveVersion(); err != nil { // v1: inner root, 2 leaves
		t.Fatal(err)
	}
	tree.Set([]byte("tw015x"), []byte("x")) // into the 31-key left leaf
	if _, _, err := tree.Remove([]byte("tw015x")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := tree.SaveVersion(); err != nil { // v2: twin path
		t.Fatal(err)
	}
	tree.Set([]byte("tw_real"), []byte("y"))
	if _, _, err := tree.SaveVersion(); err != nil { // v3
		t.Fatal(err)
	}
	if err := tree.DeleteVersionsTo(2); err != nil {
		t.Fatal(err)
	}
	assertNoGarbage(t, tree, true)
}

// TestTwinFix_SameValueRewriteTwin (M19): overwriting a key with the SAME
// value bytes allocates a new valueKey but leaves the leaf hash unchanged
// (hashes cover keys+valueHashes, not valueKeys) — a twin with different
// content at the record level.
func TestTwinFix_SameValueRewriteTwin(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 1000, NewNopLogger())
	for i := 0; i < 10; i++ {
		tree.Set(fmt.Appendf(nil, "sv%03d", i), []byte("v"))
	}
	if _, _, err := tree.SaveVersion(); err != nil { // v1
		t.Fatal(err)
	}
	tree.Set([]byte("sv005"), []byte("v"))           // same value → same hash, new valueKey
	if _, _, err := tree.SaveVersion(); err != nil { // v2: twin
		t.Fatal(err)
	}
	tree.Set([]byte("sv_real"), []byte("y"))
	if _, _, err := tree.SaveVersion(); err != nil { // v3
		t.Fatal(err)
	}
	if err := tree.DeleteVersionsTo(2); err != nil {
		t.Fatal(err)
	}
	assertNoGarbage(t, tree, true)
}

// exportInto replays an Export stream of imm into an Importer at version.
func exportInto(t testing.TB, tree *MutableTree, imm *ImmutableTree, version int64) {
	t.Helper()
	exp, err := imm.Export(tree.ndb)
	if err != nil {
		t.Fatal(err)
	}
	defer exp.Close()
	imp, err := tree.Import(version)
	if err != nil {
		t.Fatal(err)
	}
	for {
		node, err := exp.Next()
		if err == ErrExportDone {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if err := imp.Add(node); err != nil {
			t.Fatal(err)
		}
	}
	if err := imp.Commit(); err != nil {
		t.Fatal(err)
	}
}

// TestTwinFix_ImportThenPrune (M19 worst case): an Import twins the entire
// exported tree under fresh NodeKeys; hash-based sharing then skipped 100% of
// the pre-import records at prune. Values are checked for the known M21 wall
// (pre-import values are never orphan-listed — bounded, filed separately), so
// the oracle here is nodes-only.
func TestTwinFix_ImportThenPrune(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 1000, NewNopLogger())
	for i := 0; i < 100; i++ {
		tree.Set(fmt.Appendf(nil, "ip%04d", i), fmt.Appendf(nil, "v%04d", i))
	}
	if _, _, err := tree.SaveVersion(); err != nil { // v1
		t.Fatal(err)
	}
	tree.Set([]byte("ip0000"), []byte("v2"))
	if _, _, err := tree.SaveVersion(); err != nil { // v2
		t.Fatal(err)
	}
	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatal(err)
	}
	exportInto(t, tree, imm, 3) // v3 = twin of v1's content, all-fresh records
	imm.Close()

	if err := tree.DeleteVersionsTo(2); err != nil {
		t.Fatal(err)
	}
	assertNoGarbage(t, tree, false) // nodes-only: M21 import value wall
	// The imported version must be fully readable.
	imm3, err := tree.GetImmutable(3)
	if err != nil {
		t.Fatal(err)
	}
	defer imm3.Close()
	for i := 1; i < 100; i++ {
		v, err := imm3.Get(fmt.Appendf(nil, "ip%04d", i))
		if err != nil || !bytes.Equal(v, fmt.Appendf(nil, "v%04d", i)) {
			t.Fatalf("imported version key %d unreadable: %v", i, err)
		}
	}
}

// TestFix2_DivergentReplay_Values: a same-hash replay that omits a same-value
// rewrite (allocated BEFORE the real change in the original session) shifts
// valueKey nonces; keeping the replayed tree makes the working view resolve
// the wrong persisted value. The idempotent path must adopt the persisted
// version instead.
func TestFix2_DivergentReplay_Values(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	tree.Set([]byte("a"), []byte("A1"))
	tree.Set([]byte("b"), []byte("B"))
	if _, _, err := tree.SaveVersion(); err != nil { // v1
		t.Fatal(err)
	}
	// Original v2 session: same-value rewrite of b FIRST ({2,0}), then the
	// real change to a ({2,1}).
	tree.Set([]byte("b"), []byte("B"))
	tree.Set([]byte("a"), []byte("A2"))
	if _, _, err := tree.SaveVersion(); err != nil { // v2
		t.Fatal(err)
	}

	// Divergent replay: only the real change ({2,0}) → same content hash.
	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	if _, err := tree2.LoadVersion(1); err != nil {
		t.Fatal(err)
	}
	tree2.Set([]byte("a"), []byte("A2"))
	if _, _, err := tree2.SaveVersion(); err != nil { // idempotent v2
		t.Fatal(err)
	}
	got, err := tree2.Get([]byte("a"))
	if err != nil || string(got) != "A2" {
		t.Fatalf("post-idempotent Get(a) = %q, %v; want A2 (replayed graph resolved a foreign valueKey)", got, err)
	}

	// And the adopted lineage must survive a prune.
	tree2.Set([]byte("c"), []byte("C"))
	if _, _, err := tree2.SaveVersion(); err != nil { // v3
		t.Fatal(err)
	}
	if err := tree2.DeleteVersionsTo(2); err != nil {
		t.Fatal(err)
	}
	assertNoGarbage(t, tree2, true)
}

// TestFix2_DivergentReplay_Nodes: the omitted op is a net-zero Set+Remove
// AFTER the real change (nonces align, so values match) — the divergence is in
// NODE records: the persisted v2 references a twin leaf, the replayed graph
// references v1's original. Without adoption, v3 built from the replay
// references a record the persisted chain dropped, and pruning v1..v2
// over-deletes it (cold read of v3 panics).
func TestFix2_DivergentReplay_Nodes(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	// Shuffled inserts → balanced multi-leaf tree; first/last keys live in
	// different leaves.
	for i := 0; i < 40; i++ {
		k := (i * 7) % 40
		tree.Set(fmt.Appendf(nil, "dr%03d", k), fmt.Appendf(nil, "val%03d", k))
	}
	if _, _, err := tree.SaveVersion(); err != nil { // v1
		t.Fatal(err)
	}
	// Original v2: real change FIRST ({2,0}, near the keyspace minimum), then
	// a net-zero Set+Remove near the maximum (different leaf; the leaf stays
	// ≥ MinKeys, so no redistribute destroys the twin).
	tree.Set([]byte("dr001"), []byte("MOD"))
	tree.Set([]byte("dr038x"), []byte("tmp"))
	if _, _, err := tree.Remove([]byte("dr038x")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := tree.SaveVersion(); err != nil { // v2
		t.Fatal(err)
	}

	// Divergent replay: only the real change.
	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	if _, err := tree2.LoadVersion(1); err != nil {
		t.Fatal(err)
	}
	tree2.Set([]byte("dr001"), []byte("MOD"))
	if _, _, err := tree2.SaveVersion(); err != nil { // idempotent v2
		t.Fatal(err)
	}
	// v3's change must land in a leaf OTHER than the twinned one (a key
	// sorting into the twinned leaf would clone it at v3, decoupling v3 from
	// the record the over-deletion route targets). "dr000_a" sorts into the
	// first leaf; the twin is near the keyspace maximum.
	tree2.Set([]byte("dr000_a"), []byte("E"))
	if _, _, err := tree2.SaveVersion(); err != nil { // v3
		t.Fatal(err)
	}
	if err := tree2.DeleteVersionsTo(2); err != nil {
		t.Fatal(err)
	}
	assertNoGarbage(t, tree2, true)

	// Cold restart: v3 must be fully readable (over-deletion would panic here).
	tree3 := NewMutableTreeWithDB(db, 0, NewNopLogger())
	if _, err := tree3.Load(); err != nil {
		t.Fatal(err)
	}
	count := 0
	if _, err := tree3.Iterate(func(k, v []byte) bool { count++; return false }); err != nil {
		t.Fatalf("cold read of v3 failed (over-deletion): %v", err)
	}
	if count != 41 {
		t.Fatalf("v3 has %d keys, want 41", count)
	}
}

// TestFix2_CollisionDeleteReplay: a divergent replay whose net-zero Set+Remove
// runs FIRST gives the tier-1 DeleteValueDirect a valueKey that collides with
// a PERSISTED value ({2,0}); the replayed graph then references a never-
// persisted valueKey for the real change. The idempotent adoption makes the
// working view resolve persisted records only.
func TestFix2_CollisionDeleteReplay(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	tree.Set([]byte("a"), []byte("A1"))
	tree.Set([]byte("b"), []byte("B"))
	if _, _, err := tree.SaveVersion(); err != nil { // v1
		t.Fatal(err)
	}
	tree.Set([]byte("a"), []byte("A2"))              // {2,0}
	if _, _, err := tree.SaveVersion(); err != nil { // v2
		t.Fatal(err)
	}

	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	if _, err := tree2.LoadVersion(1); err != nil {
		t.Fatal(err)
	}
	tree2.Set([]byte("x"), []byte("tmp")) // {2,0} — collides with persisted a-value
	if _, _, err := tree2.Remove([]byte("x")); err != nil {
		t.Fatal(err) // tier-1: stages a delete of {2,0}
	}
	tree2.Set([]byte("a"), []byte("A2"))              // {2,1} — never persisted
	if _, _, err := tree2.SaveVersion(); err != nil { // idempotent v2
		t.Fatal(err)
	}
	got, err := tree2.Get([]byte("a"))
	if err != nil || string(got) != "A2" {
		t.Fatalf("post-idempotent Get(a) = %q, %v; want A2", got, err)
	}
	// The persisted value must also have survived the discarded staged delete.
	got, err = tree2.GetVersioned([]byte("a"), 2)
	if err != nil || string(got) != "A2" {
		t.Fatalf("GetVersioned(a,2) = %q, %v; want A2", got, err)
	}
}

// TestFix2_ImportGapReplay: the state-sync shape. Import creates v(latest+2)
// (a gap); a deep replay fills the gap and passes THROUGH the imported version
// via the idempotent path. Without adoption, the post-gap versions keep
// referencing the pre-import lineage, and pruning up to the gap over-deletes
// records they need.
func TestFix2_ImportGapReplay(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	tree.Set([]byte("a"), []byte("A"))
	tree.Set([]byte("b"), []byte("B"))
	if _, _, err := tree.SaveVersion(); err != nil { // v1
		t.Fatal(err)
	}
	tree.Set([]byte("c"), []byte("C"))
	if _, _, err := tree.SaveVersion(); err != nil { // v2
		t.Fatal(err)
	}
	tree.Set([]byte("d"), []byte("D"))
	if _, _, err := tree.SaveVersion(); err != nil { // v3
		t.Fatal(err)
	}

	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatal(err)
	}
	exportInto(t, tree, imm, 5) // v5 = v1's content, fresh records; v4 is a gap
	imm.Close()

	// Deep replay from v3 through the gap.
	if _, err := tree.LoadVersion(3); err != nil {
		t.Fatal(err)
	}
	tree.Set([]byte("e"), []byte("E"))
	if _, _, err := tree.SaveVersion(); err != nil { // v4 (fills the gap)
		t.Fatal(err)
	}
	for _, k := range []string{"c", "d", "e"} {
		if _, _, err := tree.Remove([]byte(k)); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := tree.SaveVersion(); err != nil { // content == v5 → idempotent
		t.Fatal(err)
	}
	tree.Set([]byte("f"), []byte("F"))
	if _, _, err := tree.SaveVersion(); err != nil { // v6
		t.Fatal(err)
	}

	if err := tree.DeleteVersionsTo(4); err != nil {
		t.Fatal(err)
	}
	assertNoGarbage(t, tree, false) // nodes-only (M21 import value wall)

	// Cold restart: v6 must be fully readable.
	tree2 := NewMutableTreeWithDB(db, 0, NewNopLogger())
	if _, err := tree2.Load(); err != nil {
		t.Fatal(err)
	}
	for k, want := range map[string]string{"a": "A", "b": "B", "f": "F"} {
		got, err := tree2.Get([]byte(k))
		if err != nil || string(got) != want {
			t.Fatalf("cold v6 Get(%s) = %q, %v; want %s", k, got, err, want)
		}
	}
}

// TestTwinFix_ChurnOracle: seeded random churn with twin-makers (removes and
// same-value rewrites), window-3 pruning, the exact garbage oracle after every
// prune, periodic per-version content checks, and a cold restart at the end.
func TestTwinFix_ChurnOracle(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	rng := rand.New(rand.NewSource(7))
	models := map[int64]map[string]string{}
	hashes := map[int64][]byte{}
	model := map[string]string{}

	for cycle := 1; cycle <= 60; cycle++ {
		for i := 0; i < 40; i++ {
			k := fmt.Sprintf("ch%04d", rng.Intn(800))
			switch {
			case rng.Intn(100) < 25 && len(model) > 0:
				if _, ok := model[k]; ok {
					if _, _, err := tree.Remove([]byte(k)); err != nil {
						t.Fatal(err)
					}
					delete(model, k)
				}
			case rng.Intn(100) < 15:
				// same-value rewrite (twin-maker)
				if old, ok := model[k]; ok {
					tree.Set([]byte(k), []byte(old))
					continue
				}
				fallthrough
			default:
				v := fmt.Sprintf("v%d_%d", cycle, i)
				tree.Set([]byte(k), []byte(v))
				model[k] = v
			}
		}
		h, v, err := tree.SaveVersion()
		if err != nil {
			t.Fatal(err)
		}
		snap := make(map[string]string, len(model))
		for k, val := range model {
			snap[k] = val
		}
		models[v], hashes[v] = snap, append([]byte(nil), h...)

		if v > 3 {
			if err := tree.DeleteVersionsTo(v - 3); err != nil {
				t.Fatal(err)
			}
			for pv := range models {
				if pv <= v-3 {
					delete(models, pv)
					delete(hashes, pv)
				}
			}
			assertNoGarbage(t, tree, true)
		}
		if cycle%10 == 0 {
			for rv, m := range models {
				imm, err := tree.GetImmutable(rv)
				if err != nil {
					t.Fatalf("cycle %d: GetImmutable(%d): %v", cycle, rv, err)
				}
				if !bytes.Equal(imm.Hash(), hashes[rv]) {
					t.Fatalf("cycle %d: v%d hash drift", cycle, rv)
				}
				got := 0
				imm.Iterate(func(k, val []byte) bool {
					if m[string(k)] != string(val) {
						t.Fatalf("cycle %d: v%d key %s = %q, want %q", cycle, rv, k, val, m[string(k)])
					}
					got++
					return false
				})
				if got != len(m) {
					t.Fatalf("cycle %d: v%d has %d keys, want %d", cycle, rv, got, len(m))
				}
				imm.Close()
			}
		}
	}

	// Cold restart, verify latest.
	tree2 := NewMutableTreeWithDB(db, 0, NewNopLogger())
	latest, err := tree2.Load()
	if err != nil {
		t.Fatal(err)
	}
	for k, want := range models[latest] {
		got, err := tree2.Get([]byte(k))
		if err != nil || string(got) != want {
			t.Fatalf("cold latest Get(%s) = %q, %v; want %q", k, got, err, want)
		}
	}
}

// TestM20_PruneCoveringLoadedVersion: pruning a range covering the loaded
// (non-latest) version must be refused — the working tree is an unregistered
// reader of it; pruning strictly below it is fine.
func TestM20_PruneCoveringLoadedVersion(t *testing.T) {
	tree := newPruneTree(t)
	for v := 1; v <= 3; v++ {
		for i := 0; i < 20; i++ {
			tree.Set(fmt.Appendf(nil, "lv%03d", i), fmt.Appendf(nil, "v%d", v))
		}
		if _, _, err := tree.SaveVersion(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := tree.LoadVersion(2); err != nil {
		t.Fatal(err)
	}
	if err := tree.PruneVersionsTo(2); !errors.Is(err, ErrActiveReaders) {
		t.Fatalf("covering prune: want ErrActiveReaders, got %v", err)
	}
	if err := tree.PruneVersionsTo(1); err != nil { // strictly below: fine
		t.Fatal(err)
	}
	if got, err := tree.Get([]byte("lv005")); err != nil || string(got) != "v2" {
		t.Fatalf("loaded view broken after below-prune: %q, %v", got, err)
	}
	if _, err := tree.Load(); err != nil {
		t.Fatal(err)
	}
	if err := tree.PruneVersionsTo(2); err != nil {
		t.Fatal(err)
	}
}

// TestImport_RejectsZeroKeyMarkers: zero-key leaf-boundary and inner markers
// are invalid (a legitimate Exporter never emits them; a zero-key saved node
// would break the prune's first-key routing).
func TestImport_RejectsZeroKeyMarkers(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 1000, NewNopLogger())
	imp, err := tree.Import(1)
	if err != nil {
		t.Fatal(err)
	}
	if err := imp.Add(&ExportNode{Height: -1, NumKeys: 0}); err == nil {
		t.Fatal("zero-key leaf boundary accepted")
	}
	if err := imp.Add(&ExportNode{Height: 1, NumKeys: 0}); err == nil {
		t.Fatal("zero-key inner marker accepted")
	}
}
