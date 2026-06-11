package bptree

import (
	"fmt"
	"strings"
	"testing"

	bp "github.com/gnolang/gno/tm2/pkg/bptree"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// The store wrapper enforces iterator-error acknowledgment: a consumer that
// never checks Error() must NOT see a failed iteration as normal exhaustion
// (silent truncation on the consensus path) — it panics at Valid() or Close().
// Deliberate handlers that read Error() are never interrupted.

func buildCorruptIterStore(t *testing.T) *Store {
	t.Helper()
	db := memdb.NewMemDB()
	tree := bp.NewMutableTreeWithDB(db, 0, bp.NewNopLogger())
	st := UnsafeNewStore(tree, types.StoreOptions{})
	for i := range 200 {
		st.Set(nil, fmt.Appendf(nil, "key%03d", i), []byte("v"))
	}
	st.Commit()

	// Corrupt one non-root node record.
	rootRaw, err := db.Get([]byte{'R', 0, 0, 0, 0, 0, 0, 0, 1})
	if err != nil || rootRaw == nil {
		t.Fatalf("root record: %v", err)
	}
	rootNK := string(rootRaw[:12])
	itr, err := db.Iterator([]byte{'B'}, []byte{'B' + 1})
	if err != nil {
		t.Fatal(err)
	}
	var key, val []byte
	for ; itr.Valid(); itr.Next() {
		if string(itr.Key()[1:]) == rootNK {
			continue
		}
		key = append([]byte(nil), itr.Key()...)
		val = append([]byte(nil), itr.Value()...)
		break
	}
	itr.Close()
	if key == nil {
		t.Fatal("no non-root node record")
	}
	val[len(val)/2] ^= 0x01
	if err := db.Set(key, val); err != nil {
		t.Fatal(err)
	}

	// Fresh store over the corrupted DB (no cache).
	tree2 := bp.NewMutableTreeWithDB(db, 0, bp.NewNopLogger())
	st2 := UnsafeNewStore(tree2, types.StoreOptions{})
	if err := st2.LoadLatestVersion(); err != nil {
		t.Fatal(err)
	}
	return st2
}

func TestStoreIterator_UncheckedErrorPanics(t *testing.T) {
	st := buildCorruptIterStore(t)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("draining a failed iterator without checking Error() must panic")
		}
		if !strings.Contains(fmt.Sprint(r), "never checked") {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	itr := st.Iterator(nil, nil, nil)
	rows := 0
	for ; itr.Valid(); itr.Next() { // Error() never consulted → Valid() panics
		rows++
	}
	t.Fatalf("iterator exhausted silently after %d rows", rows)
}

func TestStoreIterator_CheckedErrorNoPanic(t *testing.T) {
	st := buildCorruptIterStore(t)
	itr := st.Iterator(nil, nil, nil)
	for itr.Error() == nil && itr.Valid() {
		itr.Next()
	}
	if itr.Error() == nil {
		t.Fatal("expected an iteration error")
	}
	// Acknowledged: neither a further Error() read nor Close panics.
	if err := itr.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestStoreIterator_HealthyNoPanic(t *testing.T) {
	db := memdb.NewMemDB()
	tree := bp.NewMutableTreeWithDB(db, 100, bp.NewNopLogger())
	st := UnsafeNewStore(tree, types.StoreOptions{})
	for i := range 50 {
		st.Set(nil, fmt.Appendf(nil, "key%03d", i), []byte("v"))
	}
	st.Commit()

	itr := st.Iterator(nil, nil, nil)
	rows := 0
	for ; itr.Valid(); itr.Next() {
		rows++
	}
	if rows != 50 {
		t.Fatalf("rows = %d, want 50", rows)
	}
	if err := itr.Close(); err != nil {
		t.Fatal(err)
	}
}
