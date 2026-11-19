package bptree

import (
	"encoding/binary"
	"sync"
	"testing"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

func k2b(i int) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(i))
	return b
}

// TestStore_PendingVals_ConcurrentQueryProve_NoRace guards the cross-package
// committed-snapshot value-resolution paths through the store wrapper:
//   - st.GetImmutable(v).Get(...)  → resolver → GetCommittedValueByKey → getCommittedValue
//   - st.Query(Prove:true)         → resolver → GetCommittedValueByKey → getCommittedValue
//
// concurrently with a writer issuing st.Set (→ SaveValue writes pendingVals).
// Since 75c946820 the committed paths are DB-only and never touch the map;
// regressing them back to GetValue must make this fail under -race.
func TestStore_PendingVals_ConcurrentQueryProve_NoRace(t *testing.T) {
	db := memdb.NewMemDB()
	st := StoreConstructor(db, types.StoreOptions{}).(*Store)

	const n = 1_500
	for i := range n {
		st.Set(nil, k2b(i), k2b(i))
	}
	cid := st.Commit() // version 1 is now committed
	version := cid.Version

	imm, err := st.GetImmutable(version)
	if err != nil {
		t.Fatal(err)
	}

	var writerWg, readerWg sync.WaitGroup
	stop := make(chan struct{})

	// Writer: stage new keys (pendingVals churns) without committing.
	writerWg.Go(func() {
		k := n
		for {
			select {
			case <-stop:
				return
			default:
			}
			st.Set(nil, k2b(k), k2b(k))
			k++
		}
	})

	// Reader A: committed-snapshot Get via the immutable store.
	readerWg.Go(func() {
		for range 200 {
			for i := range n {
				_ = imm.Get(nil, k2b(i))
			}
		}
	})

	// Reader B: committed Query with proof (membership) on the writer store.
	readerWg.Go(func() {
		for range 300 {
			for i := 0; i < n; i += 25 {
				res := st.Query(abci.RequestQuery{
					Path:   "/key",
					Data:   k2b(i),
					Height: version,
					Prove:  true,
				})
				if res.Error != nil {
					t.Errorf("query error: %v", res.Error)
					return
				}
			}
		}
	})

	readerWg.Wait()
	close(stop)
	writerWg.Wait()
}
