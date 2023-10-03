package backup

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/tx-archive/log/noop"
	"github.com/gnolang/tx-archive/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackup_DetermineRightBound(t *testing.T) {
	t.Parallel()

	t.Run("unable to fetch latest block number", func(t *testing.T) {
		t.Parallel()

		var (
			fetchErr   = errors.New("unable to fetch latest height")
			mockClient = &mockClient{
				getLatestBlockNumberFn: func() (uint64, error) {
					return 0, fetchErr
				},
			}
		)

		// Determine the right bound
		_, err := determineRightBound(mockClient, nil)

		assert.ErrorIs(t, err, fetchErr)
	})

	t.Run("excessive right range", func(t *testing.T) {
		t.Parallel()

		var (
			chainLatest uint64 = 10
			requestedTo        = chainLatest + 10 // > chain latest

			mockClient = &mockClient{
				getLatestBlockNumberFn: func() (uint64, error) {
					return chainLatest, nil
				},
			}
		)

		// Determine the right bound
		rightBound, err := determineRightBound(mockClient, &requestedTo)
		require.NoError(t, err)

		assert.Equal(t, chainLatest, rightBound)
	})

	t.Run("valid right range", func(t *testing.T) {
		t.Parallel()

		var (
			chainLatest uint64 = 10
			requestedTo        = chainLatest / 2 // < chain latest

			mockClient = &mockClient{
				getLatestBlockNumberFn: func() (uint64, error) {
					return chainLatest, nil
				},
			}
		)

		// Determine the right bound
		rightBound, err := determineRightBound(mockClient, &requestedTo)
		require.NoError(t, err)

		assert.Equal(t, requestedTo, rightBound)
	})
}

func TestBackup_ExecuteBackup(t *testing.T) {
	t.Parallel()

	var (
		tempFile = createTempFile(t)

		fromBlock uint64 = 10
		toBlock          = fromBlock + 10

		exampleTx = std.Tx{
			Memo: "example transaction",
		}

		cfg = DefaultConfig()

		mockClient = &mockClient{
			getLatestBlockNumberFn: func() (uint64, error) {
				return toBlock, nil
			},
			getBlockTransactionsFn: func(blockNum uint64) ([]std.Tx, error) {
				// Sanity check
				if blockNum < fromBlock && blockNum > toBlock {
					t.Fatal("invalid block number requested")
				}

				return []std.Tx{exampleTx}, nil
			},
		}
	)

	// Temp file cleanup
	t.Cleanup(func() {
		require.NoError(t, tempFile.Close())
		require.NoError(t, os.Remove(tempFile.Name()))
	})

	// Set the config
	cfg.FromBlock = fromBlock
	cfg.ToBlock = &toBlock
	cfg.OutputFile = tempFile.Name()
	cfg.Overwrite = true

	// Run the backup procedure
	require.NoError(t, ExecuteBackup(mockClient, noop.New(), cfg))

	// Read the output file
	archiveRaw, err := os.ReadFile(tempFile.Name())
	require.NoError(t, err)

	// Unmarshal the raw archive output
	var archive types.Archive

	require.NoError(t, json.Unmarshal(archiveRaw, &archive))

	// Validate the archive
	assert.Equal(t, fromBlock, archive.Metadata.EarliestBlockHeight)
	assert.Equal(t, toBlock, archive.Metadata.LatestBlockHeight)
	assert.Equal(t, int(toBlock-fromBlock+1), len(archive.BlockData))

	for index, block := range archive.BlockData {
		assert.Equal(t, uint64(index)+fromBlock, block.BlockNum)

		for _, tx := range block.Txs {
			assert.Equal(t, exampleTx, tx)
		}
	}
}
