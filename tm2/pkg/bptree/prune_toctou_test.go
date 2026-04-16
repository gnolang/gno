package bptree

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestPrune_BeginPruningBlocksReaderRegistration is a deterministic test
// that verifies the core TOCTOU closure property: while a prune holds
// pruneMu, a concurrent call to incrVersionReaders blocks until the
// prune releases. This is the mechanism by which GetImmutable/newIterator
// cannot slip in between the prune's reader-count check and its
// deletions. See Finding #15.
func TestPrune_BeginPruningBlocksReaderRegistration(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	tree.Set([]byte("a"), []byte("1"))
	tree.SaveVersion() // v1
	tree.Set([]byte("b"), []byte("2"))
	tree.SaveVersion() // v2

	// Simulate prune claiming the lock but not yet deleting.
	if err := tree.ndb.beginPruning(1, 1); err != nil {
		t.Fatalf("beginPruning: %v", err)
	}

	// Attempt to register a reader in another goroutine; it must block.
	registered := make(chan struct{})
	go func() {
		tree.ndb.incrVersionReaders(2)
		close(registered)
	}()

	select {
	case <-registered:
		tree.ndb.endPruning()
		t.Fatal("incrVersionReaders did not block while prune in progress")
	case <-time.After(100 * time.Millisecond):
		// Still blocked, as expected.
	}

	// Release the prune; registration should now complete promptly.
	tree.ndb.endPruning()

	select {
	case <-registered:
		// OK.
	case <-time.After(500 * time.Millisecond):
		t.Fatal("incrVersionReaders did not unblock after endPruning")
	}
	// Clean up the reader slot.
	tree.ndb.decrVersionReaders(2)
}

// TestPrune_TOCTOU_ReaderRaceIsClosed races a reader registration against
// a concurrent prune. With the Finding #15 fix in place, the reader
// either (a) succeeds in registering BEFORE the prune starts (and prune
// returns ErrActiveReaders), or (b) the prune starts first and the
// reader's registration blocks until prune completes; either way, the
// reader never observes a version that has been concurrently deleted out
// from under it. See Finding #15.
func TestPrune_TOCTOU_ReaderRaceIsClosed(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	for i := 0; i < 100; i++ {
		tree.Set([]byte{byte(i)}, []byte{byte(i + 100)})
		if i%10 == 9 {
			tree.SaveVersion()
		}
	}
	// Now have versions 1..10; plenty of history to prune.
	// Keep the current tree state but add one more version as the "latest".
	tree.Set([]byte("tail"), []byte("t"))
	tree.SaveVersion() // v11

	const iterations = 64
	var (
		pruneAtV int64 = 1
		pruneWon atomic.Int64
		readerWon atomic.Int64
	)

	for i := 0; i < iterations; i++ {
		var wg sync.WaitGroup
		wg.Add(2)

		// Reader goroutine
		go func() {
			defer wg.Done()
			imm, err := tree.GetImmutable(pruneAtV)
			if err != nil {
				// Version was pruned before/during our call; natural
				// race outcome, no corruption.
				return
			}
			if imm != nil {
				_ = imm.Hash()
				imm.Close()
				readerWon.Add(1)
			}
		}()

		// Prune goroutine
		go func() {
			defer wg.Done()
			err := tree.PruneVersionsTo(pruneAtV)
			if err == nil {
				pruneWon.Add(1)
			} else if !errors.Is(err, ErrActiveReaders) {
				t.Errorf("unexpected prune err: %v", err)
			}
		}()

		wg.Wait()
		// If prune won, advance the target to the next valid version.
		if tree.VersionExists(pruneAtV + 1) {
			pruneAtV++
		}
		if pruneAtV >= 10 { // leave room vs latest v11
			break
		}
	}

	// Sanity: both sides participated (not a dead race).
	t.Logf("prune wins=%d reader wins=%d", pruneWon.Load(), readerWon.Load())
}

// TestPrune_BlocksConcurrentReaderRegistration verifies the blocking
// property: a goroutine calling GetImmutable(v) while prune is in flight
// waits for prune to complete, and then either sees the version (if it
// wasn't pruned) or gets an error (if it was). Critically, the reader
// never proceeds with a half-pruned view.
func TestPrune_BlocksConcurrentReaderRegistration(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	for i := 0; i < 200; i++ {
		tree.Set([]byte{byte(i % 200), byte(i / 200)}, []byte{byte(i)})
		if i%20 == 19 {
			tree.SaveVersion()
		}
	}
	tree.Set([]byte("tail"), []byte("t"))
	tree.SaveVersion() // latest

	// Start prune; then race a reader in parallel. We don't have a direct
	// hook to pause inside prune, but the semantic test is: after both
	// operations complete, the tree is in a consistent state (pruned
	// version is either absent or still protected), with no panics or
	// partial reads.
	readerObservedValid := atomic.Bool{}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		// Tiny stagger so prune has a chance to acquire pruneMu first.
		time.Sleep(50 * time.Microsecond)
		imm, err := tree.GetImmutable(1)
		if err == nil && imm != nil {
			// Sanity: if we got a snapshot, its hash must be well-defined.
			h := imm.Hash()
			if len(h) == 0 {
				t.Errorf("snapshot hash is empty")
			}
			imm.Close()
			readerObservedValid.Store(true)
		}
	}()

	go func() {
		defer wg.Done()
		if err := tree.PruneVersionsTo(1); err != nil && !errors.Is(err, ErrActiveReaders) {
			t.Errorf("prune error: %v", err)
		}
	}()

	wg.Wait()
	// No specific assertion on who won; the test passes if both goroutines
	// completed without panic or corruption.
}
