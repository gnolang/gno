package bptree

import (
	"encoding/binary"
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

// A seek stops at the first key inside the root prefix, so any root-prefixed
// key that is not a 9-byte version record has to be skipped rather than
// decoded. Nothing else in the suite reaches that branch: the tree itself only
// ever writes 9-byte root keys, so a stray can only arrive from a foreign
// writer sharing the DB.
func TestDiscoverVersionsSkipsStrayRootPrefixedKeys(t *testing.T) {
	t.Parallel()

	rootKey := func(v uint64) []byte {
		k := make([]byte, 9)
		k[0] = PrefixRoot
		binary.BigEndian.PutUint64(k[1:], v)
		return k
	}

	for _, tt := range []struct {
		name   string
		strays [][]byte
	}{
		{"short key below the low edge", [][]byte{{PrefixRoot, 0x00}}},
		{"long key above the high edge", [][]byte{append(rootKey(^uint64(0)), 0xff)}},
		{"bare prefix", [][]byte{{PrefixRoot}}},
		{"strays at both edges", [][]byte{
			{PrefixRoot, 0x00},
			append(rootKey(^uint64(0)), 0xff),
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := memdb.NewMemDB()
			tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())
			for i := range 3 {
				_, err := tree.Set([]byte{byte('a' + i)}, []byte{byte(i)})
				require.NoError(t, err)
				_, _, err = tree.SaveVersion()
				require.NoError(t, err)
			}
			for _, s := range tt.strays {
				require.NoError(t, db.Set(s, []byte("stray")))
			}

			fresh := NewMutableTreeWithDB(db, 1000, NewNopLogger())
			_, err := fresh.Load()
			require.NoError(t, err)

			require.Equal(t, int64(1), fresh.ndb.getFirstVersion(),
				"a stray must not be decoded as the first version")
			require.Equal(t, int64(3), fresh.ndb.getLatestVersion(),
				"a stray must not be decoded as the latest version")
		})
	}
}
