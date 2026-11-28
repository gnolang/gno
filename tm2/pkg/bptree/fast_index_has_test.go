package bptree

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestFastIndex_HasVersionGuard pins the version guard on the Has fast path.
// The index always holds the LATEST value for a key, so a snapshot older than
// the key's first write must not be told the key exists. fastHas has to apply
// the same "entry newer than the snapshot ⇒ untrusted" rule as fastGet; for
// Has the consequence of skipping it is a plain wrong answer (true for a key
// that did not exist yet) rather than a merely stale value.
func TestFastIndex_HasVersionGuard(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	defer tr.Close()

	mustSet(t, tr, []byte("old"), []byte("1"))
	v1 := mustSave(t, tr)
	mustSet(t, tr, []byte("new"), []byte("2"))
	v2 := mustSave(t, tr)

	// "new" has an 'F' entry stamped v2. A v1 snapshot must say absent.
	imm1, err := tr.GetImmutable(v1)
	if err != nil {
		t.Fatalf("GetImmutable(v1): %v", err)
	}
	defer imm1.Close()
	if has, err := imm1.Has([]byte("new")); err != nil || has {
		t.Fatalf("v1 snapshot Has(new) = %v, %v; want false", has, err)
	}
	if has, err := imm1.Has([]byte("old")); err != nil || !has {
		t.Fatalf("v1 snapshot Has(old) = %v, %v; want true", has, err)
	}

	imm2, err := tr.GetImmutable(v2)
	if err != nil {
		t.Fatalf("GetImmutable(v2): %v", err)
	}
	defer imm2.Close()
	for _, k := range []string{"old", "new"} {
		if has, err := imm2.Has([]byte(k)); err != nil || !has {
			t.Fatalf("v2 snapshot Has(%s) = %v, %v; want true", k, has, err)
		}
	}
	if has, err := imm2.Has([]byte("absent")); err != nil || has {
		t.Fatalf("v2 snapshot Has(absent) = %v, %v; want false", has, err)
	}
}

// TestFastIndex_HasAfterRemoveAndDirtySession covers the two remaining ways a
// stale index entry could leak into Has: a removed key (whose entry Remove
// must delete) and a dirty working session (where the fast path must be off
// entirely because staged writes are not in the index's committed view).
func TestFastIndex_HasAfterRemoveAndDirtySession(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	defer tr.Close()

	mustSet(t, tr, []byte("a"), []byte("1"))
	mustSet(t, tr, []byte("b"), []byte("2"))
	mustSave(t, tr)

	if _, _, err := tr.Remove([]byte("a")); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	mustSave(t, tr)
	if has, err := tr.Has([]byte("a")); err != nil || has {
		t.Fatalf("Has(a) after remove = %v, %v; want false", has, err)
	}

	// Dirty session: a staged insert must be visible, and the fast path must
	// not be consulted at all (fastReadable is false while root != lastSaved).
	mustSet(t, tr, []byte("c"), []byte("3"))
	for _, tc := range []struct {
		key  string
		want bool
	}{{"a", false}, {"b", true}, {"c", true}, {"zz", false}} {
		if has, err := tr.Has([]byte(tc.key)); err != nil || has != tc.want {
			t.Fatalf("dirty Has(%s) = %v, %v; want %v", tc.key, has, err, tc.want)
		}
	}
}
