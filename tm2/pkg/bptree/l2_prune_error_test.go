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

var errInjectedNodeRead = errors.New("injected node read failure")

// failingGetDB wraps a DB and, while armed, fails node-key Gets after
// `allow` successful ones — injecting a DB read error mid-prune.
type failingGetDB struct {
	dbm.DB
	armed int32
	allow int32
}

func (f *failingGetDB) Get(key []byte) ([]byte, error) {
	if atomic.LoadInt32(&f.armed) == 1 && len(key) > 0 && key[0] == PrefixNode {
		if atomic.AddInt32(&f.allow, -1) < 0 {
			return nil, errInjectedNodeRead
		}
	}
	return f.DB.Get(key)
}

// TestPrune_ErrorLeavesNoPartialDeletes (L2): a DB read error mid-prune must
// not leave partially-staged deletes in the shared batch — otherwise the next
// unrelated Commit flushes them, leaving the version's root ref alive with
// part of its subtree deleted: the version becomes unreadable AND unprunable
// (the retry fails on the missing nodes, and the store's pruning Commit panics
// on that error → a persistent crash loop).
func TestPrune_ErrorLeavesNoPartialDeletes(t *testing.T) {
	fdb := &failingGetDB{DB: memdb.NewMemDB()}
	tree := NewMutableTreeWithDB(fdb, 0, NewNopLogger()) // no cache: loads hit the DB

	// v1: enough keys for a multi-node tree; v2/v3: small changes.
	for i := 0; i < 200; i++ {
		tree.Set(fmt.Appendf(nil, "pe%04d", i), fmt.Appendf(nil, "v1_%04d", i))
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	tree.Set([]byte("pe0000"), []byte("v2"))
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	tree.Set([]byte("pe0001"), []byte("v3"))
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}

	// Snapshot v1's content pre-prune.
	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatal(err)
	}
	v1Hash := append([]byte(nil), imm.Hash()...)
	imm.Close()

	// Arm: let a couple of node loads through (the two roots), then fail —
	// inside pruneVersion(1, 2), after v1's root delete is already staged.
	atomic.StoreInt32(&fdb.allow, 2)
	atomic.StoreInt32(&fdb.armed, 1)
	err = tree.DeleteVersionsTo(2)
	atomic.StoreInt32(&fdb.armed, 0)
	if err == nil || !errors.Is(err, errInjectedNodeRead) {
		t.Fatalf("prune with injected failure: want wrapped errInjectedNodeRead, got %v", err)
	}

	// The unrelated commit that would flush poisoned staged deletes.
	tree.Set([]byte("after"), []byte("x"))
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}

	// v1 must be fully intact: same hash, every key readable.
	imm, err = tree.GetImmutable(1)
	if err != nil {
		t.Fatalf("v1 corrupted by failed prune + unrelated commit: %v", err)
	}
	if !bytes.Equal(imm.Hash(), v1Hash) {
		t.Fatalf("v1 hash changed after failed prune")
	}
	for i := 0; i < 200; i++ {
		v, err := imm.Get(fmt.Appendf(nil, "pe%04d", i))
		if err != nil || !bytes.Equal(v, fmt.Appendf(nil, "v1_%04d", i)) {
			t.Fatalf("v1 key %d unreadable after failed prune: %v", i, err)
		}
	}
	imm.Close()

	// Retry must succeed (no half-deleted version blocking it).
	if err := tree.DeleteVersionsTo(2); err != nil {
		t.Fatalf("retry prune: %v", err)
	}
	if tree.VersionExists(1) || tree.VersionExists(2) {
		t.Fatal("retry prune did not delete v1/v2")
	}
}

// TestPrune_RequiresCleanSession (L2 guard): pruning with uncommitted
// working-session state must be refused — the prune commits and (on error)
// discards the shared batch, which would otherwise flush or drop the session's
// staged writes.
func TestPrune_RequiresCleanSession(t *testing.T) {
	tree := newPruneTree(t)
	for i := 0; i < 30; i++ {
		tree.Set(fmt.Appendf(nil, "cs%03d", i), []byte("v"))
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	tree.Set([]byte("cs_extra"), []byte("v"))
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}

	// Structurally dirty session.
	tree.Set([]byte("dirty"), []byte("v"))
	if err := tree.DeleteVersionsTo(1); !errors.Is(err, ErrUncommittedChanges) {
		t.Fatalf("dirty prune: want ErrUncommittedChanges, got %v", err)
	}
	tree.Rollback()

	// Clean again: prune succeeds.
	if err := tree.DeleteVersionsTo(1); err != nil {
		t.Fatalf("clean prune: %v", err)
	}
}

// TestPrune_NetZeroSessionStillRejected (L2 guard, four-term): a session that
// nets back to root == lastSaved with empty pendingVals (Set then Remove after
// loading an EMPTY old version) still holds staged batch writes keyed into a
// committed version's value namespace; flushing them via prune would corrupt
// that version. The nextValueNonce guard term catches it.
func TestPrune_NetZeroSessionStillRejected(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	if _, _, err := tree.SaveVersion(); err != nil { // v1: empty
		t.Fatal(err)
	}
	tree.Set([]byte("b"), []byte("B2"))
	if _, _, err := tree.SaveVersion(); err != nil { // v2 = {b}, value key {2,0}
		t.Fatal(err)
	}
	tree.Set([]byte("c"), []byte("C3"))
	if _, _, err := tree.SaveVersion(); err != nil { // v3 (so v1/v2 are prunable)
		t.Fatal(err)
	}

	// Load empty v1; working version is 2. Set+Remove nets the root back to
	// nil == lastSaved and empties pendingVals, but the batch now stages a
	// Delete of value key {2,0} — v2's committed value for "b".
	if _, err := tree.LoadVersion(1); err != nil {
		t.Fatal(err)
	}
	tree.Set([]byte("x"), []byte("X"))
	if _, _, err := tree.Remove([]byte("x")); err != nil {
		t.Fatal(err)
	}

	if err := tree.PruneVersionsTo(1); !errors.Is(err, ErrUncommittedChanges) {
		t.Fatalf("net-zero dirty prune: want ErrUncommittedChanges, got %v", err)
	}
	tree.Rollback()
	if err := tree.PruneVersionsTo(1); err != nil {
		t.Fatalf("prune after rollback: %v", err)
	}

	// v2's committed value must be intact.
	got, err := tree.GetVersioned([]byte("b"), 2)
	if err != nil || string(got) != "B2" {
		t.Fatalf("v2 value corrupted: got %q, err %v", got, err)
	}
}
