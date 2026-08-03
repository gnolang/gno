package db

import "fmt"

// SnapshotDB wraps a Snapshot to implement the DB interface.
// All read operations delegate to the snapshot. Write operations panic.
// Close is a no-op — the caller owns the snapshot lifecycle.
type SnapshotDB struct {
	Snapshot
}

var _ DB = (*SnapshotDB)(nil)

// NewSnapshotDB returns a read-only DB backed by snap.
func NewSnapshotDB(snap Snapshot) *SnapshotDB {
	return &SnapshotDB{snap}
}

func (s *SnapshotDB) NewSnapshot() (Snapshot, error) { return s, nil }
func (s *SnapshotDB) Close() error                   { return nil }
func (s *SnapshotDB) Print() error                   { fmt.Print("(snapshot) "); return nil }
func (s *SnapshotDB) Stats() map[string]string       { return nil }

func (s *SnapshotDB) Set([]byte, []byte) error     { panic("SnapshotDB is read-only") }
func (s *SnapshotDB) SetSync([]byte, []byte) error { panic("SnapshotDB is read-only") }
func (s *SnapshotDB) Delete([]byte) error          { panic("SnapshotDB is read-only") }
func (s *SnapshotDB) DeleteSync([]byte) error      { panic("SnapshotDB is read-only") }

// NewBatch and NewBatchWithSize return a no-op batch. IAVL creates a
// BatchWithFlusher eagerly in its constructor even for immutable loads, but
// never commits it when skipFastStorageUpgrade=true. The no-op batch panics
// on Write/WriteSync to catch any unexpected write attempts.
func (s *SnapshotDB) NewBatch() Batch            { return &snapshotNoopBatch{} }
func (s *SnapshotDB) NewBatchWithSize(int) Batch { return &snapshotNoopBatch{} }

// snapshotNoopBatch silently discards Set/Delete but panics on Write/WriteSync.
type snapshotNoopBatch struct{}

var _ Batch = (*snapshotNoopBatch)(nil)

func (b *snapshotNoopBatch) Set(_, _ []byte) error     { return nil }
func (b *snapshotNoopBatch) Delete(_ []byte) error     { return nil }
func (b *snapshotNoopBatch) Close() error              { return nil }
func (b *snapshotNoopBatch) GetByteSize() (int, error) { return 0, nil }
func (b *snapshotNoopBatch) Write() error {
	panic("snapshotNoopBatch: unexpected Write on read-only DB")
}
func (b *snapshotNoopBatch) WriteSync() error {
	panic("snapshotNoopBatch: unexpected WriteSync on read-only DB")
}
