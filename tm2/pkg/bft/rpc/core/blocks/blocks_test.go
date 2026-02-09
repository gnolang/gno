package blocks

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/mock"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

func TestHandler_BlockchainInfoHandler(t *testing.T) {
	t.Parallel()

	t.Run("Invalid min height param", func(t *testing.T) {
		t.Parallel()

		var (
			store  = &mock.BlockStore{}
			params = []any{"foo", int64(10)}
		)

		h := NewHandler(store, nil)

		res, err := h.BlockchainInfoHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Filter error negative heights", func(t *testing.T) {
		t.Parallel()

		const storeHeight int64 = 10

		var (
			store = &mock.BlockStore{
				HeightFn: func() int64 {
					return storeHeight
				},
			}
			params = []any{int64(-1), int64(5)}
		)

		h := NewHandler(store, nil)

		res, err := h.BlockchainInfoHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
	})

	t.Run("Valid range default (no params)", func(t *testing.T) {
		t.Parallel()

		const storeHeight int64 = 5

		var (
			metas = map[int64]*types.BlockMeta{}
			store = &mock.BlockStore{
				HeightFn: func() int64 { return storeHeight },
				LoadBlockMetaFn: func(h int64) *types.BlockMeta {
					return metas[h]
				},
			}
		)

		// Update meta range
		for h := int64(1); h <= storeHeight; h++ {
			metas[h] = &types.BlockMeta{
				Header: types.Header{Height: h},
			}
		}

		h := NewHandler(store, nil)

		res, err := h.BlockchainInfoHandler(nil, nil)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultBlockchainInfo)
		require.True(t, ok)

		assert.Equal(t, storeHeight, result.LastHeight)
		require.Len(t, result.BlockMetas, int(storeHeight))

		expectedHeight := storeHeight
		for i := 0; i < int(storeHeight); i++ {
			assert.Equal(t, expectedHeight, result.BlockMetas[i].Header.Height)

			expectedHeight--
		}
	})

	t.Run("Valid range limited to 20", func(t *testing.T) {
		t.Parallel()

		const storeHeight int64 = 30

		var (
			metas = map[int64]*types.BlockMeta{}
			store = &mock.BlockStore{
				HeightFn: func() int64 {
					return storeHeight
				},
				LoadBlockMetaFn: func(h int64) *types.BlockMeta {
					return metas[h]
				},
			}
		)

		// Update the meta range
		for h := int64(1); h <= storeHeight; h++ {
			metas[h] = &types.BlockMeta{
				Header: types.Header{Height: h},
			}
		}

		h := NewHandler(store, nil)

		res, err := h.BlockchainInfoHandler(nil, nil)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultBlockchainInfo)
		require.True(t, ok)

		require.Len(t, result.BlockMetas, 20)

		expectedHeight := storeHeight
		for i := 0; i < 20; i++ {
			assert.Equal(t, expectedHeight, result.BlockMetas[i].Header.Height)

			expectedHeight--
		}
	})
}

func TestHandler_BlockHandler(t *testing.T) {
	t.Parallel()

	t.Run("Invalid height param", func(t *testing.T) {
		t.Parallel()

		var (
			store  = &mock.BlockStore{}
			params = []any{"foo"}
		)

		h := NewHandler(store, nil)

		res, err := h.BlockHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Height below minimum", func(t *testing.T) {
		t.Parallel()

		const storeHeight int64 = 10

		var (
			store = &mock.BlockStore{
				HeightFn: func() int64 { return storeHeight },
			}
			params = []any{int64(-1)}
		)

		h := NewHandler(store, nil)

		res, err := h.BlockHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
	})

	t.Run("Height above latest", func(t *testing.T) {
		t.Parallel()

		const storeHeight int64 = 10

		var (
			store = &mock.BlockStore{
				HeightFn: func() int64 { return storeHeight },
			}
			params = []any{storeHeight + 1}
		)

		h := NewHandler(store, nil)

		res, err := h.BlockHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
	})

	t.Run("Block meta missing", func(t *testing.T) {
		t.Parallel()

		const storeHeight int64 = 10

		var (
			store = &mock.BlockStore{
				HeightFn: func() int64 {
					return storeHeight
				},
				LoadBlockMetaFn: func(_ int64) *types.BlockMeta {
					return nil // explicit
				},
			}
			params = []any{storeHeight}
		)

		h := NewHandler(store, nil)

		res, err := h.BlockHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
	})

	t.Run("Block missing", func(t *testing.T) {
		t.Parallel()

		const storeHeight int64 = 10

		var (
			store = &mock.BlockStore{
				HeightFn: func() int64 { return storeHeight },
				LoadBlockMetaFn: func(h int64) *types.BlockMeta {
					if h == storeHeight {
						return &types.BlockMeta{
							Header: types.Header{
								Height: h,
							},
						}
					}

					return nil
				},
				LoadBlockFn: func(_ int64) *types.Block {
					return nil // explicit
				},
			}
			params = []any{storeHeight}
		)

		h := NewHandler(store, nil)

		res, err := h.BlockHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
	})

	t.Run("Valid block latest by default", func(t *testing.T) {
		t.Parallel()

		const storeHeight int64 = 10

		var (
			meta = &types.BlockMeta{
				Header: types.Header{
					Height: storeHeight,
				},
			}
			block = &types.Block{
				Header: types.Header{
					Height: storeHeight,
				},
			}

			store = &mock.BlockStore{
				HeightFn: func() int64 { return storeHeight },
				LoadBlockMetaFn: func(h int64) *types.BlockMeta {
					if h == storeHeight {
						return meta
					}

					return nil
				},
				LoadBlockFn: func(h int64) *types.Block {
					if h == storeHeight {
						return block
					}

					return nil
				},
			}
		)

		h := NewHandler(store, nil)

		res, err := h.BlockHandler(nil, nil)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultBlock)
		require.True(t, ok)

		assert.Equal(t, meta, result.BlockMeta)
		assert.Equal(t, block, result.Block)
	})

	t.Run("Valid block at explicit height", func(t *testing.T) {
		t.Parallel()

		const (
			storeHeight int64 = 10
			blockHeight int64 = 7
		)

		var (
			meta = &types.BlockMeta{
				Header: types.Header{
					Height: blockHeight,
				},
			}
			block = &types.Block{
				Header: types.Header{
					Height: blockHeight,
				},
			}

			store = &mock.BlockStore{
				HeightFn: func() int64 {
					return storeHeight
				},
				LoadBlockMetaFn: func(h int64) *types.BlockMeta {
					if h == blockHeight {
						return meta
					}

					return nil
				},
				LoadBlockFn: func(h int64) *types.Block {
					if h == blockHeight {
						return block
					}

					return nil
				},
			}
		)

		h := NewHandler(store, nil)

		res, err := h.BlockHandler(nil, []any{blockHeight})
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultBlock)
		require.True(t, ok)

		assert.Same(t, meta, result.BlockMeta)
		assert.Same(t, block, result.Block)
	})
}

func TestHandler_CommitHandler(t *testing.T) {
	t.Parallel()

	t.Run("Invalid height param", func(t *testing.T) {
		t.Parallel()

		var (
			store  = &mock.BlockStore{}
			h      = NewHandler(store, nil)
			params = []any{"foo"}
		)

		res, err := h.CommitHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Height below minimum", func(t *testing.T) {
		t.Parallel()

		const storeHeight int64 = 10

		var (
			store = &mock.BlockStore{
				HeightFn: func() int64 { return storeHeight },
			}
			params = []any{int64(-1)}
		)

		h := NewHandler(store, nil)

		res, err := h.CommitHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
	})

	t.Run("Block meta missing", func(t *testing.T) {
		t.Parallel()

		const storeHeight int64 = 10

		var (
			store = &mock.BlockStore{
				HeightFn: func() int64 {
					return storeHeight
				},
				LoadBlockMetaFn: func(_ int64) *types.BlockMeta {
					return nil // explicit
				},
			}
			params = []any{storeHeight}
		)

		h := NewHandler(store, nil)

		res, err := h.CommitHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
	})

	t.Run("Seen commit missing at latest", func(t *testing.T) {
		t.Parallel()

		const storeHeight int64 = 10

		var (
			meta = &types.BlockMeta{
				Header: types.Header{
					Height: storeHeight,
				},
			}

			store = &mock.BlockStore{
				HeightFn: func() int64 {
					return storeHeight
				},
				LoadBlockMetaFn: func(h int64) *types.BlockMeta {
					if h == storeHeight {
						return meta
					}

					return nil
				},
				LoadBlockCommitFn: func(_ int64) *types.Commit {
					return nil // explicit
				},
			}
			params = []any{storeHeight}
		)

		h := NewHandler(store, nil)

		res, err := h.CommitHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
	})

	t.Run("Canonical commit missing for past height", func(t *testing.T) {
		t.Parallel()

		const (
			storeHeight  int64 = 10
			targetHeight int64 = 9
		)

		var (
			meta = &types.BlockMeta{
				Header: types.Header{
					Height: storeHeight,
				},
			}

			store = &mock.BlockStore{
				HeightFn: func() int64 {
					return storeHeight
				},
				LoadBlockMetaFn: func(h int64) *types.BlockMeta {
					if h == targetHeight {
						return meta
					}

					return nil
				},
				LoadBlockCommitFn: func(_ int64) *types.Commit {
					return nil // explicit
				},
			}
			params = []any{targetHeight}
		)

		h := NewHandler(store, nil)

		res, err := h.CommitHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
	})

	t.Run("Non-canonical commit at latest height", func(t *testing.T) {
		t.Parallel()

		const storeHeight int64 = 10

		var (
			meta = &types.BlockMeta{
				Header: types.Header{
					Height: storeHeight,
				},
			}
			commit = &types.Commit{}

			store = &mock.BlockStore{
				HeightFn: func() int64 {
					return storeHeight
				},
				LoadBlockMetaFn: func(h int64) *types.BlockMeta {
					if h == storeHeight {
						return meta
					}

					return nil
				},
				LoadSeenCommitFn: func(h int64) *types.Commit {
					if h == storeHeight {
						return commit
					}

					return nil
				},
			}
		)

		h := NewHandler(store, nil)

		res, err := h.CommitHandler(nil, []any{storeHeight})
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultCommit)
		require.True(t, ok)

		assert.False(t, result.CanonicalCommit)
	})

	t.Run("Canonical commit at past height", func(t *testing.T) {
		t.Parallel()

		const (
			storeHeight  int64 = 10
			targetHeight int64 = 9
		)

		store := &mock.BlockStore{
			HeightFn: func() int64 {
				return storeHeight
			},
			LoadBlockMetaFn: func(h int64) *types.BlockMeta {
				if h == targetHeight {
					return &types.BlockMeta{
						Header: types.Header{
							Height: h,
						},
					}
				}

				return nil
			},
			LoadBlockCommitFn: func(h int64) *types.Commit {
				if h == targetHeight {
					return &types.Commit{}
				}

				return nil
			},
		}

		h := NewHandler(store, nil)

		res, err := h.CommitHandler(nil, []any{targetHeight})
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultCommit)
		require.True(t, ok)

		assert.True(t, result.CanonicalCommit)
	})
}

func TestHandler_BlockResultsHandler(t *testing.T) {
	t.Parallel()

	t.Run("Invalid height param", func(t *testing.T) {
		t.Parallel()

		var (
			store   = &mock.BlockStore{}
			stateDB = memdb.NewMemDB()
			params  = []any{"foo"}
		)

		h := NewHandler(store, stateDB)

		res, err := h.BlockResultsHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Height above latest", func(t *testing.T) {
		t.Parallel()

		const storeHeight int64 = 10

		var (
			store = &mock.BlockStore{
				HeightFn: func() int64 { return storeHeight },
			}
			stateDB = memdb.NewMemDB()
			params  = []any{storeHeight + 1}
		)

		h := NewHandler(store, stateDB)

		res, err := h.BlockResultsHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
	})

	t.Run("ABCI response load error", func(t *testing.T) {
		t.Parallel()

		const storeHeight int64 = 10

		var (
			store = &mock.BlockStore{
				HeightFn: func() int64 { return storeHeight },
			}
			stateDB = memdb.NewMemDB()
			params  = []any{storeHeight}
		)

		h := NewHandler(store, stateDB)

		res, err := h.BlockResultsHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
	})

	t.Run("Valid block results", func(t *testing.T) {
		t.Parallel()

		const (
			storeHeight  int64 = 10
			targetHeight int64 = 7
		)

		var (
			expectedResponses = &sm.ABCIResponses{}

			store = &mock.BlockStore{
				HeightFn: func() int64 {
					return storeHeight
				},
			}
			stateDB = memdb.NewMemDB()
		)

		h := NewHandler(store, stateDB)

		require.NotPanics(t, func() {
			sm.SaveABCIResponses(stateDB, targetHeight, expectedResponses)
		})

		res, err := h.BlockResultsHandler(nil, []any{targetHeight})
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultBlockResults)
		require.True(t, ok)

		assert.Equal(t, targetHeight, result.Height)
		assert.NotNil(t, result.Results)
	})
}

func TestFilterMinMax(t *testing.T) {
	t.Parallel()

	t.Run("Negative heights", func(t *testing.T) {
		t.Parallel()

		_, _, err := filterMinMax(10, -1, 5, 20)
		require.Error(t, err)

		assert.Contains(t, err.Error(), "heights must be non-negative")
	})

	t.Run("Defaults within limit", func(t *testing.T) {
		t.Parallel()

		low, high, err := filterMinMax(10, 0, 0, 20)
		require.NoError(t, err)

		assert.Equal(t, int64(1), low)
		assert.Equal(t, int64(10), high)
	})

	t.Run("Clamp high to current height", func(t *testing.T) {
		t.Parallel()

		low, high, err := filterMinMax(10, 5, 100, 20)
		require.NoError(t, err)

		assert.Equal(t, int64(5), low)
		assert.Equal(t, int64(10), high)
	})

	t.Run("Limit window size", func(t *testing.T) {
		t.Parallel()

		low, high, err := filterMinMax(100, 1, 100, 20)
		require.NoError(t, err)

		assert.Equal(t, int64(81), low)
		assert.Equal(t, int64(100), high)
	})

	t.Run("Low greater than high", func(t *testing.T) {
		t.Parallel()

		low, high, err := filterMinMax(5, 10, 1, 20)
		require.Error(t, err)

		assert.Greater(t, low, high)
		assert.Contains(t, err.Error(), "min height")
	})
}

func TestFilterMinMax_Legacy(t *testing.T) {
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
