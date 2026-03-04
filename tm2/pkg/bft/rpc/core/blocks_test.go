package core

import (
	"fmt"
	"testing"
	"time"

	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockchainInfo(t *testing.T) {
	t.Parallel()

	cases := []struct {
		minVal, maxVal int64
		height         int64
		limit          int64
		resultLength   int64
		wantErr        bool
	}{
		// min > max
		{0, 0, 0, 10, 0, true},  // min set to 1
		{0, 1, 0, 10, 0, true},  // max set to height (0)
		{0, 0, 1, 10, 1, false}, // max set to height (1)
		{2, 0, 1, 10, 0, true},  // max set to height (1)
		{2, 1, 5, 10, 0, true},

		// negative
		{1, 10, 14, 10, 10, false}, // control
		{-1, 10, 14, 10, 0, true},
		{1, -10, 14, 10, 0, true},
		{-9223372036854775808, -9223372036854775788, 100, 20, 0, true},

		// check limit and height
		{1, 1, 1, 10, 1, false},
		{1, 1, 5, 10, 1, false},
		{2, 2, 5, 10, 1, false},
		{1, 2, 5, 10, 2, false},
		{1, 5, 1, 10, 1, false},
		{1, 5, 10, 10, 5, false},
		{1, 15, 10, 10, 10, false},
		{1, 15, 15, 10, 10, false},
		{1, 15, 15, 20, 15, false},
		{1, 20, 15, 20, 15, false},
		{1, 20, 20, 20, 20, false},
	}

	for i, c := range cases {
		caseString := fmt.Sprintf("test %d failed", i)
		minVal, maxVal, err := filterMinMax(c.height, c.minVal, c.maxVal, c.limit)
		if c.wantErr {
			require.Error(t, err, caseString)
		} else {
			require.NoError(t, err, caseString)
			require.Equal(t, 1+maxVal-minVal, c.resultLength, caseString)
		}
	}
}

func TestGetHeight(t *testing.T) {
	t.Parallel()

	cases := []struct {
		currentHeight int64
		heightPtr     *int64
		minVal        int64
		res           int64
		wantErr       bool
	}{
		// height >= min
		{42, int64Ptr(0), 0, 0, false},
		{42, int64Ptr(1), 0, 1, false},

		// height < min
		{42, int64Ptr(0), 1, 0, true},

		// nil height
		{42, nil, 1, 42, false},
	}

	for i, c := range cases {
		caseString := fmt.Sprintf("test %d failed", i)
		res, err := getHeightWithMin(c.currentHeight, c.heightPtr, c.minVal)
		if c.wantErr {
			require.Error(t, err, caseString)
		} else {
			require.NoError(t, err, caseString)
			require.Equal(t, res, c.res, caseString)
		}
	}
}

func int64Ptr(v int64) *int64 {
	return &v
}

func TestBlockchainInfoHandler(t *testing.T) {
	t.Run("nil block meta in range", func(t *testing.T) {
		var storeHeight int64 = 10

		SetLogger(log.NewNoopLogger())
		SetBlockStore(&mockBlockStore{
			heightFn: func() int64 { return storeHeight },
			loadBlockMetaFn: func(h int64) *types.BlockMeta {
				if h == 8 {
					return nil // simulate missing block meta at height 8
				}
				return &types.BlockMeta{Header: types.Header{Height: h, Time: time.Now()}}
			},
		})

		result, err := BlockchainInfo(&rpctypes.Context{}, 5, 10)
		require.Nil(t, result)
		assert.ErrorContains(t, err, "block meta not found for height 8")
	})

	t.Run("all block metas present", func(t *testing.T) {
		var storeHeight int64 = 5

		SetLogger(log.NewNoopLogger())
		SetBlockStore(&mockBlockStore{
			heightFn: func() int64 { return storeHeight },
			loadBlockMetaFn: func(h int64) *types.BlockMeta {
				return &types.BlockMeta{Header: types.Header{Height: h, Time: time.Now()}}
			},
		})

		result, err := BlockchainInfo(&rpctypes.Context{}, 1, 5)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.BlockMetas, 5)
	})
}

func TestBlockHandler(t *testing.T) {
	t.Run("nil block meta", func(t *testing.T) {
		var height int64 = 5

		SetBlockStore(&mockBlockStore{
			heightFn:        func() int64 { return height },
			loadBlockMetaFn: func(int64) *types.BlockMeta { return nil },
		})

		result, err := Block(&rpctypes.Context{}, &height)
		require.Nil(t, result)
		assert.ErrorContains(t, err, "block meta not found for height 5")
	})

	t.Run("nil block", func(t *testing.T) {
		var height int64 = 5

		SetBlockStore(&mockBlockStore{
			heightFn: func() int64 { return height },
			loadBlockMetaFn: func(int64) *types.BlockMeta {
				return &types.BlockMeta{Header: types.Header{Height: height}}
			},
			loadBlockFn: func(int64) *types.Block { return nil },
		})

		result, err := Block(&rpctypes.Context{}, &height)
		require.Nil(t, result)
		assert.ErrorContains(t, err, "block not found for height 5")
	})

	t.Run("success", func(t *testing.T) {
		var height int64 = 5

		SetBlockStore(&mockBlockStore{
			heightFn: func() int64 { return height },
			loadBlockMetaFn: func(int64) *types.BlockMeta {
				return &types.BlockMeta{Header: types.Header{Height: height}}
			},
			loadBlockFn: func(int64) *types.Block {
				return &types.Block{Header: types.Header{Height: height}}
			},
		})

		result, err := Block(&rpctypes.Context{}, &height)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, height, result.BlockMeta.Header.Height)
		assert.Equal(t, height, result.Block.Header.Height)
	})
}

func TestCommitHandler(t *testing.T) {
	t.Run("nil block meta", func(t *testing.T) {
		var height int64 = 5

		SetBlockStore(&mockBlockStore{
			heightFn:        func() int64 { return height },
			loadBlockMetaFn: func(int64) *types.BlockMeta { return nil },
		})

		result, err := Commit(&rpctypes.Context{}, &height)
		require.Nil(t, result)
		assert.ErrorContains(t, err, "block meta not found for height 5")
	})

	t.Run("success canonical commit", func(t *testing.T) {
		var (
			height      int64 = 5
			storeHeight int64 = 10
		)

		SetBlockStore(&mockBlockStore{
			heightFn: func() int64 { return storeHeight },
			loadBlockMetaFn: func(int64) *types.BlockMeta {
				return &types.BlockMeta{Header: types.Header{Height: height}}
			},
			loadBlockCommitFn: func(int64) *types.Commit { return &types.Commit{} },
		})

		result, err := Commit(&rpctypes.Context{}, &height)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.CanonicalCommit)
	})

	t.Run("success non-canonical commit", func(t *testing.T) {
		var height int64 = 10

		SetBlockStore(&mockBlockStore{
			heightFn: func() int64 { return height }, // storeHeight == height
			loadBlockMetaFn: func(int64) *types.BlockMeta {
				return &types.BlockMeta{Header: types.Header{Height: height}}
			},
			loadSeenCommitFn: func(int64) *types.Commit { return &types.Commit{} },
		})

		result, err := Commit(&rpctypes.Context{}, &height)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.CanonicalCommit)
	})
}
