package bptree

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestContract_ConcurrentSnapshotReadsVsWriter_NoRace encodes the MutableTree
// single-goroutine concurrency contract: the SANCTIONED concurrent pattern —
// readers using ONLY the safe entry points (GetImmutable(committed) and the
// returned ImmutableTree, GetVersioned, VersionExists, AvailableVersions) while
// a writer runs Set+SaveVersion — is race-clean. It consolidates the contract
// in one named place; the same surface is also exercised by the
// getChild/pendingVals/H2/M13 race tests, and prune-vs-reader by
// TestH2_ConcurrentPruneVsReader_NoRace.
//
// Readers must NOT call MutableTree's working-tree methods (Get/Has/Hash/
// Version/proof) directly — those race a writer by design and are out of
// contract (verified to report a race under -race in review).
func TestContract_ConcurrentSnapshotReadsVsWriter_NoRace(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 1000, NewNopLogger())
	for i := 0; i < 200; i++ {
		if _, err := tree.Set(i2b(i), i2b(i)); err != nil {
			t.Fatal(err)
		}
	}
	_, v0, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}
	// Pin v0 so the writer can never prune the readers' target out from under
	// them; GetImmutable(v0) then always succeeds. The registered pin blocks a
	// prune of v0 (prune-vs-reader itself is covered by the H2 test).
	pin, err := tree.GetImmutable(v0)
	if err != nil {
		t.Fatal(err)
	}
	defer pin.Close()

	var committed int64
	var writerWg, readerWg sync.WaitGroup
	stop := make(chan struct{})

	// Writer: Set + SaveVersion, bounded so the version count (and thus
	// AvailableVersions cost) stays modest; idles until readers finish.
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()
		rng := rand.New(rand.NewSource(1))
		for c := 0; c < 800; c++ {
			select {
			case <-stop:
				return
			default:
			}
			if _, err := tree.Set(i2b(rng.Intn(200)), i2b(rng.Intn(1<<20))); err != nil {
				t.Error(err)
				return
			}
			if _, _, err := tree.SaveVersion(); err != nil {
				t.Error(err)
				return
			}
			atomic.AddInt64(&committed, 1)
		}
		<-stop
	}()

	const readers = 4
	for r := 0; r < readers; r++ {
		readerWg.Add(1)
		go func() {
			defer readerWg.Done()
			for i := 0; i < 1000; i++ {
				imm, err := tree.GetImmutable(v0) // safe concurrent entry point
				if err != nil {
					t.Error(err) // v0 is pinned — must always succeed
					return
				}
				k := i2b(i % 200)
				if _, err := imm.Has(k); err != nil {
					t.Error(err)
				}
				if _, err := imm.Get(k); err != nil {
					t.Error(err)
				}
				if p, err := imm.GetMembershipProof(k); err != nil || p == nil {
					t.Errorf("membership proof: %v", err)
				}
				if it, err := imm.Iterator(nil, nil, true); err != nil {
					t.Error(err)
				} else {
					if it.Valid() {
						_ = it.Key()
					}
					it.Close()
				}
				imm.Close()

				if _, err := tree.GetVersioned(k, v0); err != nil {
					t.Error(err)
				}
				tree.VersionExists(v0)
				if i%200 == 0 {
					tree.AvailableVersions()
				}
			}
		}()
	}
	readerWg.Wait()
	close(stop)
	writerWg.Wait()

	if c := atomic.LoadInt64(&committed); c < 50 {
		t.Fatalf("writer made too little progress (%d commits) — not exercising concurrency", c)
	}
}
