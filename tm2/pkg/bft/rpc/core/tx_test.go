package core

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
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
	t.Run("result found", func(t *testing.T) {
		// Prepare the transaction
		tx := &std.Tx{
			Memo: "example tx",
		}

		marshalledTx, err := amino.Marshal(tx)
		require.NoError(t, err)

		res := &types.TxResult{
			Height:   1,
			Index:    0,
			Tx:       marshalledTx,
			Response: abci.ResponseDeliverTx{},
		}

		// Prepare the DB
		sdb := memdb.NewMemDB()
		sdb.Set(state.CalcTxResultKey(res.Tx.Hash()), res.Bytes())

		// Set the GLOBALLY referenced db
		SetStateDB(sdb)

		// Load the result
		loadedTxResult, err := Tx(nil, res.Tx.Hash())

		require.NoError(t, err)
		require.NotNil(t, loadedTxResult)

		// Compare the result
		assert.Equal(t, res.Height, loadedTxResult.Height)
		assert.Equal(t, res.Index, loadedTxResult.Index)
		assert.Equal(t, res.Response, loadedTxResult.TxResult)
		assert.Equal(t, res.Tx, loadedTxResult.Tx)
		assert.Equal(t, res.Tx.Hash(), loadedTxResult.Tx.Hash())
	})

	t.Run("result not found", func(t *testing.T) {
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
		loadedTxResult, err := Tx(nil, hash)
		require.Nil(t, loadedTxResult)

		assert.Equal(t, expectedErr, err)
	})
}
