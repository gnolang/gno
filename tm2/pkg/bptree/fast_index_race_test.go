package bptree

import (
	"bytes"
	"fmt"
	"sync"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestFastIndex_RaceSnapshotReadersVsWriter covers fastGet concurrent with
// SaveVersion's batch commit under -race. Registered committed
// snapshots serve Get (via the fast index when trusted) while the writer
// overwrites and removes the SAME keyspace and commits version after version,
// so the snapshot's 'F' entries are concurrently re-stamped newer (the
// version-guard reject arm) and deleted (the miss arm). Readers assert
// snapshot-correct VALUES, not just no-error: a fastGet that wrongly trusted
// a too-new entry returns a later round's value and fails loudly. No pruning,
// so the snapshot's values stay resolvable throughout. Readers never touch
// working-tree fields (no tree.Version() etc.), per the MutableTree
// concurrency contract.
func TestFastIndex_RaceSnapshotReadersVsWriter(t *testing.T) {
	const (
		keyspace = 64
		readers  = 4
		rounds   = 40
	)
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 256, NewNopLogger(), FastIndexOption(true))
	key := func(i int) []byte { return fmt.Appendf(nil, "k%03d", i) }
	val := func(round, i int) []byte { return fmt.Appendf(nil, "r%d-%d", round, i) }

	for i := range keyspace {
		mustSet(t, tr, key(i), val(1, i))
	}
	snapVer := mustSave(t, tr)
	imm, err := tr.GetImmutable(snapVer)
	if err != nil {
		t.Fatalf("GetImmutable: %v", err)
	}
	defer imm.Close()

	stop := make(chan struct{})
	var wg sync.WaitGroup
	for range readers {
		wg.Go(func() {
			for {
				select {
				case <-stop:
					return
				default:
				}
				for i := range keyspace {
					got, err := imm.Get(key(i))
					if err != nil {
						t.Errorf("imm.Get(%q): %v", key(i), err)
						return
					}
					if want := val(1, i); !bytes.Equal(got, want) {
						t.Errorf("imm.Get(%q) = %q; want %q (too-new entry trusted?)", key(i), got, want)
						return
					}
				}
				if got, _ := imm.Get([]byte("absent")); got != nil {
					t.Errorf("imm.Get(absent) = %q; want nil", got)
					return
				}
			}
		})
	}

	for round := 2; round <= rounds; round++ {
		for i := range keyspace {
			if round%3 == 0 && i%2 == 0 {
				if _, _, err := tr.Remove(key(i)); err != nil {
					t.Fatalf("Remove: %v", err)
				}
			} else {
				mustSet(t, tr, key(i), val(round, i))
			}
		}
		mustSave(t, tr)
	}
	close(stop)
	wg.Wait()
}
