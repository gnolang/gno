package bptree

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// countFastEntries returns the number of 'F' (fast-index) records in db.
func countFastEntries(t *testing.T, db dbm.DB) int {
	t.Helper()
	itr, err := db.Iterator([]byte{PrefixFast}, []byte{PrefixFast + 1})
	if err != nil {
		t.Fatalf("iterator: %v", err)
	}
	defer itr.Close()
	n := 0
	for ; itr.Valid(); itr.Next() {
		n++
	}
	if err := itr.Error(); err != nil {
		t.Fatalf("iterator error: %v", err)
	}
	return n
}

func mustSet(t *testing.T, tr *MutableTree, k, v []byte) {
	t.Helper()
	if _, err := tr.Set(k, v); err != nil {
		t.Fatalf("Set(%q): %v", k, err)
	}
}

func mustSave(t *testing.T, tr *MutableTree) int64 {
	t.Helper()
	_, v, err := tr.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion: %v", err)
	}
	return v
}

// TestFastIndex_OffByDefault: with the feature off, no 'F' records are written
// and reads are correct (byte-identical to today).
func TestFastIndex_OffByDefault(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger()) // default: FastIndex off
	mustSet(t, tr, []byte("k"), []byte("v"))
	latest := mustSave(t, tr)

	if n := countFastEntries(t, db); n != 0 {
		t.Fatalf("FastIndex off but wrote %d 'F' entries", n)
	}
	got, err := tr.GetVersioned([]byte("k"), latest)
	if err != nil || string(got) != "v" {
		t.Fatalf("GetVersioned = %q, %v; want \"v\"", got, err)
	}
}

// TestFastIndex_Encoding locks the inline entry format: version(8) ‖ value.
func TestFastIndex_Encoding(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("k"), []byte("hello"))
	latest := mustSave(t, tr)

	data, err := db.Get(fastDBKey([]byte("k")))
	if err != nil || data == nil {
		t.Fatalf("Get 'F'k = %v, %v", data, err)
	}
	payload, err := verifyChecksum(data)
	if err != nil {
		t.Fatalf("verifyChecksum: %v", err)
	}
	if len(payload) != 8+len("hello") {
		t.Fatalf("payload len = %d; want %d (version+value)", len(payload), 8+len("hello"))
	}
	if v := int64(binary.BigEndian.Uint64(payload[:8])); v != latest {
		t.Fatalf("inline version = %d; want %d", v, latest)
	}
	if string(payload[8:]) != "hello" {
		t.Fatalf("inline value = %q; want \"hello\"", payload[8:])
	}
}

// TestFastIndex_Differential is the linchpin: drive a random Set/Remove/Save/
// Prune stream against two trees (index ON vs OFF) on separate DBs, and assert
// every GetVersioned agrees — for present AND absent keys — across every
// retained version, plus that the app hash is identical (merkle invariance).
func TestFastIndex_Differential(t *testing.T) {
	rng := rand.New(rand.NewSource(0x5eed))
	dbA, dbB := memdb.NewMemDB(), memdb.NewMemDB()
	tA := NewMutableTreeWithDB(dbA, 64, NewNopLogger(), FastIndexOption(true))
	tB := NewMutableTreeWithDB(dbB, 64, NewNopLogger())

	const keyspace = 160
	key := func(i int) []byte { return []byte(fmt.Sprintf("k%05d", i)) }

	model := map[string][]byte{}
	snaps := map[int64]map[string][]byte{}
	var firstRetained, latest int64
	saves := 0

	for op := 0; op < 4000; op++ {
		if rng.Intn(3) == 0 && len(model) > 0 {
			i := rng.Intn(keyspace)
			k := key(i)
			if _, ok := model[string(k)]; ok {
				if _, _, err := tA.Remove(k); err != nil {
					t.Fatalf("tA.Remove: %v", err)
				}
				if _, _, err := tB.Remove(k); err != nil {
					t.Fatalf("tB.Remove: %v", err)
				}
				delete(model, string(k))
			}
		} else {
			i := rng.Intn(keyspace)
			k := key(i)
			v := []byte(fmt.Sprintf("v%d.%d", i, op))
			mustSet(t, tA, k, v)
			mustSet(t, tB, k, v)
			model[string(k)] = v
		}

		if op%17 == 16 {
			hA, vA, err := tA.SaveVersion()
			if err != nil {
				t.Fatalf("tA.SaveVersion: %v", err)
			}
			hB, vB, err := tB.SaveVersion()
			if err != nil {
				t.Fatalf("tB.SaveVersion: %v", err)
			}
			if vA != vB {
				t.Fatalf("version mismatch %d vs %d", vA, vB)
			}
			if !bytes.Equal(hA, hB) {
				t.Fatalf("app hash differs with index on/off at v%d", vA)
			}
			latest = vA
			if firstRetained == 0 {
				firstRetained = vA
			}
			snap := make(map[string][]byte, len(model))
			for k, v := range model {
				snap[k] = v
			}
			snaps[vA] = snap
			saves++

			if rng.Intn(4) == 0 && latest-firstRetained > 3 {
				to := firstRetained + 1
				if err := tA.PruneVersionsTo(to); err != nil {
					t.Fatalf("tA.Prune(%d): %v", to, err)
				}
				if err := tB.PruneVersionsTo(to); err != nil {
					t.Fatalf("tB.Prune(%d): %v", to, err)
				}
				for v := firstRetained; v <= to; v++ {
					delete(snaps, v)
				}
				firstRetained = to + 1
			}
		}
	}

	checks := 0
	for v, snap := range snaps {
		for k, want := range snap {
			gotA, errA := tA.GetVersioned([]byte(k), v)
			gotB, errB := tB.GetVersioned([]byte(k), v)
			if errA != nil || errB != nil {
				t.Fatalf("v%d key %q: errA=%v errB=%v", v, k, errA, errB)
			}
			if !bytes.Equal(gotA, gotB) {
				t.Fatalf("v%d key %q: index-on=%q index-off=%q", v, k, gotA, gotB)
			}
			if !bytes.Equal(gotA, want) {
				t.Fatalf("v%d key %q: got %q want %q", v, k, gotA, want)
			}
			checks++
		}
		for i := 0; i < 40; i++ {
			ak := []byte(fmt.Sprintf("absent-%d", i))
			gotA, _ := tA.GetVersioned(ak, v)
			gotB, _ := tB.GetVersioned(ak, v)
			if gotA != nil || gotB != nil {
				t.Fatalf("v%d absent key %q: index-on=%q index-off=%q", v, ak, gotA, gotB)
			}
		}
	}

	if saves < 10 {
		t.Fatalf("only %d saves — test not exercising enough", saves)
	}
	if checks == 0 {
		t.Fatal("no present-key checks ran")
	}
	if n := countFastEntries(t, dbA); n == 0 {
		t.Fatal("index-on tree has no 'F' entries — fast path never populated")
	}
	if n := countFastEntries(t, dbB); n != 0 {
		t.Fatalf("index-off tree wrote %d 'F' entries", n)
	}
}

// TestFastIndex_EagerStampAvoidsRebuild: eager maintenance must advance the
// version stamp on SaveVersion, so reopening an index-on DB finds it current and
// skips the (potentially huge) rebuild.
func TestFastIndex_EagerStampAvoidsRebuild(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("a"), []byte("1"))
	mustSet(t, tr, []byte("b"), []byte("2"))
	latest := mustSave(t, tr)

	stamp, ok, err := tr.ndb.getFastIndexVersion()
	if err != nil || !ok || stamp != latest {
		t.Fatalf("stamp after eager save = (%d,%v,%v); want (%d,true,nil)", stamp, ok, err, latest)
	}

	// Reopen: ensureFastIndex sees stamp == latest and is a no-op.
	tr2 := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	if _, err := tr2.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got, _ := tr2.GetVersioned([]byte("a"), latest); string(got) != "1" {
		t.Fatalf("'a' = %q; want \"1\"", got)
	}
	if s2, _, _ := tr2.ndb.getFastIndexVersion(); s2 != latest {
		t.Fatalf("stamp after reopen = %d; want %d", s2, latest)
	}
}

// TestFastIndex_VersionCheckRejectsNewer: a committed snapshot at S must never
// see a value written after S, even though the index points at the newer entry.
func TestFastIndex_VersionCheckRejectsNewer(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	k := []byte("key")

	mustSet(t, tr, k, []byte("v1"))
	v1 := mustSave(t, tr)
	mustSet(t, tr, k, []byte("v2"))
	v2 := mustSave(t, tr)

	if got, err := tr.GetVersioned(k, v1); err != nil || string(got) != "v1" {
		t.Fatalf("GetVersioned(v1) = %q, %v; want \"v1\"", got, err)
	}
	if got, err := tr.GetVersioned(k, v2); err != nil || string(got) != "v2" {
		t.Fatalf("GetVersioned(v2) = %q, %v; want \"v2\"", got, err)
	}
}

// TestFastIndex_Rollback: keys staged then rolled back leave no 'F' entry.
func TestFastIndex_Rollback(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("committed"), []byte("x"))
	latest := mustSave(t, tr)

	mustSet(t, tr, []byte("rolledback"), []byte("y"))
	tr.Rollback()

	if has, _ := db.Has(fastDBKey([]byte("rolledback"))); has {
		t.Fatal("rolled-back key has an 'F' entry")
	}
	if got, _ := tr.GetVersioned([]byte("committed"), latest); string(got) != "x" {
		t.Fatalf("committed key = %q; want \"x\"", got)
	}
}

// TestFastIndex_RebuildOnLoad: a DB written WITHOUT the index, reopened WITH it,
// rebuilds the index on Load; latest reads are correct and removed keys absent.
func TestFastIndex_RebuildOnLoad(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger()) // no index
	mustSet(t, tr, []byte("a"), []byte("1"))
	mustSet(t, tr, []byte("b"), []byte("2"))
	mustSave(t, tr)
	mustSet(t, tr, []byte("b"), []byte("22"))
	if _, _, err := tr.Remove([]byte("a")); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	mustSet(t, tr, []byte("c"), []byte("3"))
	latest := mustSave(t, tr)
	if n := countFastEntries(t, db); n != 0 {
		t.Fatalf("expected 0 'F' entries before rebuild, got %d", n)
	}

	// Reopen WITH the index → Load rebuilds.
	tr2 := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	v, err := tr2.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if v != latest {
		t.Fatalf("Load returned v%d; want %d", v, latest)
	}
	if n := countFastEntries(t, db); n != 2 { // live keys: b, c
		t.Fatalf("after rebuild expected 2 'F' entries, got %d", n)
	}
	if got, _ := tr2.GetVersioned([]byte("a"), latest); got != nil {
		t.Fatalf("removed key 'a' = %q; want nil", got)
	}
	if got, _ := tr2.GetVersioned([]byte("b"), latest); string(got) != "22" {
		t.Fatalf("'b' = %q; want \"22\"", got)
	}
	if got, _ := tr2.GetVersioned([]byte("c"), latest); string(got) != "3" {
		t.Fatalf("'c' = %q; want \"3\"", got)
	}

	// A second Load must NOT rebuild (stamp == latest).
	tr3 := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	if _, err := tr3.Load(); err != nil {
		t.Fatalf("second Load: %v", err)
	}
	if n := countFastEntries(t, db); n != 2 {
		t.Fatalf("second Load changed 'F' count to %d; want 2", n)
	}
}

// TestFastIndex_ImportRebuildsOnLoad: Import bypasses per-entry maintenance, so
// the index must be empty AND unstamped right after import (else Load would trust
// an empty index); the next Load then rebuilds it.
func TestFastIndex_ImportRebuildsOnLoad(t *testing.T) {
	// Build a source tree and export it.
	srcDB := memdb.NewMemDB()
	src := NewMutableTreeWithDB(srcDB, 256, NewNopLogger())
	for i := 0; i < 40; i++ {
		mustSet(t, src, []byte(fmt.Sprintf("k%03d", i)), []byte(fmt.Sprintf("v%d", i)))
	}
	ver := mustSave(t, src)
	imm, err := src.GetImmutable(ver)
	if err != nil {
		t.Fatalf("GetImmutable: %v", err)
	}
	defer imm.Close()
	exp, err := imm.Export(src.ndb)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	defer exp.Close()

	// Import into a fresh DB WITH the fast index enabled.
	dstDB := memdb.NewMemDB()
	dst := NewMutableTreeWithDB(dstDB, 256, NewNopLogger(), FastIndexOption(true))
	imp, err := dst.Import(ver)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	for {
		node, nerr := exp.Next()
		if errors.Is(nerr, ErrExportDone) {
			break
		}
		if nerr != nil {
			t.Fatalf("Export.Next: %v", nerr)
		}
		if aerr := imp.Add(node); aerr != nil {
			t.Fatalf("Import.Add: %v", aerr)
		}
	}
	if cerr := imp.Commit(); cerr != nil {
		t.Fatalf("Import.Commit: %v", cerr)
	}

	// Empty and unstamped right after import.
	if n := countFastEntries(t, dstDB); n != 0 {
		t.Fatalf("expected 0 'F' entries after import, got %d", n)
	}
	if _, ok, _ := dst.ndb.getFastIndexVersion(); ok {
		t.Fatal("import left a fast-index stamp; Load would skip the rebuild")
	}

	// Reopen + Load → rebuild fires and serves correct values.
	dst2 := NewMutableTreeWithDB(dstDB, 256, NewNopLogger(), FastIndexOption(true))
	if _, err := dst2.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if n := countFastEntries(t, dstDB); n != 40 {
		t.Fatalf("Load did not rebuild after import: %d 'F' entries (want 40)", n)
	}
	for i := 0; i < 40; i++ {
		got, _ := dst2.GetVersioned([]byte(fmt.Sprintf("k%03d", i)), ver)
		if want := fmt.Sprintf("v%d", i); string(got) != want {
			t.Fatalf("k%03d = %q; want %q", i, got, want)
		}
	}
	fastValueEqualitySweep(t, dstDB, dst2)
}

// TestFastIndex_RebuildClearsStale: a stale 'F' entry for a key not in the tree
// is cleared by the rebuild (the stale-present route the clear defends against).
func TestFastIndex_RebuildClearsStale(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger()) // no index
	mustSet(t, tr, []byte("real"), []byte("v"))
	latest := mustSave(t, tr)

	// Inject a stale ghost entry, as if from a prior index incarnation.
	ghostVK := (&NodeKey{Version: latest, Nonce: 999}).GetKey()
	if err := db.Set(fastDBKey([]byte("ghost")), stampChecksum(ghostVK)); err != nil {
		t.Fatalf("inject: %v", err)
	}

	tr2 := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	if _, err := tr2.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if has, _ := db.Has(fastDBKey([]byte("ghost"))); has {
		t.Fatal("rebuild did not clear the stale ghost entry")
	}
	if got, _ := tr2.GetVersioned([]byte("real"), latest); string(got) != "v" {
		t.Fatalf("'real' = %q; want \"v\"", got)
	}
}

// TestFastIndex_Prune: pruning leaves the latest index intact and non-dangling —
// every surviving 'F' entry resolves to a live value.
func TestFastIndex_Prune(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	key := func(i int) []byte { return []byte(fmt.Sprintf("k%03d", i)) }

	var latest int64
	for v := 1; v <= 5; v++ {
		for i := 0; i < 25; i++ {
			mustSet(t, tr, key(i), []byte(fmt.Sprintf("v%d-k%d", v, i)))
		}
		latest = mustSave(t, tr)
	}
	if err := tr.PruneVersionsTo(3); err != nil {
		t.Fatalf("Prune: %v", err)
	}

	for i := 0; i < 25; i++ {
		want := fmt.Sprintf("v5-k%d", i)
		if got, _ := tr.GetVersioned(key(i), latest); string(got) != want {
			t.Fatalf("after prune key %d = %q; want %q", i, got, want)
		}
	}

	fastValueEqualitySweep(t, db, tr)
}

// fastValueEqualitySweep asserts every 'F' entry's inlined value byte-equals the
// tree's authoritative value for that key — the inline replacement for the
// pointer no-dangling check.
//
// It resolves the expected value via MutableTree.Get (the tree walk, which never
// consults the fast index), NOT via GetVersioned / the committed ImmutableTree
// (whose fast path would read the very same 'F' entry, making the check
// circular). So a wrong inline value written by setFastIndex is genuinely caught.
// tr must be at the latest committed version (the version the index reflects),
// which all callers are.
func fastValueEqualitySweep(t *testing.T, db dbm.DB, tr *MutableTree) {
	t.Helper()
	itr, err := db.Iterator([]byte{PrefixFast}, []byte{PrefixFast + 1})
	if err != nil {
		t.Fatalf("iterator: %v", err)
	}
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		payload, err := verifyChecksum(itr.Value())
		if err != nil || len(payload) < 8 {
			t.Fatalf("corrupt 'F' entry %x: %v", itr.Key(), err)
		}
		userKey := append([]byte(nil), itr.Key()[1:]...) // strip PrefixFast
		want, err := tr.Get(userKey)                     // index-free tree walk → non-circular
		if err != nil {
			t.Fatalf("tr.Get(%q): %v", userKey, err)
		}
		if got := payload[8:]; !bytes.Equal(got, want) {
			t.Fatalf("'F'%q inline value %q != tree value %q", userKey, got, want)
		}
	}
}

// TestFastIndex_EmptyValue: present-with-empty-value is distinguished from absent.
func TestFastIndex_EmptyValue(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("empty"), []byte{})
	mustSet(t, tr, []byte("full"), []byte("x"))
	latest := mustSave(t, tr)

	got, err := tr.GetVersioned([]byte("empty"), latest)
	if err != nil {
		t.Fatalf("empty: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("present-empty = %#v; want non-nil empty", got)
	}
	if got, _ := tr.GetVersioned([]byte("absent"), latest); got != nil {
		t.Fatalf("absent = %q; want nil", got)
	}
}

// TestFastIndex_CorruptEntrySelfHeals: a corrupt 'F' entry falls back to the
// authoritative tree walk rather than erroring or returning garbage.
func TestFastIndex_CorruptEntrySelfHeals(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("k"), []byte("v"))
	latest := mustSave(t, tr)

	if err := db.Set(fastDBKey([]byte("k")), []byte("garbage")); err != nil {
		t.Fatalf("corrupt: %v", err)
	}
	got, err := tr.GetVersioned([]byte("k"), latest)
	if err != nil {
		t.Fatalf("corrupt entry should self-heal, got err %v", err)
	}
	if string(got) != "v" {
		t.Fatalf("got %q; want \"v\" (tree-walk fallback)", got)
	}
}

// TestFastIndex_Has: Has is correct (it stays on the tree walk; the index must
// not change its answers) with the index enabled.
func TestFastIndex_Has(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("k"), []byte("v"))
	mustSave(t, tr)

	if has, _ := tr.Has([]byte("k")); !has {
		t.Fatal("Has(k) = false; want true")
	}
	if has, _ := tr.Has([]byte("absent")); has {
		t.Fatal("Has(absent) = true; want false")
	}
	imm, err := tr.GetImmutable(tr.Version())
	if err != nil {
		t.Fatalf("GetImmutable: %v", err)
	}
	defer imm.Close()
	if has, _ := imm.Has([]byte("k")); !has {
		t.Fatal("imm.Has(k) = false; want true")
	}
	if has, _ := imm.Has([]byte("absent")); has {
		t.Fatal("imm.Has(absent) = true; want false")
	}
}
