package bptree

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// TestStore_KeyOverMaxFailsWhenStaged nests a tx cache in a deliver-state
// cache over the tree, as BaseApp does: a key over the cap fails at Set, Get
// and Delete, and a key at the cap commits.
func TestStore_KeyOverMaxFailsWhenStaged(t *testing.T) {
	st := FastStoreConstructor(memdb.NewMemDB(), types.StoreOptions{}).(*Store)
	deliver := st.CacheWrap()
	tx := deliver.CacheWrap()

	atCap := bytes.Repeat([]byte{'a'}, types.MaxKeyLen)
	tx.Set(nil, atCap, []byte("v"))

	over := bytes.Repeat([]byte{'b'}, types.MaxKeyLen+1)
	want := fmt.Sprintf("store key length %d exceeds the maximum %d", len(over), types.MaxKeyLen)
	for name, op := range map[string]func(){
		"Set":    func() { tx.Set(nil, over, []byte("v")) },
		"Get":    func() { tx.Get(nil, over) },
		"Delete": func() { tx.Delete(nil, over) },
	} {
		require.PanicsWithValue(t, want, op, name)
	}

	tx.Write()
	deliver.Write()
	st.Commit()
	if got := st.Get(nil, atCap); !bytes.Equal(got, []byte("v")) {
		t.Fatalf("key at types.MaxKeyLen: got %q after commit, want %q", got, "v")
	}
}
