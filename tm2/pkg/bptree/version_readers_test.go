package bptree

import (
	"errors"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestGetImmutable_BlocksPruneUntilClose verifies that GetImmutable registers
// a version reader, so a concurrent PruneVersionsTo(v) is rejected until
// Close() is called. See Finding #30.
func TestGetImmutable_BlocksPruneUntilClose(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	// Build three versions so we can prune V1.
	tree.Set([]byte("a"), []byte("1"))
	tree.SaveVersion()
	tree.Set([]byte("b"), []byte("2"))
	tree.SaveVersion()
	tree.Set([]byte("c"), []byte("3"))
	tree.SaveVersion()

	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatalf("GetImmutable(1): %v", err)
	}

	// Prune must fail while the snapshot is active.
	if err := tree.PruneVersionsTo(1); !errors.Is(err, ErrActiveReaders) {
		t.Fatalf("PruneVersionsTo(1) with open snapshot: got %v, want %v", err, ErrActiveReaders)
	}

	// After Close(), prune succeeds.
	imm.Close()
	if err := tree.PruneVersionsTo(1); err != nil {
		t.Fatalf("PruneVersionsTo(1) after Close: %v", err)
	}
}

// TestImmutableTree_CloseIsIdempotent verifies that calling Close() multiple
// times does not double-decrement the version reader count.
func TestImmutableTree_CloseIsIdempotent(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	tree.Set([]byte("k"), []byte("v"))
	tree.SaveVersion()
	tree.Set([]byte("k2"), []byte("v2"))
	tree.SaveVersion()

	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatalf("GetImmutable(1): %v", err)
	}
	imm.Close()
	imm.Close() // must not underflow / double-decrement
	imm.Close()

	// Still pruneable, and no panic.
	if err := tree.PruneVersionsTo(1); err != nil {
		t.Fatalf("PruneVersionsTo(1) after multi-Close: %v", err)
	}
}

// TestGetImmutable_EmptyVersionRegistersReader verifies that even an empty
// saved version (root == nil) registers a reader so callers can hold open
// a consistent view.
func TestGetImmutable_EmptyVersionRegistersReader(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	// V1: empty. V2: non-empty.
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("V1: %v", err)
	}
	tree.Set([]byte("x"), []byte("y"))
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("V2: %v", err)
	}

	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatalf("GetImmutable(1): %v", err)
	}
	if !imm.IsEmpty() {
		t.Fatalf("V1 should be empty")
	}
	if err := tree.PruneVersionsTo(1); !errors.Is(err, ErrActiveReaders) {
		t.Fatalf("prune empty snapshot: got %v, want %v", err, ErrActiveReaders)
	}
	imm.Close()
	if err := tree.PruneVersionsTo(1); err != nil {
		t.Fatalf("prune after Close: %v", err)
	}
}

// TestIterator_BlocksPruneUntilClose verifies that an open iterator
// registers its own version reader, so a concurrent PruneVersionsTo(v)
// is rejected until the iterator is Closed. See Finding #1.
func TestIterator_BlocksPruneUntilClose(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	for i := 0; i < 50; i++ {
		tree.Set([]byte{byte(i)}, []byte{byte(i + 100)})
	}
	tree.SaveVersion()
	tree.Set([]byte("extra"), []byte("e"))
	tree.SaveVersion()

	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatalf("GetImmutable(1): %v", err)
	}

	// Build iterator via NewIteratorWithNDB; this must increment readers[1]
	// on top of the reader already held by imm.
	itr := NewIteratorWithNDB(imm, nil, nil, true, tree)

	// Close imm first — iterator must still hold its own reader.
	imm.Close()

	// Prune must still fail with only the iterator holding a reader.
	if err := tree.PruneVersionsTo(1); !errors.Is(err, ErrActiveReaders) {
		t.Fatalf("PruneVersionsTo(1) with open iterator: got %v, want %v", err, ErrActiveReaders)
	}

	// Close iterator — prune now succeeds.
	itr.Close()
	if err := tree.PruneVersionsTo(1); err != nil {
		t.Fatalf("PruneVersionsTo(1) after iterator Close: %v", err)
	}
}

// TestIterator_CloseIsIdempotent verifies that closing an iterator multiple
// times does not double-decrement the version reader count.
func TestIterator_CloseIsIdempotent(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	tree.Set([]byte("a"), []byte("1"))
	tree.SaveVersion()
	tree.Set([]byte("b"), []byte("2"))
	tree.SaveVersion()

	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatalf("GetImmutable(1): %v", err)
	}
	defer imm.Close()

	itr := NewIteratorWithNDB(imm, nil, nil, true, tree)
	itr.Close()
	itr.Close() // must not underflow
	itr.Close()

	// imm still holds its reader — prune should still fail.
	if err := tree.PruneVersionsTo(1); !errors.Is(err, ErrActiveReaders) {
		t.Fatalf("prune with imm open: got %v, want %v", err, ErrActiveReaders)
	}
}

// TestIterator_NoReaderForInMemoryTree verifies that a MutableTree iterator
// (working tree, version == 0) does NOT register a reader. Such iterators
// are iterating pre-commit state; there is no saved version to guard.
func TestIterator_NoReaderForInMemoryTree(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	tree.Set([]byte("a"), []byte("1"))
	tree.SaveVersion()
	tree.Set([]byte("b"), []byte("2"))
	// No SaveVersion — working tree remains pre-commit.

	itr, err := tree.Iterator(nil, nil, true)
	if err != nil {
		t.Fatalf("Iterator: %v", err)
	}
	defer itr.Close()

	// With only a working-tree iterator open, prune of V1 must still be
	// blocked ONLY by the caller's own retention policy, not by this
	// iterator. But V1 is the latest saved version, so prune should
	// return "cannot prune latest version" (not ErrActiveReaders).
	err = tree.PruneVersionsTo(1)
	if errors.Is(err, ErrActiveReaders) {
		t.Fatalf("working-tree iterator should not register as version reader; got %v", err)
	}
}

// TestProof_BlocksPruneDuringGeneration verifies that GetMembershipProof
// and GetNonMembershipProof hold an active reader on t.version while they
// run, but release it afterward. We cannot easily interleave prune inside
// proof generation, but we can verify the reader count is zero afterward
// by pruning successfully.
func TestProof_ReleasesReaderAfterReturn(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	tree.Set([]byte("proof-key"), []byte("proof-val"))
	tree.SaveVersion()
	tree.Set([]byte("other-key"), []byte("other-val"))
	tree.SaveVersion()

	if _, err := tree.GetMembershipProof([]byte("proof-key")); err != nil {
		t.Fatalf("GetMembershipProof: %v", err)
	}
	if _, err := tree.GetNonMembershipProof([]byte("missing")); err != nil {
		t.Fatalf("GetNonMembershipProof: %v", err)
	}

	// Both proofs completed; their readers must have been released.
	if err := tree.PruneVersionsTo(1); err != nil {
		t.Fatalf("PruneVersionsTo(1) after proofs: %v", err)
	}
}

// TestGetImmutable_BlocksOnInFlightPrune is the Finding #40 regression
// guard. Before the fix, GetImmutable called GetRoot BEFORE
// incrVersionReaders, so a prune that had already passed beginPruning
// could delete the root record and node entries between the two calls.
// After the fix, incrVersionReaders is called FIRST, so GetImmutable
// blocks on pruneMu until any in-flight prune completes — at which
// point it either sees the version cleanly or returns ErrVersionDoesNotExist.
func TestGetImmutable_BlocksOnInFlightPrune(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	tree.Set([]byte("a"), []byte("1"))
	tree.SaveVersion() // v1
	tree.Set([]byte("b"), []byte("2"))
	tree.SaveVersion() // v2
	tree.Set([]byte("c"), []byte("3"))
	tree.SaveVersion() // v3

	// Simulate a prune that has already passed beginPruning but has
	// not yet released pruneMu.
	if err := tree.ndb.beginPruning(1, 1); err != nil {
		t.Fatalf("beginPruning: %v", err)
	}

	type result struct {
		imm *ImmutableTree
		err error
	}
	got := make(chan result, 1)
	go func() {
		imm, err := tree.GetImmutable(2)
		got <- result{imm, err}
	}()

	// With the Finding #40 fix in place, GetImmutable must block on
	// pruneMu. Before the fix, it would happily read GetRoot and begin
	// loading nodes while the prune was free to delete them.
	select {
	case r := <-got:
		tree.ndb.endPruning()
		if r.imm != nil {
			r.imm.Close()
		}
		t.Fatalf("GetImmutable did not block while prune in progress (err=%v)", r.err)
	case <-time.After(100 * time.Millisecond):
		// Still blocked, as expected.
	}

	// Release the prune; GetImmutable should complete promptly.
	tree.ndb.endPruning()

	select {
	case r := <-got:
		if r.err != nil {
			t.Fatalf("GetImmutable(2) after endPruning: %v", r.err)
		}
		r.imm.Close()
	case <-time.After(500 * time.Millisecond):
		t.Fatal("GetImmutable did not unblock after endPruning")
	}
}

// TestGetImmutable_ReleasesReaderOnMissingVersion verifies that when
// GetImmutable is asked for a non-existent version, it does not leak a
// version reader. Before the Finding #40 fix, incrVersionReaders ran
// only on the success path; after the fix, it runs first and MUST be
// decremented on every error path. A leak would leave versionReaders
// permanently nonzero and prevent future prunes of that version.
func TestGetImmutable_ReleasesReaderOnMissingVersion(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	tree.Set([]byte("a"), []byte("1"))
	tree.SaveVersion() // v1
	tree.Set([]byte("b"), []byte("2"))
	tree.SaveVersion() // v2

	// Version 99 does not exist.
	imm, err := tree.GetImmutable(99)
	if err == nil {
		imm.Close()
		t.Fatalf("GetImmutable(99) should have failed")
	}

	// versionReaders[99] must be zero. We verify indirectly by
	// pruning: a phantom reader would be flagged by beginPruning.
	tree.ndb.mtx.Lock()
	n := tree.ndb.versionReaders[99]
	tree.ndb.mtx.Unlock()
	if n != 0 {
		t.Fatalf("GetImmutable error path leaked a reader: versionReaders[99]=%d", n)
	}
}
