package bptree

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"testing"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// Tests for the clean-working-tree fast path: MutableTree.Get serves from the
// committed fast index while the session has no staged mutations (see
// fastReadable), and falls back to the authoritative walk the moment anything
// is staged.
//
// Probe technique: the index is advisory and unauthenticated, so planting a
// valid-checksum entry with a deliberately WRONG value ("doctoring") makes the
// chosen read path observable — Get returning the doctored value proves the
// fast path fired; the real value proves the walk fired. Each probe
// self-verifies at plant time so a malformed probe cannot silently demote a
// fast-path assertion into a vacuous walk test.

// doctorFastEntry plants a valid-checksum fast entry mapping key → wrongValue,
// stamped at version, directly in the DB. It must run after the last
// SaveVersion that touched key (the batch commit would clobber it), and the
// test must not Load afterward (the rebuild clears it).
func doctorFastEntry(t *testing.T, tr *MutableTree, db dbm.DB, key []byte, version int64, wrongValue []byte) {
	t.Helper()
	payload := make([]byte, 8+len(wrongValue))
	binary.BigEndian.PutUint64(payload[:8], uint64(version))
	copy(payload[8:], wrongValue)
	if err := db.Set(fastDBKey(key), stampChecksum(payload)); err != nil {
		t.Fatalf("doctor %q: %v", key, err)
	}
	// fastGet has no feature gate, so this self-check arms even when the tree
	// under test has the index disabled.
	if got, ok := tr.ndb.fastGet(key, version); !ok || !bytes.Equal(got, wrongValue) {
		t.Fatalf("probe for %q did not arm: (%q, %v)", key, got, ok)
	}
}

func mustGet(t *testing.T, tr *MutableTree, k []byte) []byte {
	t.Helper()
	v, err := tr.Get(k)
	if err != nil {
		t.Fatalf("Get(%q): %v", k, err)
	}
	return v
}

// doctorAndAssertServed plants a doctored entry and asserts the clean tree
// serves it — the anti-vacuity check that the fast path is actually open
// before a test moves to the state under test.
func doctorAndAssertServed(t *testing.T, tr *MutableTree, db dbm.DB, key []byte, version int64, wrongValue []byte) {
	t.Helper()
	doctorFastEntry(t, tr, db, key, version, wrongValue)
	if got := mustGet(t, tr, key); !bytes.Equal(got, wrongValue) {
		t.Fatalf("clean Get(%q) = %q; want the doctored value %q (fast path must be open)", key, got, wrongValue)
	}
}

// pumpExportImport drains exp into imp and commits the import.
func pumpExportImport(t *testing.T, exp *Exporter, imp *Importer) {
	t.Helper()
	for {
		node, err := exp.Next()
		if errors.Is(err, ErrExportDone) {
			break
		}
		if err != nil {
			t.Fatalf("Export.Next: %v", err)
		}
		if err := imp.Add(node); err != nil {
			t.Fatalf("Import.Add: %v", err)
		}
	}
	if err := imp.Commit(); err != nil {
		t.Fatalf("Import.Commit: %v", err)
	}
}

// TestFastIndex_WorkingCleanHit: on a clean tree, Get serves from the index.
func TestFastIndex_WorkingCleanHit(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("k"), []byte("real"))
	latest := mustSave(t, tr)

	doctorAndAssertServed(t, tr, db, []byte("k"), latest, []byte("doctored"))
}

// TestFastIndex_WorkingReadYourWrites: any staged mutation of the key routes
// Get back to the authoritative walk, even with a trusted entry in the DB.
func TestFastIndex_WorkingReadYourWrites(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("k"), []byte("v1"))
	latest := mustSave(t, tr)
	doctorAndAssertServed(t, tr, db, []byte("k"), latest, []byte("doctored"))

	// Staged overwrite → walk → staged value.
	mustSet(t, tr, []byte("k"), []byte("v2"))
	if got := mustGet(t, tr, []byte("k")); string(got) != "v2" {
		t.Fatalf("dirty Get = %q; want staged \"v2\"", got)
	}

	// Staged remove → walk → absent, while the doctored entry is still in the DB.
	if _, _, err := tr.Remove([]byte("k")); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if got := mustGet(t, tr, []byte("k")); got != nil {
		t.Fatalf("removed Get = %q; want nil", got)
	}
	if has, _ := tr.Has([]byte("k")); has {
		t.Fatal("removed Has = true; want false")
	}
}

// TestFastIndex_WorkingWholeTreeGate: a staged mutation of ANY key disables
// the fast path for ALL keys (the gate is per-session, not per-key).
func TestFastIndex_WorkingWholeTreeGate(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("k"), []byte("kv"))
	mustSet(t, tr, []byte("other"), []byte("ov"))
	latest := mustSave(t, tr)
	doctorAndAssertServed(t, tr, db, []byte("k"), latest, []byte("doctored"))

	mustSet(t, tr, []byte("other"), []byte("ov2"))
	if got := mustGet(t, tr, []byte("k")); string(got) != "kv" {
		t.Fatalf("Get(k) with another key staged = %q; want the walk value \"kv\"", got)
	}
}

// TestFastIndex_WorkingRollbackResumes: Rollback restores the clean session,
// so the fast path resumes.
func TestFastIndex_WorkingRollbackResumes(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("k"), []byte("real"))
	latest := mustSave(t, tr)
	doctorAndAssertServed(t, tr, db, []byte("k"), latest, []byte("doctored"))

	mustSet(t, tr, []byte("k"), []byte("staged"))
	if got := mustGet(t, tr, []byte("k")); string(got) != "staged" {
		t.Fatalf("dirty Get = %q; want \"staged\"", got)
	}
	tr.Rollback()
	if got := mustGet(t, tr, []byte("k")); string(got) != "doctored" {
		t.Fatalf("post-Rollback Get = %q; want doctored (fast path resumed)", got)
	}
}

// TestFastIndex_WorkingSaveResumes: SaveVersion re-cleans the session; an
// entry for a key the session did not touch survives the save and serves.
func TestFastIndex_WorkingSaveResumes(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("k"), []byte("v1"))
	mustSet(t, tr, []byte("other"), []byte("x"))
	latest := mustSave(t, tr)
	doctorAndAssertServed(t, tr, db, []byte("k"), latest, []byte("doctored"))

	mustSet(t, tr, []byte("other"), []byte("y"))
	if got := mustGet(t, tr, []byte("k")); string(got) != "v1" {
		t.Fatalf("dirty Get = %q; want the walk value \"v1\"", got)
	}
	mustSave(t, tr)
	if got := mustGet(t, tr, []byte("k")); string(got) != "doctored" {
		t.Fatalf("post-save Get = %q; want doctored (fast path resumed, entry untouched)", got)
	}
}

// TestFastIndex_WorkingRemovedKeyAfterSave: a committed remove deletes the
// entry (clobbering anything doctored), so the clean read walks to absent.
func TestFastIndex_WorkingRemovedKeyAfterSave(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("k"), []byte("real"))
	mustSet(t, tr, []byte("keep"), []byte("kv"))
	latest := mustSave(t, tr)
	doctorAndAssertServed(t, tr, db, []byte("k"), latest, []byte("doctored"))

	if _, _, err := tr.Remove([]byte("k")); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	mustSave(t, tr)
	if has, _ := db.Has(fastDBKey([]byte("k"))); has {
		t.Fatal("committed remove left the 'F' entry behind")
	}
	if got := mustGet(t, tr, []byte("k")); got != nil {
		t.Fatalf("clean Get of removed key = %q; want nil", got)
	}
}

// TestFastIndex_WorkingFeatureOff: with the feature off, a trusted-looking
// entry is never consulted.
func TestFastIndex_WorkingFeatureOff(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger()) // FastIndex off
	mustSet(t, tr, []byte("k"), []byte("real"))
	latest := mustSave(t, tr)

	doctorFastEntry(t, tr, db, []byte("k"), latest, []byte("doctored"))
	if got := mustGet(t, tr, []byte("k")); string(got) != "real" {
		t.Fatalf("feature-off Get = %q; want the walk value \"real\"", got)
	}
}

// TestFastIndex_WorkingVersionGuard: after LoadVersion(old), entries stamped
// newer than the loaded version are rejected (walk), entries stamped at or
// before it are trusted — both halves.
func TestFastIndex_WorkingVersionGuard(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("k"), []byte("a"))
	v1 := mustSave(t, tr)
	mustSet(t, tr, []byte("k"), []byte("b"))
	mustSave(t, tr)

	if _, err := tr.LoadVersion(v1); err != nil {
		t.Fatalf("LoadVersion(%d): %v", v1, err)
	}
	// Reject half: the natural entry is stamped v2 > v1 → walk at the v1 root.
	if got := mustGet(t, tr, []byte("k")); string(got) != "a" {
		t.Fatalf("Get at v1 = %q; want \"a\" (v2-stamped entry must be rejected)", got)
	}
	// Trust half: an entry stamped ≤ v1 is served.
	doctorAndAssertServed(t, tr, db, []byte("k"), v1, []byte("doctored"))
}

// TestFastIndex_WorkingEmptyValue: the fast path preserves the
// present-with-empty-value vs absent distinction (non-nil empty).
func TestFastIndex_WorkingEmptyValue(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("e"), []byte{})
	mustSet(t, tr, []byte("canary"), []byte("c"))
	latest := mustSave(t, tr)

	// Canary proves the gate is open for this tree state.
	doctorAndAssertServed(t, tr, db, []byte("canary"), latest, []byte("doctored"))
	got := mustGet(t, tr, []byte("e"))
	if got == nil || len(got) != 0 {
		t.Fatalf("present-empty via fast path = %#v; want non-nil empty", got)
	}
	if got := mustGet(t, tr, []byte("absent")); got != nil {
		t.Fatalf("absent = %q; want nil", got)
	}
}

// TestFastIndex_WorkingCommittedEmpty: on a committed-empty tree the nil-root
// return keeps Get/Has off the index entirely — including after a staged
// Set+Remove round-trip that brings root back to nil == lastSaved.
func TestFastIndex_WorkingCommittedEmpty(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	v1 := mustSave(t, tr) // commit the empty tree

	doctorFastEntry(t, tr, db, []byte("k"), v1, []byte("phantom"))
	if got := mustGet(t, tr, []byte("k")); got != nil {
		t.Fatalf("empty-tree Get = %q; want nil (index must not fabricate keys)", got)
	}
	if has, _ := tr.Has([]byte("k")); has {
		t.Fatal("empty-tree Has = true; want false")
	}

	// Staged round-trip: root goes leaf → nil again (== lastSaved == nil).
	mustSet(t, tr, []byte("k"), []byte("x"))
	if _, _, err := tr.Remove([]byte("k")); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if got := mustGet(t, tr, []byte("k")); got != nil {
		t.Fatalf("post-round-trip Get = %q; want nil", got)
	}
	if has, _ := tr.Has([]byte("k")); has {
		t.Fatal("post-round-trip Has = true; want false")
	}
}

// TestFastIndex_ImportClearsStaleEntries: Import drops pre-existing 'F'
// entries and the stamp up front, so reads between Importer.Commit and the
// next Load walk the imported tree instead of trusting pre-import values.
func TestFastIndex_ImportClearsStaleEntries(t *testing.T) {
	// Destination with prior index-on history: natural 'F' entries exist.
	dstDB := memdb.NewMemDB()
	dst := NewMutableTreeWithDB(dstDB, 256, NewNopLogger(), FastIndexOption(true))
	for i := range 40 {
		mustSet(t, dst, fmt.Appendf(nil, "k%03d", i), fmt.Appendf(nil, "old%d", i))
	}
	dstVer := mustSave(t, dst)
	if n := countFastEntries(t, dstDB); n != 40 {
		t.Fatalf("precondition: expected 40 'F' entries, got %d", n)
	}

	// Source with the same keys but different values.
	srcDB := memdb.NewMemDB()
	src := NewMutableTreeWithDB(srcDB, 256, NewNopLogger())
	for i := range 40 {
		mustSet(t, src, fmt.Appendf(nil, "k%03d", i), fmt.Appendf(nil, "new%d", i))
	}
	srcVer := mustSave(t, src)
	imm, err := src.GetImmutable(srcVer)
	if err != nil {
		t.Fatalf("GetImmutable: %v", err)
	}
	defer imm.Close()
	exp, err := imm.Export(src.ndb)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	defer exp.Close()

	imp, err := dst.Import(dstVer + 1)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	// The clear happens in Import itself, before any nodes are staged.
	if n := countFastEntries(t, dstDB); n != 0 {
		t.Fatalf("Import left %d stale 'F' entries; want 0", n)
	}
	if _, ok, _ := dst.ndb.getFastIndexVersion(); ok {
		t.Fatal("Import left the fast-index stamp; an aborted import would skip the rebuild")
	}
	pumpExportImport(t, exp, imp)

	// Clean tree, gate open, index empty → every read walks the imported tree.
	if n := countFastEntries(t, dstDB); n != 0 {
		t.Fatalf("post-Commit: %d 'F' entries; want 0", n)
	}
	for i := range 40 {
		k := fmt.Appendf(nil, "k%03d", i)
		if got := mustGet(t, dst, k); string(got) != fmt.Sprintf("new%d", i) {
			t.Fatalf("post-import Get(%q) = %q; want %q (stale pre-import value served?)", k, got, "new"+fmt.Sprint(i))
		}
	}
}

// TestFastIndex_WorkingNaturalEntryServed: a NATURALLY maintained entry
// (built by setFastIndex's single-buffer seal path, not a test-doctored one)
// is trusted and served by the working-tree fast path. Proven by deleting
// the authoritative out-of-line value record: only the index's inline copy
// can still produce the value, so a walk — or a format break in the
// natural-entry writer — fails loudly instead of passing vacuously.
func TestFastIndex_WorkingNaturalEntryServed(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("k"), []byte("real"))
	mustSave(t, tr)

	_, _, vk, found, err := treeLookup(tr.root, []byte("k"))
	if err != nil || !found {
		t.Fatalf("treeLookup: err=%v found=%v", err, found)
	}
	if err := db.Delete(valueDBKey(vk)); err != nil {
		t.Fatalf("delete value record: %v", err)
	}
	if got := mustGet(t, tr, []byte("k")); string(got) != "real" {
		t.Fatalf("Get = %q; want \"real\" served from the natural fast-index entry", got)
	}
}
