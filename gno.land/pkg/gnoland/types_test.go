package gnoland

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTxs generates dummy transactions
func generateTxs(t *testing.T, count int) []TxWithMetadata {
	t.Helper()

	txs := make([]TxWithMetadata, count)

	for i := 0; i < count; i++ {
		txs[i] = TxWithMetadata{
			Tx: std.Tx{
				Msgs: []std.Msg{
					bank.MsgSend{
						FromAddress: crypto.Address{byte(i)},
						ToAddress:   crypto.Address{byte(i)},
						Amount:      std.NewCoins(std.NewCoin(ugnot.Denom, 1)),
					},
				},
				Fee: std.Fee{
					GasWanted: 10,
					GasFee:    std.NewCoin(ugnot.Denom, 1000000),
				},
				Memo: fmt.Sprintf("tx %d", i),
			},
		}
	}

	return txs
}

func TestReadGenesisTxs(t *testing.T) {
	t.Parallel()

	createFile := func(path, data string) {
		file, err := os.Create(path)
		require.NoError(t, err)

		_, err = file.WriteString(data)
		require.NoError(t, err)
	}

	t.Run("invalid path", func(t *testing.T) {
		t.Parallel()

		path := "" // invalid

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		txs, err := ReadGenesisTxs(ctx, path)
		assert.Nil(t, txs)

		assert.Error(t, err)
	})

	t.Run("invalid tx format", func(t *testing.T) {
		t.Parallel()

		var (
			dir  = t.TempDir()
			path = filepath.Join(dir, "txs.jsonl")
		)

		// Create the file
		createFile(
			path,
			"random data",
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		txs, err := ReadGenesisTxs(ctx, path)
		assert.Nil(t, txs)

		assert.Error(t, err)
	})

	t.Run("valid txs", func(t *testing.T) {
		t.Parallel()

		var (
			dir  = t.TempDir()
			path = filepath.Join(dir, "txs.jsonl")
			txs  = generateTxs(t, 1000)
		)

		// Create the file
		file, err := os.Create(path)
		require.NoError(t, err)

		// Write the transactions
		for _, tx := range txs {
			encodedTx, err := amino.MarshalJSON(tx)
			require.NoError(t, err)

			_, err = file.WriteString(fmt.Sprintf("%s\n", encodedTx))
			require.NoError(t, err)
		}

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Load the transactions
		readTxs, err := ReadGenesisTxs(ctx, path)
		require.NoError(t, err)

		require.Len(t, readTxs, len(txs))

		for index, readTx := range readTxs {
			assert.Equal(t, txs[index], readTx)
		}
	})
}
