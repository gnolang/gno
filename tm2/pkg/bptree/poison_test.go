package bptree

import (
	"bytes"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// failingBatchDB wraps a DB so its batches fail selectively. Arming flags
// live on the DB wrapper, not the batch: nodedb recycles batches via
// DiscardBatch, so per-batch state would silently disarm.
type failingBatchDB struct {
	dbm.DB
	failSetPrefix    atomic.Int32 // -1 disarmed, else the prefix byte
	failDeletePrefix atomic.Int32
	failWrite        atomic.Bool
}

func newFailingBatchDB(inner dbm.DB) *failingBatchDB {
	f := &failingBatchDB{DB: inner}
	f.failSetPrefix.Store(-1)
	f.failDeletePrefix.Store(-1)
	return f
}

func (f *failingBatchDB) NewBatch() dbm.Batch { return &failingBatch{Batch: f.DB.NewBatch(), db: f} }
func (f *failingBatchDB) NewBatchWithSize(size int) dbm.Batch {
	return &failingBatch{Batch: f.DB.NewBatchWithSize(size), db: f}
}

type failingBatch struct {
	dbm.Batch
	db *failingBatchDB
}

func (b *failingBatch) Set(key, value []byte) error {
	if p := b.db.failSetPrefix.Load(); p >= 0 && len(key) > 0 && key[0] == byte(p) {
		return errors.New("simulated batch.Set failure")
	}
	return b.Batch.Set(key, value)
}

func (b *failingBatch) Delete(key []byte) error {
	if p := b.db.failDeletePrefix.Load(); p >= 0 && len(key) > 0 && key[0] == byte(p) {
		return errors.New("simulated batch.Delete failure")
	}
	return b.Batch.Delete(key)
}

func (b *failingBatch) Write() error {
	if b.db.failWrite.Load() {
		return errors.New("simulated batch.Write failure")
	}
	return b.Batch.Write()
}

// N49: a Set whose value staging fails leaves the tree referencing a value
// that was never staged. SaveVersion must REFUSE (not silently commit the
// dangling valueKey); Rollback recovers; the replay matches a never-failed
// control run.
func TestPoison_SetValueStagingFailure(t *testing.T) {
	inner := memdb.NewMemDB()
	fdb := newFailingBatchDB(inner)
	tree := NewMutableTreeWithDB(fdb, 100, NewNopLogger())
	if _, err := tree.Set([]byte("k1"), []byte("v1")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}

	fdb.failSetPrefix.Store(int32(PrefixVal))
	if _, err := tree.Set([]byte("k2"), []byte("v2")); err == nil {
		t.Fatal("Set should have failed on value staging")
	}
	fdb.failSetPrefix.Store(-1)

	// The session is poisoned: a save would commit k2 with a dangling value.
	if _, _, err := tree.SaveVersion(); !errors.Is(err, ErrSessionPoisoned) {
		t.Fatalf("SaveVersion on poisoned session: want ErrSessionPoisoned, got %v", err)
	}
	if _, err := tree.Set([]byte("k3"), []byte("v3")); !errors.Is(err, ErrSessionPoisoned) {
		t.Fatalf("Set on poisoned session: want ErrSessionPoisoned, got %v", err)
	}
	if _, _, err := tree.Remove([]byte("k1")); !errors.Is(err, ErrSessionPoisoned) {
		t.Fatalf("Remove on poisoned session: want ErrSessionPoisoned, got %v", err)
	}

	// Rollback recovers; the replay commits cleanly and matches a control.
	tree.Rollback()
	if _, err := tree.Set([]byte("k2"), []byte("v2")); err != nil {
		t.Fatal(err)
	}
	hash, _, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}

	control := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger())
	control.Set([]byte("k1"), []byte("v1"))
	control.SaveVersion()
	control.Set([]byte("k2"), []byte("v2"))
	controlHash, _, err := control.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(hash, controlHash) {
		t.Fatalf("recovered replay hash %x != control %x", hash, controlHash)
	}

	// Cold reload: everything readable.
	fresh := NewMutableTreeWithDB(inner, 100, NewNopLogger())
	if _, err := fresh.Load(); err != nil {
		t.Fatal(err)
	}
	for k, want := range map[string]string{"k1": "v1", "k2": "v2"} {
		if got, err := fresh.Get([]byte(k)); err != nil || string(got) != want {
			t.Fatalf("Get(%s) = %q, %v", k, got, err)
		}
	}
}

// N8: a SaveVersion that fails at Commit must poison — a bare retry
// previously SUCCEEDED while committing an unloadable version (saveNode
// skipped the "already saved" nodes whose records the failed attempt
// discarded).
func TestPoison_SaveVersionCommitFailureRetryRefused(t *testing.T) {
	inner := memdb.NewMemDB()
	fdb := newFailingBatchDB(inner)
	tree := NewMutableTreeWithDB(fdb, 100, NewNopLogger())
	for i := range 50 {
		if _, err := tree.Set(fmt.Appendf(nil, "key%03d", i), []byte("v")); err != nil {
			t.Fatal(err)
		}
	}

	fdb.failWrite.Store(true)
	if _, _, err := tree.SaveVersion(); err == nil {
		t.Fatal("SaveVersion should have failed at Commit")
	}
	fdb.failWrite.Store(false)

	// Bare retry must be refused, not silently brick the version.
	if _, _, err := tree.SaveVersion(); !errors.Is(err, ErrSessionPoisoned) {
		t.Fatalf("retry after failed Commit: want ErrSessionPoisoned, got %v", err)
	}

	// Rollback + replay produces a loadable version with the control hash.
	tree.Rollback()
	for i := range 50 {
		if _, err := tree.Set(fmt.Appendf(nil, "key%03d", i), []byte("v")); err != nil {
			t.Fatal(err)
		}
	}
	hash, version, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}
	if version != 1 {
		t.Fatalf("version = %d, want 1", version)
	}

	control := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger())
	for i := range 50 {
		control.Set(fmt.Appendf(nil, "key%03d", i), []byte("v"))
	}
	controlHash, _, _ := control.SaveVersion()
	if !bytes.Equal(hash, controlHash) {
		t.Fatalf("replay hash %x != control %x", hash, controlHash)
	}

	fresh := NewMutableTreeWithDB(inner, 100, NewNopLogger())
	if _, err := fresh.Load(); err != nil {
		t.Fatalf("recovered version does not load: %v", err)
	}
	for i := range 50 {
		if v, err := fresh.Get(fmt.Appendf(nil, "key%03d", i)); err != nil || string(v) != "v" {
			t.Fatalf("key%03d = %q, %v", i, v, err)
		}
	}
}

// The B-1 review catch: even a PRE-save-phase SaveVersion failure (transient
// versionExistsE error) destroys the staged values via the deferred
// DiscardBatch, so a retry would commit dangling valueKeys. The uniform rule
// poisons that exit too.
func TestPoison_VersionExistsFailureRetryRefused(t *testing.T) {
	inner := memdb.NewMemDB()
	wrapped := &failingKeyHasDB{DB: inner}
	tree := NewMutableTreeWithDB(wrapped, 100, NewNopLogger())
	if _, err := tree.Set([]byte("k"), []byte("v")); err != nil {
		t.Fatal(err)
	}

	wrapped.failKey = rootDBKey(1)
	wrapped.armed = true
	if _, _, err := tree.SaveVersion(); err == nil {
		t.Fatal("SaveVersion should have failed on versionExistsE")
	}
	wrapped.armed = false

	// The transient fault cleared — but the staged value of k is GONE
	// (deferred DiscardBatch). A bare retry would commit k dangling.
	if _, _, err := tree.SaveVersion(); !errors.Is(err, ErrSessionPoisoned) {
		t.Fatalf("retry after transient existence failure: want ErrSessionPoisoned, got %v", err)
	}

	tree.Rollback()
	if _, err := tree.Set([]byte("k"), []byte("v")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	fresh := NewMutableTreeWithDB(inner, 100, NewNopLogger())
	if _, err := fresh.Load(); err != nil {
		t.Fatal(err)
	}
	if v, err := fresh.Get([]byte("k")); err != nil || string(v) != "v" {
		t.Fatalf("k = %q, %v (dangling valueKey?)", v, err)
	}
}

// N50: a Remove whose displaced-value delete fails would, if committed, leak
// the record permanently (it appears in no orphan list). Poison forces the
// Rollback that un-leaks it.
func TestPoison_RemoveOrphanFailure(t *testing.T) {
	inner := memdb.NewMemDB()
	fdb := newFailingBatchDB(inner)
	tree := NewMutableTreeWithDB(fdb, 100, NewNopLogger())
	if _, err := tree.Set([]byte("keep"), []byte("v")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}

	// Same-session Set + Remove → Tier-1 orphan → DeleteValueDirect.
	if _, err := tree.Set([]byte("temp"), []byte("v")); err != nil {
		t.Fatal(err)
	}
	fdb.failDeletePrefix.Store(int32(PrefixVal))
	if _, _, err := tree.Remove([]byte("temp")); err == nil {
		t.Fatal("Remove should have failed on the staged-value delete")
	}
	fdb.failDeletePrefix.Store(-1)

	if _, _, err := tree.SaveVersion(); !errors.Is(err, ErrSessionPoisoned) {
		t.Fatalf("SaveVersion after failed Remove: want ErrSessionPoisoned, got %v", err)
	}
	tree.Rollback()
	if _, _, err := tree.SaveVersion(); err != nil { // clean empty diff → idempotent adopt
		t.Fatal(err)
	}

	// No leaked v2-namespace value records.
	itr, err := inner.Iterator([]byte{PrefixVal}, []byte{PrefixVal + 1})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for ; itr.Valid(); itr.Next() {
		count++
	}
	itr.Close()
	if count != 1 { // just "keep"'s record
		t.Fatalf("leaked value records: have %d, want 1", count)
	}
}

// N7 (reorder): a failed LoadVersion leaves the OLD session fully intact —
// the staged key still reads, and committing it persists exactly that
// session (previously the session was wiped first, leaving the root
// referencing destroyed values, and the commit SUCCEEDED with dangling keys).
func TestLoadVersion_ErrorKeepsSessionIntact(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())
	if _, err := tree.Set([]byte("staged"), []byte("v")); err != nil {
		t.Fatal(err)
	}

	if _, err := tree.LoadVersion(999); err == nil {
		t.Fatal("LoadVersion(999) should fail")
	}

	// Old session intact: staged key readable, and the save commits it.
	if v, err := tree.Get([]byte("staged")); err != nil || string(v) != "v" {
		t.Fatalf("staged key lost after failed LoadVersion: %q, %v", v, err)
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	fresh := NewMutableTreeWithDB(db, 100, NewNopLogger())
	if _, err := fresh.Load(); err != nil {
		t.Fatal(err)
	}
	if v, err := fresh.Get([]byte("staged")); err != nil || string(v) != "v" {
		t.Fatalf("committed value dangling after failed-LoadVersion session: %q, %v", v, err)
	}
}

// Load() on an empty DB is a no-op and must NOT clear poison (clearing there
// would launder a poisoned never-saved session into a silent dangling
// commit).
func TestPoison_EmptyLoadDoesNotClear(t *testing.T) {
	fdb := newFailingBatchDB(memdb.NewMemDB())
	tree := NewMutableTreeWithDB(fdb, 100, NewNopLogger())

	fdb.failSetPrefix.Store(int32(PrefixVal))
	if _, err := tree.Set([]byte("k"), []byte("v")); err == nil {
		t.Fatal("Set should have failed")
	}
	fdb.failSetPrefix.Store(-1)

	if _, err := tree.Load(); err != nil { // empty DB: no-op
		t.Fatal(err)
	}
	if _, _, err := tree.SaveVersion(); !errors.Is(err, ErrSessionPoisoned) {
		t.Fatalf("empty Load() must not launder poison: got %v", err)
	}
}
