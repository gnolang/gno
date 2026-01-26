package tx

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/mock"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_TxHandler(t *testing.T) {
	t.Parallel()

	t.Run("Missing hash param", func(t *testing.T) {
		t.Parallel()

		var (
			sdb            = memdb.NewMemDB()
			mockBlockStore = &mock.BlockStore{
				HeightFn: func() int64 {
					t.FailNow()

					return 0
				},
			}
		)

		h := NewHandler(mockBlockStore, sdb)

		res, err := h.TxHandler(nil, nil)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Tx result index not found", func(t *testing.T) {
		t.Parallel()

		var (
			sdb            = memdb.NewMemDB()
			hash           = []byte("hash")
			mockBlockStore = &mock.BlockStore{
				HeightFn: func() int64 {
					t.FailNow()

					return 0
				},
			}
		)

		h := NewHandler(mockBlockStore, sdb)

		res, err := h.TxHandler(nil, []any{hash})
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
	})

	t.Run("Invalid tx height (> store height)", func(t *testing.T) {
		t.Parallel()

		var (
			storeHeight = int64(9)
			txHeight    = int64(10)

			mockBlockStore = &mock.BlockStore{
				HeightFn: func() int64 {
					return storeHeight
				},
				LoadBlockFn: func(_ int64) *types.Block {
					t.FailNow()

					return nil
				},
			}

			idx = state.TxResultIndex{
				BlockNum: txHeight,
				TxIndex:  0,
			}
		)

		stdTx := &std.Tx{
			Memo: "example tx",
		}
		raw, err := amino.Marshal(stdTx)
		require.NoError(t, err)

		tx := types.Tx(raw)

		sdb := memdb.NewMemDB()
		require.NoError(t, sdb.Set(state.CalcTxResultKey(tx.Hash()), idx.Bytes()))

		h := NewHandler(mockBlockStore, sdb)

		res, e := h.TxHandler(nil, []any{tx.Hash()})
		require.Nil(t, res)
		require.NotNil(t, e)

		assert.Equal(t, spec.ServerErrorCode, e.Code)
	})

	t.Run("Block not found", func(t *testing.T) {
		t.Parallel()

		var (
			height = int64(10)

			mockBlockStore = &mock.BlockStore{
				HeightFn: func() int64 { return height },
				LoadBlockFn: func(hh int64) *types.Block {
					assert.Equal(t, height, hh)

					return nil
				},
			}

			idx = state.TxResultIndex{
				BlockNum: height,
				TxIndex:  0,
			}
		)

		stdTx := &std.Tx{
			Memo: "example tx",
		}
		raw, err := amino.Marshal(stdTx)
		require.NoError(t, err)

		tx := types.Tx(raw)

		sdb := memdb.NewMemDB()
		require.NoError(t, sdb.Set(state.CalcTxResultKey(tx.Hash()), idx.Bytes()))

		h := NewHandler(mockBlockStore, sdb)

		res, e := h.TxHandler(nil, []any{tx.Hash()})
		require.Nil(t, res)
		require.NotNil(t, e)

		assert.Equal(t, spec.ServerErrorCode, e.Code)
	})

	t.Run("Invalid block transaction index (empty txs)", func(t *testing.T) {
		t.Parallel()

		var (
			height         = int64(10)
			mockBlockStore = &mock.BlockStore{
				HeightFn: func() int64 { return height },
				LoadBlockFn: func(hh int64) *types.Block {
					assert.Equal(t, height, hh)
					return &types.Block{
						Data: types.Data{
							Txs: []types.Tx{},
						},
					}
				},
			}

			idx = state.TxResultIndex{
				BlockNum: height,
				TxIndex:  0,
			}
		)

		stdTx := &std.Tx{
			Memo: "example tx",
		}
		raw, err := amino.Marshal(stdTx)
		require.NoError(t, err)

		tx := types.Tx(raw)

		sdb := memdb.NewMemDB()
		require.NoError(t, sdb.Set(state.CalcTxResultKey(tx.Hash()), idx.Bytes()))

		h := NewHandler(mockBlockStore, sdb)

		res, e := h.TxHandler(nil, []any{tx.Hash()})
		require.Nil(t, res)
		require.NotNil(t, e)

		assert.Equal(t, spec.ServerErrorCode, e.Code)
	})

	t.Run("Unable to load block results", func(t *testing.T) {
		t.Parallel()

		stdTx := &std.Tx{
			Memo: "example tx",
		}
		raw, err := amino.Marshal(stdTx)
		require.NoError(t, err)

		tx := types.Tx(raw)

		var (
			height         = int64(10)
			mockBlockStore = &mock.BlockStore{
				HeightFn: func() int64 { return height },
				LoadBlockFn: func(hh int64) *types.Block {
					assert.Equal(t, height, hh)
					return &types.Block{
						Data: types.Data{
							Txs: []types.Tx{tx},
						},
					}
				},
			}

			idx = state.TxResultIndex{
				BlockNum: height,
				TxIndex:  0,
			}
		)

		sdb := memdb.NewMemDB()
		require.NoError(t, sdb.Set(state.CalcTxResultKey(tx.Hash()), idx.Bytes()))

		h := NewHandler(mockBlockStore, sdb)

		res, e := h.TxHandler(nil, []any{tx.Hash()})
		require.Nil(t, res)
		require.NotNil(t, e)

		assert.Equal(t, spec.ServerErrorCode, e.Code)
	})

	t.Run("Invalid ABCI response index", func(t *testing.T) {
		t.Parallel()

		stdTx := &std.Tx{
			Memo: "example tx",
		}
		raw, err := amino.Marshal(stdTx)
		require.NoError(t, err)

		tx := types.Tx(raw)

		var (
			height = int64(10)

			mockBlockStore = &mock.BlockStore{
				HeightFn: func() int64 { return height },
				LoadBlockFn: func(hh int64) *types.Block {
					assert.Equal(t, height, hh)
					return &types.Block{
						Data: types.Data{
							Txs: []types.Tx{tx},
						},
					}
				},
			}

			idx = state.TxResultIndex{
				BlockNum: height,
				TxIndex:  0,
			}

			responses = &state.ABCIResponses{
				DeliverTxs: []abci.ResponseDeliverTx{}, // empty
			}
		)

		sdb := memdb.NewMemDB()
		require.NoError(t, sdb.Set(state.CalcTxResultKey(tx.Hash()), idx.Bytes()))
		require.NoError(t, sdb.Set(state.CalcABCIResponsesKey(height), responses.Bytes()))

		h := NewHandler(mockBlockStore, sdb)

		res, e := h.TxHandler(nil, []any{tx.Hash()})
		require.Nil(t, res)
		require.NotNil(t, e)

		assert.Equal(t, spec.ServerErrorCode, e.Code)
	})

	t.Run("Valid tx result", func(t *testing.T) {
		t.Parallel()

		stdTx := &std.Tx{
			Memo: "example tx",
		}
		raw, err := amino.Marshal(stdTx)
		require.NoError(t, err)

		tx := types.Tx(raw)

		var (
			height = int64(10)

			mockBlockStore = &mock.BlockStore{
				HeightFn: func() int64 { return height },
				LoadBlockFn: func(hh int64) *types.Block {
					require.Equal(t, height, hh)
					return &types.Block{
						Data: types.Data{
							Txs: []types.Tx{tx},
						},
					}
				},
			}

			txResultIndex = state.TxResultIndex{
				BlockNum: height,
				TxIndex:  0,
			}

			responses = &state.ABCIResponses{
				DeliverTxs: []abci.ResponseDeliverTx{
					{
						GasWanted: 100,
					},
				},
			}
		)

		sdb := memdb.NewMemDB()
		require.NoError(t, sdb.Set(state.CalcTxResultKey(tx.Hash()), txResultIndex.Bytes()))
		require.NoError(t, sdb.Set(state.CalcABCIResponsesKey(height), responses.Bytes()))

		h := NewHandler(mockBlockStore, sdb)

		out, e := h.TxHandler(nil, []any{tx.Hash()})
		require.Nil(t, e)
		require.NotNil(t, out)

		result, ok := out.(*ResultTx)
		require.True(t, ok)

		assert.Equal(t, txResultIndex.BlockNum, result.Height)
		assert.Equal(t, txResultIndex.TxIndex, result.Index)
		assert.Equal(t, responses.DeliverTxs[0], result.TxResult)
		assert.Equal(t, tx, result.Tx)
		assert.Equal(t, tx.Hash(), result.Tx.Hash())
		assert.Equal(t, tx.Hash(), result.Hash)
	})
}
