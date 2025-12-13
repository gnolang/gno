package backup

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/stretchr/testify/require"
)

func TestInfo(t *testing.T) {
	dir := t.TempDir()

	testWriteBlocks(t, dir, 142, 1)

	state, err := readState(dir)
	require.NoError(t, err)
	require.Equal(t, int64(1), state.StartHeight, "expected valid initial height after first write")
	require.Equal(t, int64(142), state.EndHeight, "expected valid end height after first write")

	testWriteBlocks(t, dir, 142, 101)

	state, err = readState(dir)
	require.NoError(t, err)
	require.Equal(t, int64(1), state.StartHeight, "expected same initial height after second write")
	require.Equal(t, int64(242), state.EndHeight, "expected valid end height after second write")
}

func testWriteBlocks(t *testing.T, dir string, count int64, expectedStart int64) {
	t.Helper()
	require.NoError(t, WithWriter(dir, 0, 0, nil, func(startHeight int64, write Writer) error {
		require.Equal(t, expectedStart, startHeight, "expected correct start height in callback")
		for i := range count {
			data, err := amino.Marshal(&types.Block{Header: types.Header{Height: i + startHeight}})
			require.NoError(t, err)
			if err := write(data); err != nil {
				return err
			}
		}
		return nil
	}))
}
