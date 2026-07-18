package bptree

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/stretchr/testify/require"
)

// discoverVersions seeks the first and last root keys instead of scanning every
// retained root. After pruning opens a gap at the low end, the discovered edges
// must still be the smallest and largest surviving versions, matching what a
// full scan would find.
func TestDiscoverVersionsSeeksEdgesAfterPruneGap(t *testing.T) {
	t.Parallel()

	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	for i := range 5 {
		_, err := tree.Set([]byte{byte('a' + i)}, []byte{byte(i)})
		require.NoError(t, err)
		_, _, err = tree.SaveVersion()
		require.NoError(t, err)
	}
	// Versions 1..5 exist; drop 1 and 2 so the surviving floor is 3.
	require.NoError(t, tree.PruneVersionsTo(2))

	// A fresh tree over the same DB rediscovers versions during Load.
	fresh := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	_, err := fresh.Load()
	require.NoError(t, err)

	require.Equal(t, int64(3), fresh.ndb.getFirstVersion(),
		"first must be the smallest surviving version")
	require.Equal(t, int64(5), fresh.ndb.getLatestVersion(),
		"latest must be the largest surviving version")
	require.Equal(t, []int{3, 4, 5}, fresh.AvailableVersions())
}
