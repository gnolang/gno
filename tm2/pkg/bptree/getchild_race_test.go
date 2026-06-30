package bptree

import (
	"sync"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestGetChild_ConcurrentReadWrite_NoRace runs a reader traversing a committed
// snapshot concurrently with a writer mutating + saving new versions, sharing
// the same nodeDB (and thus cached node objects). The reader uses Has — node
// traversal only, no value resolution — so it exercises getChild without
// touching the pendingVals map (an orthogonal, accepted single-writer race).
//
// Before removing the getChild write-back, this fails under `go test -race`:
// the reader memoizes a loaded child on a shared cached node while the writer
// reads that same node via Clone (`c := *n`). With the write-back gone, reads
// never mutate shared nodes, so it is clean.
func TestGetChild_ConcurrentReadWrite_NoRace(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 1000, NewNopLogger())
	const n = 5_000
	for i := 0; i < n; i++ {
		if _, err := tree.Set(i2b(i), i2b(i)); err != nil {
			t.Fatal(err)
		}
	}
	_, version, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}
	// Take the snapshot before starting concurrency so the reader only touches
	// the snapshot + shared cache, never the mutable tree's fields.
	imm, err := tree.GetImmutable(version)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for round := 0; round < 50; round++ {
			for i := 0; i < n; i++ {
				if _, err := imm.Has(i2b(i)); err != nil {
					t.Error(err)
					return
				}
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 3_000; i++ {
			if _, err := tree.Set(i2b(n+i), i2b(i)); err != nil {
				t.Error(err)
				return
			}
		}
		if _, _, err := tree.SaveVersion(); err != nil {
			t.Error(err)
		}
	}()
	wg.Wait()
}
