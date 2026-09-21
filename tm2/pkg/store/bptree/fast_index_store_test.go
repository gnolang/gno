package bptree

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bp "github.com/gnolang/gno/tm2/pkg/bptree"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

func countStoreFastEntries(t *testing.T, db dbm.DB) int {
	t.Helper()
	itr, err := db.Iterator([]byte{bp.PrefixFast}, []byte{bp.PrefixFast + 1})
	if err != nil {
		t.Fatalf("iterator: %v", err)
	}
	defer itr.Close()
	n := 0
	for ; itr.Valid(); itr.Next() {
		n++
	}
	return n
}

// TestStore_FastIndexParity: a store with the fast index ON must produce the
// same app hash, query values, and ICS23 proofs as one with it OFF — across the
// default query height (latest-1), the latest height (fast-path hit), and an
// older height (version-check fallback), for present and absent keys.
func TestStore_FastIndexParity(t *testing.T) {
	build := func(fast bool) (*Store, dbm.DB) {
		db := memdb.NewMemDB()
		ctor := StoreConstructor
		if fast {
			ctor = FastStoreConstructor
		}
		return ctor(db, types.StoreOptions{}).(*Store), db
	}
	stOn, dbOn := build(true)
	stOff, _ := build(false)

	keys := [][]byte{[]byte("alpha"), []byte("beta"), []byte("gamma"), []byte("delta")}
	var cid types.CommitID
	for round := range 3 {
		for i, k := range keys {
			v := fmt.Appendf(nil, "v%d.%d", round, i)
			stOn.Set(nil, k, v)
			stOff.Set(nil, k, v)
		}
		cidOn := stOn.Commit()
		cidOff := stOff.Commit()
		if !bytes.Equal(cidOn.Hash, cidOff.Hash) {
			t.Fatalf("app hash differs with fast index on/off at v%d", cidOn.Version)
		}
		cid = cidOn
	}

	if n := countStoreFastEntries(t, dbOn); n == 0 {
		t.Fatal("fast-index store wrote no 'F' entries")
	}

	query := func(st *Store, k []byte, h int64) abci.ResponseQuery {
		return st.Query(abci.RequestQuery{Path: "/key", Data: k, Height: h, Prove: true})
	}

	// h=0 → default (latest-1, fallback regime); latest → fast-path hit;
	// latest-1 explicit → version-check fallback.
	heights := []int64{0, cid.Version, cid.Version - 1}
	probe := append(append([][]byte{}, keys...), []byte("absent-key"))
	for _, h := range heights {
		for _, k := range probe {
			qOn := query(stOn, k, h)
			qOff := query(stOff, k, h)
			if !bytes.Equal(qOn.Value, qOff.Value) {
				t.Fatalf("h=%d key=%q value: on=%q off=%q", h, k, qOn.Value, qOff.Value)
			}
			if qOn.Log != qOff.Log {
				t.Fatalf("h=%d key=%q log: on=%q off=%q", h, k, qOn.Log, qOff.Log)
			}
			if !reflect.DeepEqual(qOn.Proof, qOff.Proof) {
				t.Fatalf("h=%d key=%q: ICS23 proof differs with index on/off", h, k)
			}
		}
	}
}
