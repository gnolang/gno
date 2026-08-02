package bptree

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestVerifyFastIndex_Healthy: a normally-maintained index reports current and
// clean.
func TestVerifyFastIndex_Healthy(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("a"), []byte("va"))
	mustSet(t, tr, []byte("b"), []byte("vb"))
	v := mustSave(t, tr)

	rep, err := tr.VerifyFastIndex()
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Healthy() {
		t.Fatalf("healthy index reported unhealthy: %+v", rep)
	}
	if rep.Entries != 2 || rep.MismatchCount != 0 {
		t.Fatalf("entries=%d mismatches=%d; want 2, 0", rep.Entries, rep.MismatchCount)
	}
	if !rep.StampPresent || rep.Stamp != v || rep.Version != v {
		t.Fatalf("stamp/version = (%v, %d, %d); want (true, %d, %d)", rep.StampPresent, rep.Stamp, rep.Version, v, v)
	}
}

// TestVerifyFastIndex_StaleValue: a stamp-current index whose entry disagrees
// with the tree — the gno#6011 signature — is caught and classified.
func TestVerifyFastIndex_StaleValue(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("k"), []byte("authoritative"))
	v := mustSave(t, tr)

	// Freeze a wrong inline value for a key the tree still holds, at the current
	// stamp (as a stale query-path rebuild did in #6011).
	doctorFastEntry(t, tr, db, []byte("k"), v, []byte("stale"))

	rep, err := tr.VerifyFastIndex()
	if err != nil {
		t.Fatal(err)
	}
	if rep.Healthy() {
		t.Fatal("stale entry not detected: report says healthy")
	}
	if rep.MismatchCount != 1 || len(rep.Mismatches) != 1 {
		t.Fatalf("mismatches = %d (sample %d); want 1", rep.MismatchCount, len(rep.Mismatches))
	}
	m := rep.Mismatches[0]
	if m.Kind != MismatchStaleValue || string(m.Key) != "k" || m.FastVersion != v || m.TreeVersion != v {
		t.Fatalf("mismatch = %+v; want stale-value key=k fast=%d tree=%d", m, v, v)
	}
}

// TestVerifyFastIndex_Orphan: an 'F' entry for a key absent from the tree is
// reported as an orphan (a leftover for a removed key).
func TestVerifyFastIndex_Orphan(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("real"), []byte("v"))
	v := mustSave(t, tr)

	doctorFastEntry(t, tr, db, []byte("ghost"), v, []byte("x"))

	rep, err := tr.VerifyFastIndex()
	if err != nil {
		t.Fatal(err)
	}
	if rep.Healthy() || rep.MismatchCount != 1 {
		t.Fatalf("orphan not detected: %+v", rep)
	}
	if m := rep.Mismatches[0]; m.Kind != MismatchOrphan || string(m.Key) != "ghost" {
		t.Fatalf("mismatch = %+v; want orphan key=ghost", m)
	}
}

// TestVerifyFastIndex_ReadOnly: verifying does not mutate persisted state — the
// 'F' entry count and stamp are unchanged after a run over a doctored DB.
func TestVerifyFastIndex_ReadOnly(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	mustSet(t, tr, []byte("k"), []byte("v"))
	v := mustSave(t, tr)
	doctorFastEntry(t, tr, db, []byte("k"), v, []byte("stale"))

	before := countFastEntries(t, db)
	if _, err := tr.VerifyFastIndex(); err != nil {
		t.Fatal(err)
	}
	if after := countFastEntries(t, db); after != before {
		t.Fatalf("VerifyFastIndex changed 'F' entry count: %d -> %d", before, after)
	}
	// The doctored (stale) entry must still be there — verify must not "heal".
	if got, ok := tr.ndb.fastGet([]byte("k"), v); !ok || string(got) != "stale" {
		t.Fatalf("VerifyFastIndex mutated the index: fastGet = (%q, %v)", got, ok)
	}
}
