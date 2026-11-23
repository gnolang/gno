package db_test

import (
	"bytes"
	"errors"
	"testing"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// syncLog wraps a database and records how its writes reach it.
type syncLog struct {
	dbm.DB
	syncs       int          // SetSync calls
	batchSyncs  int          // batches written with WriteSync
	sets        []string     // the Sets of its batches, as key=value
	onBatchSync func() error // runs inside a batch's WriteSync, before it lands
}

func (l *syncLog) SetSync(key, value []byte) error {
	l.syncs++
	return l.DB.SetSync(key, value)
}

func (l *syncLog) NewBatch() dbm.Batch { return &syncLogBatch{Batch: l.DB.NewBatch(), log: l} }

func (l *syncLog) NewBatchWithSize(size int) dbm.Batch {
	return &syncLogBatch{Batch: l.DB.NewBatchWithSize(size), log: l}
}

type syncLogBatch struct {
	dbm.Batch
	log *syncLog
}

func (b *syncLogBatch) Set(key, value []byte) error {
	b.log.sets = append(b.log.sets, string(key)+"="+string(value))
	return b.Batch.Set(key, value)
}

func (b *syncLogBatch) WriteSync() error {
	b.log.batchSyncs++
	if b.log.onBatchSync != nil {
		if err := b.log.onBatchSync(); err != nil {
			return err
		}
	}
	return b.Batch.WriteSync()
}

// TestHoldingDB_PassesThrough pins that a HoldingDB that is not holding is the
// database it wraps: writes land at once and a SetSync stays a SetSync.
func TestHoldingDB_PassesThrough(t *testing.T) {
	disk := &syncLog{DB: memdb.NewMemDB()}
	h := dbm.NewHoldingDB(disk)

	if err := h.Set([]byte("a"), []byte("1")); err != nil {
		t.Fatal(err)
	}
	if err := h.SetSync([]byte("b"), []byte("2")); err != nil {
		t.Fatal(err)
	}
	if v, _ := disk.Get([]byte("a")); !bytes.Equal(v, []byte("1")) {
		t.Fatalf("Set did not reach the database: %q", v)
	}
	if disk.syncs != 1 {
		t.Fatalf("SetSync reached the database %d times as a sync, want 1", disk.syncs)
	}
	it, err := h.Iterator(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	it.Close()
}

// TestHoldingDB_HoldsUntilWriteHeld pins the holding contract: writes stay out
// of the database, readable through the HoldingDB, until WriteHeld writes all
// of them in order as one synced batch; then the database has them and
// nothing is held.
func TestHoldingDB_HoldsUntilWriteHeld(t *testing.T) {
	disk := &syncLog{DB: memdb.NewMemDB()}
	disk.Set([]byte("gone"), []byte("old"))
	h := dbm.NewHoldingDB(disk)
	h.Hold(true)

	h.Set([]byte("a"), []byte("1"))
	h.SetSync([]byte("a"), []byte("2")) // the later write wins
	h.Set([]byte("empty"), []byte{})
	h.Delete([]byte("gone"))
	b := h.NewBatch()
	b.Set([]byte("batched"), []byte("3"))
	if err := b.WriteSync(); err != nil {
		t.Fatal(err)
	}

	if v, _ := disk.Get([]byte("a")); v != nil || disk.syncs != 0 || disk.batchSyncs != 0 {
		t.Fatalf("a held write reached the database: a=%q syncs=%d batches=%d", v, disk.syncs, disk.batchSyncs)
	}
	for key, want := range map[string]string{"a": "2", "empty": "", "batched": "3"} {
		v, _ := h.Get([]byte(key))
		if v == nil || string(v) != want {
			t.Fatalf("held %q reads %#v, want %q", key, v, want)
		}
	}
	if ok, _ := h.Has([]byte("gone")); ok {
		t.Fatal("held delete does not mask the database's value")
	}
	if h.HeldBytes() == 0 {
		t.Fatal("HeldBytes is zero with writes held")
	}

	// A record stays readable while its batch is being written.
	disk.onBatchSync = func() error {
		if v, _ := h.Get([]byte("a")); !bytes.Equal(v, []byte("2")) {
			return errors.New("held write unreadable during WriteHeld")
		}
		return nil
	}
	if err := h.WriteHeld(); err != nil {
		t.Fatal(err)
	}
	if disk.batchSyncs != 1 || disk.syncs != 0 {
		t.Fatalf("WriteHeld wrote %d synced batches and %d syncs, want one batch", disk.batchSyncs, disk.syncs)
	}
	if h.HeldBytes() != 0 {
		t.Fatalf("writes still held after WriteHeld: %d bytes", h.HeldBytes())
	}
	for key, want := range map[string]string{"a": "2", "empty": "", "batched": "3"} {
		if v, _ := disk.Get([]byte(key)); v == nil || string(v) != want {
			t.Fatalf("database has %q = %#v, want %q", key, v, want)
		}
	}
	if ok, _ := disk.Has([]byte("gone")); ok {
		t.Fatal("held delete not written")
	}
	h.Hold(false)
}

// TestHoldingDB_FailedWriteKeepsHeld pins that a failed WriteHeld drops
// nothing: the writes stay held and readable.
func TestHoldingDB_FailedWriteKeepsHeld(t *testing.T) {
	disk := &syncLog{DB: memdb.NewMemDB()}
	h := dbm.NewHoldingDB(disk)
	h.Hold(true)
	h.Set([]byte("a"), []byte("1"))

	disk.onBatchSync = func() error { return errors.New("disk full") }
	if err := h.WriteHeld(); err == nil {
		t.Fatal("WriteHeld reported success for a failed write")
	}
	if v, _ := h.Get([]byte("a")); !bytes.Equal(v, []byte("1")) {
		t.Fatalf("held write lost by a failed WriteHeld: %q", v)
	}
}

// TestHoldingDB_ReleaseWithHeldWritesPanics pins that holding cannot be turned
// off while writes are held, which would lose them or write them late.
func TestHoldingDB_ReleaseWithHeldWritesPanics(t *testing.T) {
	h := dbm.NewHoldingDB(memdb.NewMemDB())
	h.Hold(true)
	h.Set([]byte("a"), []byte("1"))
	defer func() {
		if recover() == nil {
			t.Fatal("Hold(false) with writes held did not panic")
		}
	}()
	h.Hold(false)
}

// TestHoldingDB_IteratorRefusedWhileHolding pins that an iterator panics while
// writes are held, rather than read the database without them.
func TestHoldingDB_IteratorRefusedWhileHolding(t *testing.T) {
	h := dbm.NewHoldingDB(memdb.NewMemDB())
	h.Hold(true)
	defer func() {
		err, _ := recover().(error)
		if !errors.Is(err, dbm.ErrIteratorOverHeldWrites) {
			t.Fatalf("iterator while holding: recovered %v, want ErrIteratorOverHeldWrites", err)
		}
	}()
	h.Iterator(nil, nil)
}
