package db_test

import (
	"bytes"
	"errors"
	"fmt"
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

	// Drain applies everything in order to a destination: the delete staged
	// after the direct set must win. It leaves the collector as it was, so
	// readers keep finding the ops until the destination is written; Reset
	// then clears it.
	dst := memdb.NewMemDB()
	dstBatch := dst.NewBatch()
	defer dstBatch.Close()
	if err := c.Drain(dstBatch); err != nil {
		t.Fatal(err)
	}
	if c.Len() != 3 {
		t.Fatalf("Drain changed the collector: len = %d", c.Len())
	}
	if v, _ := cdb.Get([]byte("k1")); !bytes.Equal(v, []byte("v1")) {
		t.Fatalf("op not readable between Drain and Reset: %q", v)
	}
	if err := dstBatch.Write(); err != nil {
		t.Fatal(err)
	}
	c.Reset()
	if c.Len() != 0 || c.Size() != 0 {
		t.Fatalf("collector not cleared by Reset: len %d, size %d", c.Len(), c.Size())
	}
	if v, _ := dst.Get([]byte("direct")); v != nil {
		t.Fatalf("drain order broken: deleted key present: %q", v)
	}
	if v, _ := dst.Get([]byte("k1")); !bytes.Equal(v, []byte("v1")) {
		t.Fatalf("drained value wrong: %q", v)
	}
}

// TestCollectingDB_EmptyValue pins that a key set to an empty value reads as
// present and empty, as it does from every backend once written, and not as
// missing: an empty block's results encode to no bytes, and the state store
// reports a missing record as "no results for height".
func TestCollectingDB_EmptyValue(t *testing.T) {
	c := dbm.NewBatchCollector()
	cdb := dbm.NewCollectingDB(memdb.NewMemDB(), c)

	if err := cdb.Set([]byte("empty"), []byte{}); err != nil {
		t.Fatal(err)
	}
	v, err := cdb.Get([]byte("empty"))
	if err != nil {
		t.Fatal(err)
	}
	if v == nil || len(v) != 0 {
		t.Fatalf("empty value read back as %#v, want a non-nil empty slice", v)
	}
	if ok, _ := cdb.Has([]byte("empty")); !ok {
		t.Fatal("empty value reported missing by Has")
	}
}

// TestBatchCollector_Size pins that Size counts every op's key and value,
// overwrites included: every op stays in memory until the write.
func TestBatchCollector_Size(t *testing.T) {
	c := dbm.NewBatchCollector()
	cdb := dbm.NewCollectingDB(memdb.NewMemDB(), c)

	cdb.Set([]byte("ab"), []byte("cde"))  // 5
	cdb.Set([]byte("ab"), []byte("fghi")) // 6, the overwrite counts
	cdb.Delete([]byte("xyz"))             // 3
	b := cdb.NewBatch()
	b.Set([]byte("k"), []byte("vv")) // 3, counted once written
	if got := c.Size(); got != 14 {
		t.Fatalf("size before the batch is written: %d, want 14", got)
	}
	if err := b.Write(); err != nil {
		t.Fatal(err)
	}
	if got := c.Size(); got != 17 {
		t.Fatalf("size: %d, want 17", got)
	}
	c.Reset()
	if got := c.Size(); got != 0 {
		t.Fatalf("size after Reset: %d", got)
	}
}

// TestBatchCollector_WriteToKeepsMemory pins that, outside holding, a write
// keeps the collector's memory for the next commit: every consensus commit
// goes through WriteTo, and one after the first must not grow the op log and
// the key index again.
func TestBatchCollector_WriteToKeepsMemory(t *testing.T) {
	const runs = 5
	keys := make([][]byte, 1000)
	for i := range keys {
		keys[i] = fmt.Appendf(nil, "key-%04d", i)
	}
	db := memdb.NewMemDB()
	commit := func(c *dbm.BatchCollector) {
		cdb := dbm.NewCollectingDB(db, c)
		for _, k := range keys {
			cdb.Set(k, k)
		}
		if err := c.WriteTo(db); err != nil {
			t.Fatal(err)
		}
	}
	commit(dbm.NewBatchCollector()) // the keys are on disk for every run below

	// AllocsPerRun calls its function runs+1 times.
	fresh := make([]*dbm.BatchCollector, runs+1)
	for i := range fresh {
		fresh[i] = dbm.NewBatchCollector()
	}
	first := testing.AllocsPerRun(runs, func() {
		commit(fresh[0])
		fresh = fresh[1:]
	})
	c := dbm.NewBatchCollector()
	later := testing.AllocsPerRun(runs, func() { commit(c) })
	// Growing a log of 1,000 ops from nothing reallocates it at least 10 times.
	if first-later < 10 {
		t.Fatalf("a later commit allocates %v times, a first one %v: WriteTo dropped the collector's memory", later, first)
	}
}

// TestCollectingDB_IteratorRefusedWhileHolding pins that, while the collector
// holds writes across commits, an iterator panics instead of silently reading
// only the underlying database, and that the refusal is recorded once.
func TestCollectingDB_IteratorRefusedWhileHolding(t *testing.T) {
	c := dbm.NewBatchCollector()
	cdb := dbm.NewCollectingDB(memdb.NewMemDB(), c)

	mustPanic := func(name string, open func()) {
		t.Helper()
		defer func() {
			err, _ := recover().(error)
			if !errors.Is(err, dbm.ErrIteratorOverHeldWrites) {
				t.Fatalf("%s while holding: recovered %v, want ErrIteratorOverHeldWrites", name, err)
			}
		}()
		open()
	}

	c.SetHolding(true)
	cdb.Set([]byte("k"), []byte("v"))
	mustPanic("Iterator", func() { cdb.Iterator(nil, nil) })
	if !c.TakeRefusedIterator() {
		t.Fatal("refused iterator not recorded")
	}
	if c.TakeRefusedIterator() {
		t.Fatal("record not cleared by TakeRefusedIterator")
	}
	mustPanic("ReverseIterator", func() { cdb.ReverseIterator(nil, nil) })
	if !c.TakeRefusedIterator() {
		t.Fatal("refused reverse iterator not recorded")
	}

	c.Reset()
	c.SetHolding(false)
	it, err := cdb.Iterator(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	it.Close()
	if c.TakeRefusedIterator() {
		t.Fatal("iterator recorded as refused while not holding")
	}
}

// TestBatchCollector_WriteTo pins that WriteTo writes only the last op on
// each key, as one synced batch, keeps every op readable while the batch is
// being written, and only then drops them.
func TestBatchCollector_WriteTo(t *testing.T) {
	c := dbm.NewBatchCollector()
	cdb := dbm.NewCollectingDB(memdb.NewMemDB(), c)
	cdb.Set([]byte("a"), []byte("1"))
	cdb.Set([]byte("b"), []byte("1"))
	cdb.Set([]byte("a"), []byte("2")) // overrides the first
	cdb.Delete([]byte("b"))           // overrides the second
	cdb.Set([]byte("b"), []byte("3")) // overrides the delete

	disk := memdb.NewMemDB()
	db := &syncLog{DB: disk, onBatchSync: func() error {
		if v, _ := cdb.Get([]byte("a")); string(v) != "2" {
			t.Fatalf("op unreadable while its batch is written: %q", v)
		}
		return nil
	}}
	if err := c.WriteTo(db); err != nil {
		t.Fatal(err)
	}
	if got, want := fmt.Sprint(db.sets), "[a=2 b=3]"; got != want {
		t.Fatalf("batch got %s, want %s", got, want)
	}
	if db.batchSyncs != 1 {
		t.Fatalf("%d synced batches, want 1", db.batchSyncs)
	}
	if c.Len() != 0 {
		t.Fatalf("ops kept after WriteTo: %d", c.Len())
	}
	if v, _ := disk.Get([]byte("b")); string(v) != "3" {
		t.Fatalf("disk has b=%q, want 3", v)
	}
}

// TestBatchCollector_HoldingSwitchWithOpsPanics pins that holding cannot be
// turned off with ops held, which would lose them or write them late, nor on
// with ops collected, which would hold them as part of the commits after.
func TestBatchCollector_HoldingSwitchWithOpsPanics(t *testing.T) {
	for _, on := range []bool{false, true} {
		c := dbm.NewBatchCollector()
		c.SetHolding(!on)
		dbm.NewCollectingDB(memdb.NewMemDB(), c).Set([]byte("a"), []byte("1"))
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("SetHolding(%t) with ops collected did not panic", on)
				}
			}()
			c.SetHolding(on)
		}()
	}
}
