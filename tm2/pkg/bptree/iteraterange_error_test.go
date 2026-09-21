package bptree

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// N19: IterateRange must surface value-resolution and node-load failures as
// an error — previously it called fn with (key, nil) and returned
// false-as-"completed", so a corrupt value read as a clean short result.
// fn must never see the failing row.
func TestIterateRange_PropagatesValueError(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 0, NewNopLogger())
	for i := range 50 {
		if _, err := tree.Set(fmt.Appendf(nil, "key%03d", i), []byte("v")); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}

	// Destroy one value record inside the range.
	vkKey := func() []byte {
		_, _, vk, found, err := treeLookup(tree.root, []byte("key025"))
		if err != nil || !found {
			t.Fatalf("setup: %v", err)
		}
		return valueDBKey(vk)
	}()
	if err := db.Delete(vkKey); err != nil {
		t.Fatal(err)
	}

	check := func(name string, run func(fn func(key, value []byte) bool) (bool, error)) {
		t.Helper()
		var sawNilValue bool
		rows := 0
		_, err := run(func(key, value []byte) bool {
			rows++
			if value == nil {
				sawNilValue = true
			}
			return false
		})
		if err == nil {
			t.Fatalf("%s: corrupt value read as clean completion after %d rows", name, rows)
		}
		if sawNilValue {
			t.Fatalf("%s: fn was called with the failing row's nil value", name)
		}
	}

	check("MutableTree", func(fn func(key, value []byte) bool) (bool, error) {
		return tree.IterateRange([]byte("key000"), []byte("key049"), true, fn)
	})

	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatal(err)
	}
	defer imm.Close()
	check("ImmutableTree", func(fn func(key, value []byte) bool) (bool, error) {
		return imm.IterateRange([]byte("key000"), []byte("key049"), true, fn)
	})

	// Early-stop before the failing row still works and reports no error.
	rows := 0
	stopped, err := tree.IterateRange([]byte("key000"), []byte("key049"), true, func(key, value []byte) bool {
		rows++
		return rows >= 3
	})
	if err != nil || !stopped || rows != 3 {
		t.Fatalf("early stop before failure: stopped=%v rows=%d err=%v", stopped, rows, err)
	}
}
