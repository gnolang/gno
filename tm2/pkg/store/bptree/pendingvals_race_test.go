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

// TestStore_PendingVals_ConcurrentQueryProve_NoRace drives the cross-package
// committed-snapshot value-resolution paths through the store wrapper:
//   - st.GetImmutable(v).Get(...)  → store.go:81 resolver → GetValueByKey → GetValue
//   - st.Query(Prove:true)         → store.go:318 resolver → GetValueByKey → GetValue
//
// concurrently with a writer issuing st.Set (→ SaveValue writes pendingVals).
//
// On HEAD this fails under -race; after the plan's fix it must be clean.
func TestStore_PendingVals_ConcurrentQueryProve_NoRace(t *testing.T) {
	db := memdb.NewMemDB()
	st := StoreConstructor(db, types.StoreOptions{}).(*Store)

	const n = 1_500
	for i := 0; i < n; i++ {
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
	writerWg.Add(1)
	go func() {
		defer writerWg.Done()
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
	}()

	// Reader A: committed-snapshot Get via the immutable store.
	readerWg.Add(1)
	go func() {
		defer readerWg.Done()
		for round := 0; round < 200; round++ {
			for i := 0; i < n; i++ {
				_ = imm.Get(nil, k2b(i))
			}
		}
	}()

	// Reader B: committed Query with proof (membership) on the writer store.
	readerWg.Add(1)
	go func() {
		defer readerWg.Done()
		for round := 0; round < 300; round++ {
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
	}()

	readerWg.Wait()
	close(stop)
	writerWg.Wait()
}
