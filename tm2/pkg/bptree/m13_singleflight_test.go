package bptree

import (
	"sync"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestM13_ConcurrentGetNode hammers GetNode for the same NodeKey from many
// goroutines with the cache disabled (cacheSize=0) so every call goes through
// the singleflight load path (M13). It must be race-clean and every call must
// return a correct node.
func TestM13_ConcurrentGetNode(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 0, NewNopLogger()) // cacheSize=0 → no cache; all loads hit singleflight
	for i := 0; i < 500; i++ {
		if _, err := tree.Set(i2b(i), i2b(i)); err != nil {
			t.Fatal(err)
		}
	}
	_, v, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}
	rootNK, _, err := tree.ndb.GetRoot(v)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := tree.ndb.GetNode(rootNK)
	if err != nil {
		t.Fatal(err)
	}
	refHash := ref.Hash()

	const goroutines, iters = 32, 300
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				n, err := tree.ndb.GetNode(rootNK)
				if err != nil {
					t.Error(err)
					return
				}
				if n == nil || n.Hash() != refHash {
					t.Error("GetNode returned a wrong or nil node under concurrency")
					return
				}
			}
		}()
	}
	wg.Wait()
}
