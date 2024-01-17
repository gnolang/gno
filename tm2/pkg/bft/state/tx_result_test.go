package state

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	var (
		stateDB   = dbm.NewMemDB()
		txResults = generateTxResults(t, 100)
	)

	// Save the results
	for _, txResult := range txResults {
		saveTxResult(stateDB, txResult)
	}

	// Verify they are saved correctly
	for _, txResult := range txResults {
		dbResult, err := LoadTxResult(stateDB, txResult.Tx.Hash())
		require.NoError(t, err)

		assert.Equal(t, txResult, dbResult)
	}
}
