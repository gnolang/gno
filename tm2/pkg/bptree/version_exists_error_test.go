package bptree

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// failingHasDB wraps a DB and forces Has() to return an error for all keys
// whose first byte matches hasFailPrefix. Get/Set/Delete pass through.
type failingHasDB struct {
	dbm.DB
	hasFailPrefix byte
	hasFailErr    error
}

func (d *failingHasDB) Has(key []byte) (bool, error) {
	if len(key) > 0 && key[0] == d.hasFailPrefix {
		return false, d.hasFailErr
	}
	return d.DB.Has(key)
}

// TestSaveVersion_PropagatesVersionExistsError (H10) verifies that a DB error in
// versionExistsE during SaveVersion is propagated to the caller. With the prior
// VersionExists (which swallowed the error and returned false), SaveVersion
// proceeded into the "fresh save" branch and would overwrite an existing
// version with new data.
func TestSaveVersion_PropagatesVersionExistsError(t *testing.T) {
	inner := memdb.NewMemDB()
	wrapped := &failingHasDB{
		DB:            inner,
		hasFailPrefix: PrefixRoot,
		hasFailErr:    errors.New("simulated root Has failure"),
	}
	tree := NewMutableTreeWithDB(wrapped, 100, NewNopLogger())

	if _, err := tree.Set([]byte("k"), []byte("v")); err != nil {
		t.Fatalf("Set: %v", err)
	}

	_, _, err := tree.SaveVersion()
	if err == nil {
		t.Fatalf("SaveVersion should have propagated the DB error")
	}
	if !strings.Contains(err.Error(), "simulated root Has failure") {
		t.Fatalf("error does not wrap underlying cause: %v", err)
	}
}

// failingKeyHasDB forces Has() to error for one exact key while armed.
type failingKeyHasDB struct {
	dbm.DB
	failKey []byte
	armed   bool
}

func (d *failingKeyHasDB) Has(key []byte) (bool, error) {
	if d.armed && bytes.Equal(key, d.failKey) {
		return false, errors.New("simulated Has failure")
	}
	return d.DB.Has(key)
}

// TestPrune_PropagatesVersionExistsError (M22) verifies that a DB error while
// checking version existence during pruning aborts the prune instead of being
// read as "version absent". With the error-swallowing VersionExists, pruning
// took destructive wrong turns: an existing version read as absent was skipped
// while its successor was pruned against a later version (deleting records the
// skipped version still shares), and a skipped RETAINED successor made the
// dual-walk delete records and orphan-listed values that version still uses.
func TestPrune_PropagatesVersionExistsError(t *testing.T) {
	// Two failure sites: the pruneRange existence check on the version being
	// pruned (fail v1, prune to 2) and findNextVersion's successor scan (fail
	// v2, prune to 1).
	cases := []struct {
		name        string
		failVersion int64
		pruneTo     int64
	}{
		{"pruneRangeCheck", 1, 2},
		{"findNextVersion", 2, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inner := memdb.NewMemDB()
			fdb := &failingKeyHasDB{DB: inner}
			tree := NewMutableTreeWithDB(fdb, 100, NewNopLogger())

			// Three versions with shared subtrees: v2 changes only k00,
			// v3 changes only k50, so most leaf records span versions.
			expect := map[int64]map[string]string{1: {}, 2: {}, 3: {}}
			for i := range 100 {
				k := fmt.Sprintf("k%02d", i)
				for v := int64(1); v <= 3; v++ {
					expect[v][k] = "v1"
				}
			}
			set := func(k, val string) {
				t.Helper()
				if _, err := tree.Set([]byte(k), []byte(val)); err != nil {
					t.Fatalf("Set(%s): %v", k, err)
				}
			}
			save := func() {
				t.Helper()
				if _, _, err := tree.SaveVersion(); err != nil {
					t.Fatalf("SaveVersion: %v", err)
				}
			}
			for i := range 100 {
				set(fmt.Sprintf("k%02d", i), "v1")
			}
			save()
			set("k00", "v2")
			expect[2]["k00"], expect[3]["k00"] = "v2", "v2"
			save()
			set("k50", "v3")
			expect[3]["k50"] = "v3"
			save()

			fdb.failKey = rootDBKey(tc.failVersion)
			fdb.armed = true
			err := tree.PruneVersionsTo(tc.pruneTo)
			fdb.armed = false
			if err == nil {
				t.Fatalf("PruneVersionsTo(%d) should have propagated the DB error", tc.pruneTo)
			}
			if !strings.Contains(err.Error(), "simulated Has failure") {
				t.Fatalf("error does not wrap underlying cause: %v", err)
			}

			// Nothing was deleted and the floor did not advance: every
			// version reads back in full from a cold handle.
			fresh := NewMutableTreeWithDB(inner, 100, NewNopLogger())
			if _, err := fresh.Load(); err != nil {
				t.Fatalf("fresh Load: %v", err)
			}
			for v := int64(1); v <= 3; v++ {
				if !fresh.VersionExists(v) {
					t.Fatalf("v%d missing after aborted prune", v)
				}
				for k, want := range expect[v] {
					got, err := fresh.GetVersioned([]byte(k), v)
					if err != nil {
						t.Fatalf("GetVersioned(%s, v%d): %v", k, v, err)
					}
					if string(got) != want {
						t.Fatalf("GetVersioned(%s, v%d) = %q, want %q", k, v, got, want)
					}
				}
			}
		})
	}
}
