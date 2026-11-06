//go:build !race

// The race detector drops sync.Pool items at random and pebble pools its
// iterators, so allocation counts vary under it.

package pebbledb

import (
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/db"
)

func TestPebbleIteratorAllocatesOnlyItsWrapper(t *testing.T) {
	// Not parallel: allocations are counted across the process.
	pdb, r := newPrunedPebbleDB(t, 0)
	s, err := pdb.NewSnapshot()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, s.Close()) })
	snap := s.(*pebbleSnapshot)

	// Each bare iterator calls NewIter on the concrete reader, so its options
	// stay on the stack and it counts pebble's own allocations only.
	dbIter := func() (*pebble.Iterator, error) {
		return pdb.db.NewIter(&pebble.IterOptions{LowerBound: r.Start, UpperBound: r.End})
	}
	snapIter := func() (*pebble.Iterator, error) {
		return snap.snapshot.NewIter(&pebble.IterOptions{LowerBound: r.Start, UpperBound: r.End})
	}
	cases := map[string]struct {
		open func(start, end []byte) (db.Iterator, error)
		bare func() (*pebble.Iterator, error)
		seek func(*pebble.Iterator) bool
	}{
		"db/forward":       {pdb.Iterator, dbIter, (*pebble.Iterator).First},
		"db/reverse":       {pdb.ReverseIterator, dbIter, (*pebble.Iterator).Last},
		"snapshot/forward": {snap.Iterator, snapIter, (*pebble.Iterator).First},
		"snapshot/reverse": {snap.ReverseIterator, snapIter, (*pebble.Iterator).Last},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			bare := testing.AllocsPerRun(100, func() {
				it, err := tc.bare()
				require.NoError(t, err)
				tc.seek(it)
				require.NoError(t, it.Close())
			})
			wrapped := testing.AllocsPerRun(100, func() {
				itr, err := tc.open(r.Start, r.End)
				require.NoError(t, err)
				require.NoError(t, itr.Close())
			})
			require.LessOrEqual(t, wrapped, bare+1, "an iterator allocates more than pebble's iterator and its wrapper")
		})
	}
}
