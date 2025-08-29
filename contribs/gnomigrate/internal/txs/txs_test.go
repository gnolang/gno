package txs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateDummyTxs generates dummy transactions
func generateDummyTxs(t *testing.T, count int) []std.Tx {
	t.Helper()

	txs := make([]std.Tx, count)

	for i := 0; i < count; i++ {
		txs[i] = std.Tx{
			Msgs: []std.Msg{
				bank.MsgSend{
					FromAddress: crypto.Address{byte(i)},
					ToAddress:   crypto.Address{byte((i + 1) % count)},
					Amount:      std.NewCoins(std.NewCoin(ugnot.Denom, 1)),
				},
			},
			Fee: std.Fee{
				GasWanted: 1,
				GasFee:    std.NewCoin(ugnot.Denom, 1000000),
			},
			Memo: fmt.Sprintf("tx %d", i),
		}
	}

	return txs
}

func TestMigrate_Txs(t *testing.T) {
	t.Parallel()

	t.Run("invalid input dir", func(t *testing.T) {
		t.Parallel()

		// Perform the migration
		cmd := NewTxsCmd(commands.NewTestIO())
		args := []string{
			"--input-dir",
			"",
			"--output-dir",
			t.TempDir(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errInvalidInputDir)
	})

	t.Run("invalid output dir", func(t *testing.T) {
		t.Parallel()

		// Perform the migration
		cmd := NewTxsCmd(commands.NewTestIO())
		args := []string{
			"--input-dir",
			t.TempDir(),
			"--output-dir",
			"",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorIs(t, cmdErr, errInvalidOutputDir)
	})

	t.Run("valid tx sheet migration", func(t *testing.T) {
		t.Parallel()

		var (
			inputDir  = t.TempDir()
			outputDir = t.TempDir()

			txs = generateDummyTxs(t, 10000)

			chunks    = 4
			chunkSize = len(txs) / chunks
		)

		getSheetPath := func(dir string, index int) string {
			return filepath.Join(dir, fmt.Sprintf("transactions-sheet-%d.jsonl", index))
		}

		// Generate the initial sheet files
		files := make([]*os.File, 0, chunks)
		for i := 0; i < chunks; i++ {
			f, err := os.Create(getSheetPath(inputDir, i))
			require.NoError(t, err)

			files = append(files, f)
		}

		for i := 0; i < chunks; i++ {
			var (
				start = i * chunkSize
				end   = start + chunkSize
			)

			if end > len(txs) {
				end = len(txs)
			}

			tx := txs[start:end]

			f := files[i]

			jsonData, err := amino.MarshalJSON(tx)
			require.NoError(t, err)

			_, err = fmt.Fprintf(f, "%s\n", jsonData)
			require.NoError(t, err)
		}

		// Perform the migration
		cmd := NewTxsCmd(commands.NewTestIO())
		args := []string{
			"--input-dir",
			inputDir,
			"--output-dir",
			outputDir,
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		require.NoError(t, cmdErr)

		metadataTxs := make([]gnoland.TxWithMetadata, 0, len(txs))
		for i := 0; i < chunks; i++ {
			readTxs, err := gnoland.ReadGenesisTxs(context.Background(), getSheetPath(outputDir, i))
			require.NoError(t, err)

			metadataTxs = append(metadataTxs, readTxs...)
		}

		// Make sure the metadata txs match
		for index, tx := range metadataTxs {
			assert.Equal(t, txs[index], tx.Tx)
		}
	})
}
