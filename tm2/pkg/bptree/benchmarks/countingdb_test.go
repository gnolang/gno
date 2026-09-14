package benchmarks

import (
	"sync/atomic"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
)

// countingDB wraps a dbm.DB and counts the DB operations a tree issues against
// it: reads (Get/Has) and writes (Set/Delete, direct or via a batch). These are
// the node loads/stores per tree op, so they are deterministic (a function of
// tree shape + node LRU), backend-agnostic, and — crucially — unaffected by
// pebble's background flush/compaction, which lives below this interface. That
// makes them a stable basis for the depth gas params, unlike pebble's global
// block-cache miss counter (which folds in bursty compaction I/O).
//
// It embeds dbm.DB so the un-counted methods (Iterator, Close, Stats, ...) pass
// straight through.
type countingDB struct {
	dbm.DB
	reads  atomic.Int64
	writes atomic.Int64
}

func newCountingDB(inner dbm.DB) *countingDB { return &countingDB{DB: inner} }

// stats returns cumulative (reads, writes) since the wrapper was created.
func (c *countingDB) stats() (reads, writes int64) {
	return c.reads.Load(), c.writes.Load()
}

func (c *countingDB) Get(key []byte) ([]byte, error) {
	c.reads.Add(1)
	return c.DB.Get(key)
}

func (c *countingDB) Has(key []byte) (bool, error) {
	c.reads.Add(1)
	return c.DB.Has(key)
}

func (c *countingDB) Set(key, value []byte) error {
	c.writes.Add(1)
	return c.DB.Set(key, value)
}

func (c *countingDB) SetSync(key, value []byte) error {
	c.writes.Add(1)
	return c.DB.SetSync(key, value)
}

func (c *countingDB) Delete(key []byte) error {
	c.writes.Add(1)
	return c.DB.Delete(key)
}

func (c *countingDB) DeleteSync(key []byte) error {
	c.writes.Add(1)
	return c.DB.DeleteSync(key)
}

func (c *countingDB) NewBatch() dbm.Batch {
	return &countingBatch{Batch: c.DB.NewBatch(), writes: &c.writes}
}

func (c *countingDB) NewBatchWithSize(n int) dbm.Batch {
	return &countingBatch{Batch: c.DB.NewBatchWithSize(n), writes: &c.writes}
}

// countingBatch tallies Set/Delete into the parent DB's write counter — batched
// writes are how trees stage their node stores during SaveVersion. It embeds
// dbm.Batch so Write/WriteSync/Close/GetByteSize pass through.
type countingBatch struct {
	dbm.Batch
	writes *atomic.Int64
}

func (b *countingBatch) Set(key, value []byte) error {
	b.writes.Add(1)
	return b.Batch.Set(key, value)
}

func (b *countingBatch) Delete(key []byte) error {
	b.writes.Add(1)
	return b.Batch.Delete(key)
}
