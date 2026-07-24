package db

import "sync"

// BatchCollector accumulates Set/Delete operations in memory without touching
// disk. It backs CollectingDB and lets rootmulti fuse multiple sub-store write
// sites (dbadapter flush, IAVL SaveVersion, rootmulti metadata) into one
// atomic disk write: each writer appends via a batchHandle, and the caller
// drains the accumulated ops into a real Batch at the end of the commit.
//
// pending indexes the last op per key so CollectingDB can serve read-your-
// writes without touching the underlying DB — required when a caller writes
// and then reads back before the next drain (e.g. deploying a package in one
// tx and importing it in the next, before the block-level Commit runs).
//
// Set/Delete/Get are safe for concurrent use. Reset and Drain must not race
// with writers; they are called before/after the commit window when no writer
// is active.
type BatchCollector struct {
	mu      sync.Mutex
	ops     []collectOp
	pending map[string]int // key → index into ops of the latest op for that key
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

// Reset drops all accumulated ops. Call at the start of a commit window.
func (c *BatchCollector) Reset() {
	c.mu.Lock()
	c.ops = c.ops[:0]
	clear(c.pending)
	c.mu.Unlock()
}

// Drain replays every collected op into dst in the order recorded, then clears
// the collector. The caller flushes dst (typically via WriteSync) to persist
// them atomically.
func (c *BatchCollector) Drain(dst Batch) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, op := range c.ops {
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
	c.ops = c.ops[:0]
	clear(c.pending)
	return nil
}

// get returns the value the collector currently has recorded for key. found is
// true when the collector has any op for key; when it also has del=true the
// caller must treat the key as absent (a pending delete masks any real value).
func (c *BatchCollector) get(key []byte) (val []byte, del, found bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	idx, ok := c.pending[string(key)]
	if !ok {
		return nil, false, false
	}
	op := c.ops[idx]
	return op.val, op.del, true
}

// Len returns the number of ops currently buffered.
func (c *BatchCollector) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.ops)
}

// NewBatch returns a Batch that appends into this collector. Callers use it
// to write into the same op-log as CollectingDB-wrapped sub-stores; it lets
// rootmulti feed its own metadata (commitInfo, latestVersion) into the same
// atomic drain as IAVL and dbadapter writes.
func (c *BatchCollector) NewBatch() Batch {
	return &batchHandle{collector: c}
}

func (c *BatchCollector) set(key, value []byte) {
	// Copy inputs — callers (BatchWithFlusher, cache flush) reuse buffers.
	k := append([]byte(nil), key...)
	v := append([]byte(nil), value...)
	c.mu.Lock()
	c.ops = append(c.ops, collectOp{key: k, val: v})
	c.pending[string(k)] = len(c.ops) - 1
	c.mu.Unlock()
}

func (c *BatchCollector) delete(key []byte) {
	k := append([]byte(nil), key...)
	c.mu.Lock()
	c.ops = append(c.ops, collectOp{del: true, key: k})
	c.pending[string(k)] = len(c.ops) - 1
	c.mu.Unlock()
}

func (c *BatchCollector) byteSize() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := 0
	for _, op := range c.ops {
		n += len(op.key) + len(op.val)
	}
	return n
}

// CollectingDB wraps a real DB and routes every write through a shared
// BatchCollector while reads pass through untouched. It is installed under
// rootmulti's sub-stores so that IAVL, dbadapter, and rootmulti's own metadata
// writes accumulate in one place during CommitAll and get flushed as a single
// atomic batch.
//
// Reads always hit the underlying real DB — the collector never serves reads.
// This is safe because:
//   - IAVL's SaveVersion never re-reads a node it just wrote in the same call
//     (newly-created nodes live in ndb.nodeCache, which GetNode checks before
//     falling through to db.Get).
//   - dbadapter's cache flush uses NewBatch()/batch.Set/batch.Write and never
//     reads back its own uncommitted writes.
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
// window, so this is safe today; revisit if that changes.
func (c *CollectingDB) Iterator(start, end []byte) (Iterator, error) {
	return c.real.Iterator(start, end)
}

func (c *CollectingDB) ReverseIterator(start, end []byte) (Iterator, error) {
	return c.real.ReverseIterator(start, end)
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

// NewBatch and NewBatchWithSize return a Batch whose Set/Delete route into the
// same collector as direct writes. Write/WriteSync/Close are no-ops so IAVL's
// BatchWithFlusher can auto-flush freely without losing ops or forcing early
// disk writes — the collector is drained externally when the commit closes.
func (c *CollectingDB) NewBatch() Batch            { return &batchHandle{collector: c.collector} }
func (c *CollectingDB) NewBatchWithSize(int) Batch { return &batchHandle{collector: c.collector} }

// batchHandle is a Batch that forwards Set/Delete into a shared BatchCollector.
// Write/WriteSync/Close return nil without persisting anything — the collector
// is drained externally at the end of the commit window.
type batchHandle struct {
	collector *BatchCollector
}

var _ Batch = (*batchHandle)(nil)

func (b *batchHandle) Set(key, value []byte) error { b.collector.set(key, value); return nil }
func (b *batchHandle) Delete(key []byte) error     { b.collector.delete(key); return nil }
func (b *batchHandle) Write() error                { return nil }
func (b *batchHandle) WriteSync() error            { return nil }
func (b *batchHandle) Close() error                { return nil }
func (b *batchHandle) GetByteSize() (int, error)   { return b.collector.byteSize(), nil }
