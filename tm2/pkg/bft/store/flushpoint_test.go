package store

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

func saveTestBlocks(t *testing.T, bs *BlockStore, heights ...int64) {
	t.Helper()

	for _, h := range heights {
		b := makeBlock(h, state, new(types.Commit))
		bs.SaveBlock(b, b.MakePartSet(2), makeTestCommit(h, tmtime.Now()))
	}
}

// TestLowerHeight pins that LowerHeight lowers the height on disk and in
// memory, and that the heights above can be saved again over the records it
// leaves.
func TestLowerHeight(t *testing.T) {
	t.Parallel()

	bs, db := freshBlockStore()
	saveTestBlocks(t, bs, 1, 2, 3)

	bs.LowerHeight(1)
	require.Equal(t, int64(1), bs.Height())
	require.Equal(t, int64(1), NewBlockStore(db).Height(), "the lower height is not on disk")
	saveTestBlocks(t, bs, 2, 3)
	require.Equal(t, int64(3), bs.Height())
	require.NotNil(t, bs.LoadBlock(3))

	require.Panics(t, func() { bs.LowerHeight(3) }, "LowerHeight to the same height")
	require.Panics(t, func() { bs.LowerHeight(4) }, "LowerHeight upward")
}

// TestFlushPoint pins the flush-point record's round trip, its absence on a
// store no flush point wrote, and that saving it never syncs: it rides in the
// batch of the block store's held writes.
func TestFlushPoint(t *testing.T) {
	t.Parallel()

	bs, db := freshBlockStore()
	_, ok := bs.FlushPoint()
	require.False(t, ok)

	FlushPoint{From: 4, To: 9}.Save(failingSyncDB{db}) // a SetSync would panic
	fp, ok := bs.FlushPoint()
	require.True(t, ok)
	require.Equal(t, FlushPoint{From: 4, To: 9}, fp)
}

type failingSyncDB struct{ dbm.DB }

func (failingSyncDB) SetSync([]byte, []byte) error { return errors.New("disk full") }

// TestBlockStoreStateSaveChecksError pins that a failed height record ends
// the save instead of passing for written.
func TestBlockStoreStateSaveChecksError(t *testing.T) {
	t.Parallel()

	require.Panics(t, func() { BlockStoreStateJSON{Height: 1}.Save(failingSyncDB{memdb.NewMemDB()}) })
}
