package db_test

import (
	"bytes"
	"testing"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestCollectingDB_BatchSemantics pins the batchHandle contract the store
// stack relies on: ops are buffered locally, become collector-visible (and
// read-your-writes-visible) only on Write, are DISCARDED by Close without
// Write (bptree's DiscardBatch depends on this for session rollback), the
// handle is reusable, and GetByteSize measures this batch only.
func TestCollectingDB_BatchSemantics(t *testing.T) {
	c := dbm.NewBatchCollector()
	cdb := dbm.NewCollectingDB(memdb.NewMemDB(), c)

	// Direct writes are visible immediately (read-your-writes).
	if err := cdb.Set([]byte("direct"), []byte("d1")); err != nil {
		t.Fatal(err)
	}
	if v, err := cdb.Get([]byte("direct")); err != nil || !bytes.Equal(v, []byte("d1")) {
		t.Fatalf("direct set not visible: %q %v", v, err)
	}
	if c.Len() != 1 {
		t.Fatalf("collector len = %d, want 1", c.Len())
	}

	// Batch ops are buffered: invisible to reads and the collector pre-Write.
	b := cdb.NewBatch()
	if err := b.Set([]byte("k1"), []byte("v1")); err != nil {
		t.Fatal(err)
	}
	if err := b.Delete([]byte("direct")); err != nil {
		t.Fatal(err)
	}
	if c.Len() != 1 {
		t.Fatalf("staged batch ops leaked into collector: len = %d", c.Len())
	}
	if v, _ := cdb.Get([]byte("k1")); v != nil {
		t.Fatalf("unwritten batch op visible: %q", v)
	}
	if sz, err := b.GetByteSize(); err != nil || sz != len("k1")+len("v1")+len("direct") {
		t.Fatalf("GetByteSize = %d, %v", sz, err)
	}

	// Write flushes in order; reads see the result (delete masks direct set).
	if err := b.Write(); err != nil {
		t.Fatal(err)
	}
	if c.Len() != 3 {
		t.Fatalf("collector len = %d, want 3", c.Len())
	}
	if v, _ := cdb.Get([]byte("k1")); !bytes.Equal(v, []byte("v1")) {
		t.Fatalf("written batch op not visible: %q", v)
	}
	if v, _ := cdb.Get([]byte("direct")); v != nil {
		t.Fatalf("pending delete not masking: %q", v)
	}

	// The handle is empty and reusable after Write.
	if sz, _ := b.GetByteSize(); sz != 0 {
		t.Fatalf("buffer not cleared by Write: size %d", sz)
	}
	if err := b.Set([]byte("k2"), []byte("v2")); err != nil {
		t.Fatal(err)
	}

	// Close without Write DISCARDS the buffer.
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	if err := b.Write(); err != nil { // reusable post-Close; nothing to flush
		t.Fatal(err)
	}
	if c.Len() != 3 {
		t.Fatalf("Close leaked staged ops: len = %d", c.Len())
	}
	if v, _ := cdb.Get([]byte("k2")); v != nil {
		t.Fatalf("discarded op visible: %q", v)
	}

	// Drain applies everything in order to a destination and clears the
	// collector: the delete staged after the direct set must win.
	dst := memdb.NewMemDB()
	dstBatch := dst.NewBatch()
	defer dstBatch.Close()
	if err := c.Drain(dstBatch); err != nil {
		t.Fatal(err)
	}
	if err := dstBatch.Write(); err != nil {
		t.Fatal(err)
	}
	if c.Len() != 0 {
		t.Fatalf("collector not cleared by Drain: %d", c.Len())
	}
	if v, _ := dst.Get([]byte("direct")); v != nil {
		t.Fatalf("drain order broken: deleted key present: %q", v)
	}
	if v, _ := dst.Get([]byte("k1")); !bytes.Equal(v, []byte("v1")) {
		t.Fatalf("drained value wrong: %q", v)
	}
}
