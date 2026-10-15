package db

// HoldingDB is a DB that passes every call through to the database it wraps
// until Hold(true). From then on it keeps every write in memory, where Get and
// Has see it, until WriteHeld writes all of them as one synced batch; its
// iterators panic meanwhile (see ErrIteratorOverHeldWrites). Block sync uses
// it to write the records of many blocks at once (tm2/adr, deferred writes
// during block sync).
//
// Hold and WriteHeld are called by the one goroutine that writes, between its
// writes. Reads may run concurrently at any time: a held write stays readable
// until it is on disk.
type HoldingDB struct {
	c *CollectingDB // its collector holds the writes; c.real is the database
}

var _ DB = (*HoldingDB)(nil)

// NewHoldingDB returns a HoldingDB over db, passing every call through.
func NewHoldingDB(db DB) *HoldingDB {
	return &HoldingDB{c: NewCollectingDB(db, NewBatchCollector())}
}

// Hold turns holding on or off. It panics when turned off with writes still
// held: they would be lost, or written after later ones. (Nothing is collected
// while not holding, so turning it on finds the collector empty.)
func (h *HoldingDB) Hold(on bool) { h.c.collector.SetHolding(on) }

// Holding reports whether writes are held.
func (h *HoldingDB) Holding() bool { return h.c.collector.Holding() }

// HeldBytes returns the bytes of keys and values held, every write counted.
func (h *HoldingDB) HeldBytes() int { return h.c.collector.Size() }

// WriteHeld writes every held write as one batch with a sync, and only then
// drops them from memory.
func (h *HoldingDB) WriteHeld() error { return h.c.collector.WriteTo(h.c.real) }

// Get and Has read held writes first, then the database.
func (h *HoldingDB) Get(key []byte) ([]byte, error) {
	if !h.Holding() {
		return h.c.real.Get(key) // nothing is held: Hold(false) requires it
	}
	return h.c.Get(key)
}

func (h *HoldingDB) Has(key []byte) (bool, error) {
	if !h.Holding() {
		return h.c.real.Has(key)
	}
	return h.c.Has(key)
}

func (h *HoldingDB) Set(key, value []byte) error     { return h.writer().Set(key, value) }
func (h *HoldingDB) SetSync(key, value []byte) error { return h.writer().SetSync(key, value) }
func (h *HoldingDB) Delete(key []byte) error         { return h.writer().Delete(key) }
func (h *HoldingDB) DeleteSync(key []byte) error     { return h.writer().DeleteSync(key) }
func (h *HoldingDB) NewBatch() Batch                 { return h.writer().NewBatch() }
func (h *HoldingDB) NewBatchWithSize(size int) Batch { return h.writer().NewBatchWithSize(size) }
func (h *HoldingDB) NewSnapshot() (Snapshot, error)  { return h.c.real.NewSnapshot() }
func (h *HoldingDB) Print() error                    { return h.c.real.Print() }
func (h *HoldingDB) Stats() map[string]string        { return h.c.real.Stats() }
func (h *HoldingDB) Close() error                    { return h.c.real.Close() }

// Iterator and ReverseIterator read the database, and panic while writes are
// held, which they would miss (see CollectingDB).
func (h *HoldingDB) Iterator(start, end []byte) (Iterator, error) { return h.c.Iterator(start, end) }

func (h *HoldingDB) ReverseIterator(start, end []byte) (Iterator, error) {
	return h.c.ReverseIterator(start, end)
}

// writer is where writes and batches go: the collector while holding, the
// database otherwise.
func (h *HoldingDB) writer() DB {
	if h.Holding() {
		return h.c
	}
	return h.c.real
}
