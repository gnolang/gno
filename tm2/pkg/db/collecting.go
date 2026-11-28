package db

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// ErrIteratorOverHeldWrites is the panic value of an iterator opened on a
// CollectingDB while its collector holds writes across blocks: the iterator
// reads only the underlying database, so it would miss them.
var ErrIteratorOverHeldWrites = errors.New("db: iterator would miss the writes held in memory")

// BatchCollector accumulates Set/Delete operations in memory without touching
// disk. It backs CollectingDB and lets rootmulti fuse multiple sub-store write
// sites (dbadapter flush, IAVL/bptree SaveVersion, rootmulti metadata) into
// one atomic disk write: direct DB writes enter the collector immediately,
// batch writes enter when their batch is Written (see batchHandle), and
// WriteTo writes the accumulated ops as one synced batch, at the end of a
// commit or, while holding, at a flush point.
//
// pending indexes the last op per key so CollectingDB can serve read-your-
// writes without touching the underlying DB — required when a caller writes
// and then reads back before the next drain (e.g. deploying a package in one
// tx and importing it in the next, before the block-level Commit runs).
//
// Set/Delete/Get are safe for concurrent use. WriteTo, Drain and Reset must
// not race with writers; they run when no writer is active.
//
// While holding (SetHolding), the collector keeps its ops across commits and
// CollectingDB's iterators refuse to run (see ErrIteratorOverHeldWrites); a
// refused iterator is recorded for TakeRefusedIterator.
type BatchCollector struct {
	mu      sync.RWMutex
	ops     []collectOp
	pending map[string]int // key → index into ops of the latest op for that key
	size    int            // sum of the ops' key and value lengths

	holding atomic.Bool
	refused atomic.Bool
}

type collectOp struct {
	del bool // true = Delete, false = Set
	key []byte
	val []byte // nil for Delete
}

// NewBatchCollector returns an empty collector.
func NewBatchCollector() *BatchCollector {
	return &BatchCollector{pending: make(map[string]int)}
}

// Reset drops all accumulated ops. Call it once the batch they were drained
// into is written. While holding it also frees the memory they held; outside
// it keeps it for the next commit.
func (c *BatchCollector) Reset() {
	c.mu.Lock()
	if c.holding.Load() {
		// A held window can reach 64 MiB: give its memory back.
		c.ops = nil
		c.pending = make(map[string]int)
	} else {
		clear(c.ops)
		c.ops = c.ops[:0]
		clear(c.pending)
	}
	c.size = 0
	c.mu.Unlock()
}

// Drain replays the collected ops into dst in the order recorded, but for an
// op a later one on the same key overrides: dst is written as one batch, so
// only the last op on each key can reach the disk. The caller flushes dst
// (typically via WriteSync) to persist them atomically, then calls Reset:
// until then the ops stay readable through CollectingDB, so a concurrent
// reader finds each record in the collector or on disk. WriteTo does all of
// it.
func (c *BatchCollector) Drain(dst Batch) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for i, op := range c.ops {
		if c.pending[string(op.key)] != i {
			continue
		}
		if op.del {
			if err := dst.Delete(op.key); err != nil {
				return err
			}
			continue
		}
		if err := dst.Set(op.key, op.val); err != nil {
			return err
		}
	}
	return nil
}

// WriteTo writes the collected ops to db as one synced batch (see Drain), and
// only then drops them.
func (c *BatchCollector) WriteTo(db DB) error {
	batch := db.NewBatchWithSize(c.Size())
	defer batch.Close()
	if err := c.Drain(batch); err != nil {
		return err
	}
	if err := batch.WriteSync(); err != nil {
		return err
	}
	c.Reset()
	return nil
}

// get returns the value the collector currently has recorded for key. found is
// true when the collector has any op for key; when it also has del=true the
// caller must treat the key as absent (a pending delete masks any real value).
func (c *BatchCollector) get(key []byte) (val []byte, del, found bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	idx, ok := c.pending[string(key)]
	if !ok {
		return nil, false, false
	}
	op := c.ops[idx]
	return op.val, op.del, true
}

// Len returns the number of ops currently buffered.
func (c *BatchCollector) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.ops)
}

// Size returns the bytes of keys and values currently buffered, counting
// every op, overwrites included.
func (c *BatchCollector) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.size
}

// SetHolding turns holding on or off; see BatchCollector. It panics when
// turned on or off with ops collected: on, they would be held as part of the
// commits that follow; off, they would be lost, or written with later ones.
func (c *BatchCollector) SetHolding(on bool) {
	if n := c.Len(); n > 0 {
		panic(fmt.Sprintf("db: holding turned on=%t with %d writes collected", on, n))
	}
	c.holding.Store(on)
}

// Holding reports whether the collector holds its ops across commits.
func (c *BatchCollector) Holding() bool { return c.holding.Load() }

// TakeRefusedIterator reports whether an iterator was refused since the last
// call, and clears the record.
func (c *BatchCollector) TakeRefusedIterator() bool { return c.refused.Swap(false) }

// NewBatch returns a Batch that flushes into this collector on Write. Callers
// use it to write into the same op-log as CollectingDB-wrapped sub-stores; it
// lets rootmulti feed its own metadata (commitInfo, latestVersion) into the
// same atomic drain as IAVL and dbadapter writes.
func (c *BatchCollector) NewBatch() Batch {
	return &batchHandle{collector: c}
}

// appendOps moves a batch's staged ops into the collector, preserving order.
// Ownership of ops (and their key/val backing) transfers to the collector;
// callers must not reuse them.
func (c *BatchCollector) appendOps(ops []collectOp) {
	if len(ops) == 0 {
		return
	}
	c.mu.Lock()
	for _, op := range ops {
		c.ops = append(c.ops, op)
		c.pending[string(op.key)] = len(c.ops) - 1
		c.size += len(op.key) + len(op.val)
	}
	c.mu.Unlock()
}

func (c *BatchCollector) set(key, value []byte) {
	// Copy inputs — callers (BatchWithFlusher, cache flush) reuse buffers. cp,
	// not append: an empty value must stay non-nil, or Get reports it missing.
	c.appendOps([]collectOp{{key: cp(key), val: cp(value)}})
}

func (c *BatchCollector) delete(key []byte) {
	c.appendOps([]collectOp{{del: true, key: cp(key)}})
}

// CollectingDB wraps a real DB and routes every write through a shared
// BatchCollector while iterators pass through untouched. It is installed under
// rootmulti's sub-stores so that IAVL/bptree, dbadapter, and rootmulti's own
// metadata writes accumulate in one place during CommitAll and get flushed as
// a single atomic batch.
//
// Point reads (Get/Has) consult the collector first for read-your-writes over
// ops already in it — direct Sets/Deletes land there immediately, batch ops
// once their batch is Written. Iterators read the underlying DB ONLY (see
// Iterator below); consumers that iterate over keys they stage must not
// depend on pending ops being visible (bptree's clearFastIndex resumes by
// key for exactly this reason).
type CollectingDB struct {
	real      DB
	collector *BatchCollector
}

var _ DB = (*CollectingDB)(nil)

// NewCollectingDB returns a DB whose writes accumulate in c and whose reads
// pass through to db. Reads and writes may run concurrently.
func NewCollectingDB(db DB, c *BatchCollector) *CollectingDB {
	return &CollectingDB{real: db, collector: c}
}

// Get consults the collector first for read-your-writes, then falls through
// to the real DB. A pending Delete masks any real value for the key.
func (c *CollectingDB) Get(key []byte) ([]byte, error) {
	if val, del, found := c.collector.get(key); found {
		if del {
			return nil, nil
		}
		return val, nil
	}
	return c.real.Get(key)
}

// Has mirrors Get: a pending Set makes the key present, a pending Delete
// makes it absent regardless of real DB state, and otherwise we defer to the
// real DB.
func (c *CollectingDB) Has(key []byte) (bool, error) {
	if _, del, found := c.collector.get(key); found {
		return !del, nil
	}
	return c.real.Has(key)
}

// Iterator and ReverseIterator do NOT merge pending writes with the real DB.
// Callers that iterate before the next Drain will miss pending Sets and see
// stale keys through pending Deletes. Current consumers (IAVL SaveVersion,
// dbadapter cache flush, rootmulti metadata) don't iterate during a commit
// window, so this is safe today; revisit if that changes. While the
// collector holds writes across commits they panic instead, with
// ErrIteratorOverHeldWrites, and record it.
func (c *CollectingDB) Iterator(start, end []byte) (Iterator, error) {
	c.refuseWhileHolding()
	return c.real.Iterator(start, end)
}

func (c *CollectingDB) ReverseIterator(start, end []byte) (Iterator, error) {
	c.refuseWhileHolding()
	return c.real.ReverseIterator(start, end)
}

func (c *CollectingDB) refuseWhileHolding() {
	if c.collector.Holding() {
		c.collector.refused.Store(true)
		panic(ErrIteratorOverHeldWrites)
	}
}

func (c *CollectingDB) NewSnapshot() (Snapshot, error) { return c.real.NewSnapshot() }
func (c *CollectingDB) Print() error                   { return c.real.Print() }
func (c *CollectingDB) Stats() map[string]string       { return c.real.Stats() }

// Close is a no-op — the caller owns the real DB.
func (c *CollectingDB) Close() error { return nil }

// Direct writes route into the collector. SetSync/DeleteSync collapse to Set/
// Delete because the collector is drained under an explicit WriteSync at the
// end of the commit.
func (c *CollectingDB) Set(key, value []byte) error     { c.collector.set(key, value); return nil }
func (c *CollectingDB) SetSync(key, value []byte) error { c.collector.set(key, value); return nil }
func (c *CollectingDB) Delete(key []byte) error         { c.collector.delete(key); return nil }
func (c *CollectingDB) DeleteSync(key []byte) error     { c.collector.delete(key); return nil }

// NewBatch and NewBatchWithSize return a Batch that buffers Set/Delete locally
// and moves them into the collector on Write/WriteSync (no disk I/O — the
// collector is drained externally when the commit closes). Close without Write
// DISCARDS the buffered ops, honoring the standard dbm.Batch contract: callers
// like bptree's DiscardBatch rely on Close dropping staged-but-uncommitted
// writes (session rollback, error paths). IAVL's BatchWithFlusher auto-flush
// (Write, Close, NewBatch cycles) works unchanged.
func (c *CollectingDB) NewBatch() Batch            { return &batchHandle{collector: c.collector} }
func (c *CollectingDB) NewBatchWithSize(int) Batch { return &batchHandle{collector: c.collector} }

// batchHandle is a Batch buffering ops locally until Write/WriteSync moves
// them into the shared BatchCollector (order preserved). Close discards the
// buffer. Unlike real backends (whose contract allows only Close after
// Write), the handle is reusable after Write or Close — leniency, not a
// contract to depend on. Like every dbm.Batch, it is NOT safe for concurrent
// use (the local buffer is unsynchronized).
type batchHandle struct {
	collector *BatchCollector
	ops       []collectOp
	size      int // running sum of buffered key+val bytes (O(1) GetByteSize)
}

var _ Batch = (*batchHandle)(nil)

func (b *batchHandle) Set(key, value []byte) error {
	// Copy inputs — callers (BatchWithFlusher, cache flush) reuse buffers.
	k, v := cp(key), cp(value)
	b.ops = append(b.ops, collectOp{key: k, val: v})
	b.size += len(k) + len(v)
	return nil
}

func (b *batchHandle) Delete(key []byte) error {
	k := cp(key)
	b.ops = append(b.ops, collectOp{del: true, key: k})
	b.size += len(k)
	return nil
}

func (b *batchHandle) Write() error {
	b.collector.appendOps(b.ops)
	b.ops = nil
	b.size = 0
	return nil
}

func (b *batchHandle) WriteSync() error { return b.Write() }

func (b *batchHandle) Close() error {
	b.ops = nil
	b.size = 0
	return nil
}

// GetByteSize returns the size of THIS batch's buffered ops (not the whole
// collector), so size-threshold flush loops (bptree prune, BatchWithFlusher)
// measure their own staging. O(1): BatchWithFlusher calls it on every op.
func (b *batchHandle) GetByteSize() (int, error) {
	return b.size, nil
}
