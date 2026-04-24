package bptree

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestLoad_CleansUpCrashedSessionValues simulates a crash between Set and
// SaveVersion: SaveValue wrote eagerly to DB, sessionValues tracked the
// vks in memory only, and then the process "died". A fresh MutableTree
// constructed over the same DB must clean up the orphan values on Load()
// rather than leaking them forever.
func TestLoad_CleansUpCrashedSessionValues(t *testing.T) {
	db := memdb.NewMemDB()

	// First session: commit v1 cleanly.
	t1 := NewMutableTreeWithDB(db, 100, NewNopLogger())
	t1.Set([]byte("k_saved"), []byte("v_saved"))
	if _, _, err := t1.SaveVersion(); err != nil {
		t.Fatalf("SaveVersion(1): %v", err)
	}
	committed := countDBValues(db)
	if committed != 1 {
		t.Fatalf("setup: committed values = %d, want 1", committed)
	}

	// Second session: Set(s) without SaveVersion, then simulate a crash by
	// abandoning the tree instance (GC'd without Rollback).
	t2 := NewMutableTreeWithDB(db, 100, NewNopLogger())
	if _, err := t2.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	t2.Set([]byte("k_leak1"), []byte("v_leak1"))
	t2.Set([]byte("k_leak2"), []byte("v_leak2"))
	if leaked := countDBValues(db); leaked != committed+2 {
		t.Fatalf("mid-session: values = %d, want %d", leaked, committed+2)
	}
	// Don't call SaveVersion or Rollback — simulate a crash.
	t2 = nil //nolint:wastedassign // simulate process death
	_ = t2

	// Third session: fresh tree over the same DB, Load() must clean orphans.
	t3 := NewMutableTreeWithDB(db, 100, NewNopLogger())
	if _, err := t3.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	after := countDBValues(db)
	if after != committed {
		t.Fatalf("after recovery Load: values = %d, want %d (orphans not cleaned)", after, committed)
	}

	// Committed value still reads back.
	if v, _ := t3.Get([]byte("k_saved")); string(v) != "v_saved" {
		t.Fatalf("Get k_saved after recovery = %q, want v_saved", v)
	}
	// Leaked keys are not in the tree (never committed).
	if v, _ := t3.Get([]byte("k_leak1")); v != nil {
		t.Fatalf("Get k_leak1 = %q, want nil", v)
	}
}

// TestLoad_CleanShutdownNoCleanup verifies cleanupCrashedSessionValues is
// a no-op on a clean shutdown (nothing to clean, no spurious deletes).
func TestLoad_CleanShutdownNoCleanup(t *testing.T) {
	db := memdb.NewMemDB()
	t1 := NewMutableTreeWithDB(db, 100, NewNopLogger())
	for i := byte('a'); i <= byte('j'); i++ {
		t1.Set([]byte{i}, []byte{i, i})
	}
	if _, _, err := t1.SaveVersion(); err != nil {
		t.Fatalf("SaveVersion: %v", err)
	}
	committed := countDBValues(db)

	t2 := NewMutableTreeWithDB(db, 100, NewNopLogger())
	if _, err := t2.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if after := countDBValues(db); after != committed {
		t.Fatalf("Load deleted legitimate values: %d vs %d", after, committed)
	}
}
