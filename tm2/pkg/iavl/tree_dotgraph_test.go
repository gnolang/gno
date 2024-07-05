package iavl

import (
	"io"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

func TestWriteDOTGraph(t *testing.T) {
	t.Parallel()

	tree := NewMutableTree(memdb.NewMemDB(), 0)
	for _, ikey := range []byte{
		0x0a, 0x11, 0x2e, 0x32, 0x50, 0x72, 0x99, 0xa1, 0xe4, 0xf7,
	} {
		key := []byte{ikey}
		tree.Set(key, key)
	}
	WriteDOTGraph(io.Discard, tree.ImmutableTree, []PathToLeaf{})
}
