package bptree

import (
	"errors"
	"math/rand"
	"sync"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// h2h3Tree builds a 2-version DB-backed tree for the version-reader tests.
func h2h3Tree(t *testing.T) *MutableTree {
	t.Helper()
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 1000, NewNopLogger())
	for i := range 50 {
		if _, err := tree.Set(i2b(i), i2b(i)); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := tree.SaveVersion(); err != nil { // v1
		t.Fatal(err)
	}
	if _, err := tree.Set(i2b(1000), i2b(1000)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := tree.SaveVersion(); err != nil { // v2
		t.Fatal(err)
	}
	return tree
}

// TestH2_GetImmutableBlocksPrune: a registered Get snapshot blocks a prune of its
// version until it is Closed (H2).
func TestH2_GetImmutableBlocksPrune(t *testing.T) {
	tree := h2h3Tree(t)
	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatal(err)
	}
	if err := tree.DeleteVersionsTo(1); !errors.Is(err, ErrActiveReaders) {
		t.Fatalf("prune with open snapshot: want ErrActiveReaders, got %v", err)
	}
	if err := imm.Close(); err != nil {
		t.Fatal(err)
	}
	if err := tree.DeleteVersionsTo(1); err != nil {
		t.Fatalf("prune after Close: %v", err)
	}
}

// TestH2_GetVersionedNoLeak: GetVersioned Closes its internal snapshot, leaving no
// reader (else that version could never prune — a hard panic at the next Commit).
func TestH2_GetVersionedNoLeak(t *testing.T) {
	tree := h2h3Tree(t)
	if _, err := tree.GetVersioned(i2b(0), 1); err != nil {
		t.Fatal(err)
	}
	if n := tree.ndb.versionReaders[1]; n != 0 {
		t.Fatalf("GetVersioned leaked %d reader(s) on v1", n)
	}
	if err := tree.DeleteVersionsTo(1); err != nil {
		t.Fatalf("prune after GetVersioned: %v", err)
	}
}

// TestH2_ProofNoLeak: proof generation Closes its internal snapshot.
func TestH2_ProofNoLeak(t *testing.T) {
	tree := h2h3Tree(t)
	if _, err := tree.GetMembershipProof(i2b(0)); err != nil {
		t.Fatal(err)
	}
	if _, err := tree.GetNonMembershipProof(i2b(99999)); err != nil {
		t.Fatal(err)
	}
	if n := tree.ndb.versionReaders[tree.Version()]; n != 0 {
		t.Fatalf("proof generation leaked %d reader(s)", n)
	}
}

// TestH2_SnapshotCloseNoUnderflow: Snapshot is unregistered; Closing it must NOT
// decrement another live reader's count on the same version (the Q-G underflow).
func TestH2_SnapshotCloseNoUnderflow(t *testing.T) {
	tree := h2h3Tree(t)
	imm, err := tree.GetImmutable(1) // a real registered reader on v1
	if err != nil {
		t.Fatal(err)
	}
	snap := tree.Snapshot(1) // unregistered (committed=false)
	if err := snap.Close(); err != nil {
		t.Fatal(err)
	}
	// If snap.Close had decremented v1, this prune would wrongly succeed.
	if err := tree.DeleteVersionsTo(1); !errors.Is(err, ErrActiveReaders) {
		t.Fatalf("Snapshot.Close underflowed v1's reader count: prune returned %v", err)
	}
	if err := imm.Close(); err != nil {
		t.Fatal(err)
	}
	if err := tree.DeleteVersionsTo(1); err != nil {
		t.Fatalf("prune after real Close: %v", err)
	}
}

// TestH2_UnregisteredDoesNotBlockPrune: GetImmutableUnregistered holds no reader
// (the store's long-lived immutable load), so a prune of its version proceeds.
func TestH2_UnregisteredDoesNotBlockPrune(t *testing.T) {
	tree := h2h3Tree(t)
	imm, err := tree.GetImmutableUnregistered(1)
	if err != nil {
		t.Fatal(err)
	}
	if imm.registered {
		t.Fatal("GetImmutableUnregistered must not register a reader")
	}
	if err := tree.DeleteVersionsTo(1); err != nil {
		t.Fatalf("unregistered snapshot wrongly blocked prune: %v", err)
	}
	if err := imm.Close(); err != nil { // no-op; must not panic/underflow
		t.Fatal(err)
	}
}

// TestH2_ConcurrentPruneVsReader_NoRace: a reader repeatedly snapshots + reads a
// fixed version while a writer commits and prunes. Exercises the pruneMu /
// versionReaders machinery; must be clean under `go test -race` with no
// use-after-delete (a held snapshot blocks the prune of its version).
func TestH2_ConcurrentPruneVsReader_NoRace(t *testing.T) {
	tree := h2h3Tree(t)
	var wg sync.WaitGroup
	stop := make(chan struct{})

	wg.Go(func() { // writer: commit + prune old versions
		rng := rand.New(rand.NewSource(1))
		for {
			select {
			case <-stop:
				return
			default:
			}
			if _, err := tree.Set(i2b(rng.Intn(1000)), i2b(1)); err != nil {
				t.Error(err)
				return
			}
			_, v, err := tree.SaveVersion()
			if err != nil {
				t.Error(err)
				return
			}
			if v > 3 {
				// A held reader legitimately blocks its version's prune.
				if err := tree.DeleteVersionsTo(v - 2); err != nil && !errors.Is(err, ErrActiveReaders) {
					t.Error(err)
					return
				}
			}
		}
	})

	wg.Go(func() { // reader: snapshot a fixed version, read, close
		defer close(stop)
		for range 3000 {
			imm, err := tree.GetImmutable(2)
			if err != nil {
				continue // v2 may have already been pruned
			}
			_, _ = imm.Has(i2b(0))
			imm.Close()
		}
	})

	wg.Wait()
}
