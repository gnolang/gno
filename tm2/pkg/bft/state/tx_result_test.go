package state

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTxResults generates test transaction results
func generateTxResults(t *testing.T, count int) []*types.TxResult {
	t.Helper()

	results := make([]*types.TxResult, count)

	for i := 0; i < count; i++ {
		tx := &std.Tx{
			Memo: fmt.Sprintf("tx %d", i),
		}

		marshalledTx, err := amino.Marshal(tx)
		require.NoError(t, err)

		results[i] = &types.TxResult{
			Height:   10,
			Index:    uint32(i),
			Tx:       marshalledTx,
			Response: abci.ResponseDeliverTx{},
		}
	}

	return results
}

func TestStoreLoadTxResult(t *testing.T) {
	t.Parallel()

	t.Run("results found", func(t *testing.T) {
		t.Parallel()

		var (
			stateDB   = memdb.NewMemDB()
			txResults = generateTxResults(t, 100)
		)

		// Save the results
		for _, txResult := range txResults {
			saveTxResultIndex(
				stateDB,
				txResult.Tx.Hash(),
				TxResultIndex{
					BlockNum: txResult.Height,
					TxIndex:  txResult.Index,
				},
			)
		}

		// Verify they are saved correctly
		for _, txResult := range txResults {
			result := TxResultIndex{
				BlockNum: txResult.Height,
				TxIndex:  txResult.Index,
			}

			dbResult, err := LoadTxResultIndex(stateDB, txResult.Tx.Hash())
			require.NoError(t, err)

			assert.Equal(t, result.BlockNum, dbResult.BlockNum)
			assert.Equal(t, result.TxIndex, dbResult.TxIndex)
		}
	})

	t.Run("results not found", func(t *testing.T) {
		t.Parallel()

		var (
			stateDB   = memdb.NewMemDB()
			txResults = generateTxResults(t, 10)
		)

		// Verify they are not present
		for _, txResult := range txResults {
			_, err := LoadTxResultIndex(stateDB, txResult.Tx.Hash())

			expectedErr := NoTxResultForHashError{
				Hash: txResult.Tx.Hash(),
			}

			require.ErrorContains(t, err, expectedErr.Error())
		}
	})

	t.Run("results corrupted", func(t *testing.T) {
		t.Parallel()

		var (
			stateDB         = memdb.NewMemDB()
			corruptedResult = "totally valid amino"
			hash            = []byte("tx hash")
		)

		// Save the "corrupted" result to the DB
		stateDB.SetSync(CalcTxResultKey(hash), []byte(corruptedResult))

		txResult, err := LoadTxResultIndex(stateDB, hash)
		require.Nil(t, txResult)

		assert.ErrorIs(t, err, errTxResultIndexCorrupted)
	})
}
