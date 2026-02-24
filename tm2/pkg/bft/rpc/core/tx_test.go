package core

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxHandler(t *testing.T) {
	// Tests are not run in parallel because the JSON-RPC
	// handlers utilize global package-level variables,
	// that are not friendly with concurrent test runs (or anything, really)
	t.Run("tx result generated", func(t *testing.T) {
		var (
			height = int64(10)

			stdTx = &std.Tx{
				Memo: "example tx",
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

		// Prepare the transaction
		marshalledTx, err := amino.Marshal(stdTx)
		require.NoError(t, err)

		tx := types.Tx(marshalledTx)

		// Prepare the DB
		sdb := memdb.NewMemDB()

		// Save the result index to the DB
		sdb.Set(state.CalcTxResultKey(tx.Hash()), txResultIndex.Bytes())

		// Save the ABCI response to the DB
		sdb.Set(state.CalcABCIResponsesKey(height), responses.Bytes())

		// Set the GLOBALLY referenced db
		SetStateDB(sdb)

		// Set the GLOBALLY referenced blockstore
		blockStore := &mockBlockStore{
			heightFn: func() int64 {
				return height
			},
			loadBlockFn: func(h int64) *types.Block {
				require.Equal(t, height, h)

				return &types.Block{
					Data: types.Data{
						Txs: []types.Tx{
							tx,
						},
					},
				}
			},
		}

		SetBlockStore(blockStore)

		// Load the result
		loadedTxResult, err := Tx(&rpctypes.Context{}, tx.Hash())

		require.NoError(t, err)
		require.NotNil(t, loadedTxResult)

		// Compare the result
		assert.Equal(t, txResultIndex.BlockNum, loadedTxResult.Height)
		assert.Equal(t, txResultIndex.TxIndex, loadedTxResult.Index)
		assert.Equal(t, responses.DeliverTxs[0], loadedTxResult.TxResult)
		assert.Equal(t, tx, loadedTxResult.Tx)
		assert.Equal(t, tx.Hash(), loadedTxResult.Tx.Hash())
	})

	t.Run("tx result index not found", func(t *testing.T) {
		var (
			sdb         = memdb.NewMemDB()
			hash        = []byte("hash")
			expectedErr = state.NoTxResultForHashError{
				Hash: hash,
			}
		)

		// Set the GLOBALLY referenced db
		SetStateDB(sdb)

		// Load the result
		loadedTxResult, err := Tx(&rpctypes.Context{}, hash)
		require.Nil(t, loadedTxResult)

		assert.Equal(t, expectedErr, err)
	})

	t.Run("invalid block transaction index", func(t *testing.T) {
		var (
			height = int64(10)

			stdTx = &std.Tx{
				Memo: "example tx",
			}

			txResultIndex = state.TxResultIndex{
				BlockNum: height,
				TxIndex:  0,
			}
		)

		// Prepare the transaction
		marshalledTx, err := amino.Marshal(stdTx)
		require.NoError(t, err)

		tx := types.Tx(marshalledTx)

		// Prepare the DB
		sdb := memdb.NewMemDB()

		// Save the result index to the DB
		sdb.Set(state.CalcTxResultKey(tx.Hash()), txResultIndex.Bytes())

		// Set the GLOBALLY referenced db
		SetStateDB(sdb)

		// Set the GLOBALLY referenced blockstore
		blockStore := &mockBlockStore{
			heightFn: func() int64 {
				return height
			},
			loadBlockFn: func(h int64) *types.Block {
				require.Equal(t, height, h)

				return &types.Block{
					Data: types.Data{
						Txs: []types.Tx{}, // empty
					},
				}
			},
		}

		SetBlockStore(blockStore)

		// Load the result
		loadedTxResult, err := Tx(&rpctypes.Context{}, tx.Hash())
		require.Nil(t, loadedTxResult)

		assert.ErrorContains(t, err, "unable to get block transaction")
	})

	t.Run("invalid ABCI response index (corrupted state)", func(t *testing.T) {
		var (
			height = int64(10)

			stdTx = &std.Tx{
				Memo: "example tx",
			}

			txResultIndex = state.TxResultIndex{
				BlockNum: height,
				TxIndex:  0,
			}
		)

		// Prepare the transaction
		marshalledTx, err := amino.Marshal(stdTx)
		require.NoError(t, err)

		tx := types.Tx(marshalledTx)

		// Prepare the DB
		sdb := memdb.NewMemDB()

		// Save the result index to the DB
		sdb.Set(state.CalcTxResultKey(tx.Hash()), txResultIndex.Bytes())

		// Set the GLOBALLY referenced db
		SetStateDB(sdb)

		// Set the GLOBALLY referenced blockstore
		blockStore := &mockBlockStore{
			heightFn: func() int64 {
				return height
			},
			loadBlockFn: func(h int64) *types.Block {
				require.Equal(t, height, h)

				return &types.Block{
					Data: types.Data{
						Txs: []types.Tx{
							tx,
						},
					},
				}
			},
		}

		SetBlockStore(blockStore)

		// Load the result
		loadedTxResult, err := Tx(&rpctypes.Context{}, tx.Hash())
		require.Nil(t, loadedTxResult)

		assert.ErrorContains(t, err, "unable to load block results")
	})
}
