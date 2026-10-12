package db_test

import (
	"testing"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestImmutableDB_BatchIsUsable pins that ImmutableDB hands out a usable
// read-only no-op batch rather than nil: consumers construct batches eagerly
// (bptree's nodeDB, IAVL's BatchWithFlusher) and Set/Close them on read-only
// load paths, which must not nil-deref. ImmutableDB is the query-path
// fallback on backends without snapshot support, so a nil batch would panic
// every query there. Writes must still fail loud.
func TestImmutableDB_BatchIsUsable(t *testing.T) {
	idb := dbm.NewImmutableDB(memdb.NewMemDB())

	for _, b := range []dbm.Batch{idb.NewBatch(), idb.NewBatchWithSize(16)} {
		if b == nil {
			t.Fatal("ImmutableDB returned a nil batch")
		}
		if err := b.Set([]byte("k"), []byte("v")); err != nil {
			t.Fatalf("staging on a read-only batch must be a silent no-op: %v", err)
		}
		if err := b.Delete([]byte("k")); err != nil {
			t.Fatalf("staging on a read-only batch must be a silent no-op: %v", err)
		}
		if sz, err := b.GetByteSize(); err != nil || sz != 0 {
			t.Fatalf("GetByteSize = %d, %v", sz, err)
		}
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("Write on a read-only batch must panic")
				}
			}()
			_ = b.Write()
		}()
		if err := b.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}
	}
}
