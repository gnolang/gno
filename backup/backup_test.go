package backup

import (
	"bufio"
	"context"
	"errors"
	"os"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/tx-archive/backup/writer/standard"
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

				return []std.Tx{exampleTx}, nil // 1 tx per block
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

	s := NewService(mockClient, standard.NewWriter(tempFile), WithLogger(noop.New()))

	// Run the backup procedure
	require.NoError(
		t,
		s.ExecuteBackup(
			context.Background(),
			cfg,
		),
	)

	// Read the output file
	fileRaw, err := os.Open(tempFile.Name())
	require.NoError(t, err)

	// Set up a line-by-line scanner
	scanner := bufio.NewScanner(fileRaw)

	expectedBlock := fromBlock

	// Iterate over each line in the file
	for scanner.Scan() {
		var txData types.TxData

		// Unmarshal the JSON data into the Person struct
		if err := amino.UnmarshalJSON(scanner.Bytes(), &txData); err != nil {
			t.Fatalf("unable to unmarshal JSON line, %v", err)
		}

		assert.Equal(t, expectedBlock, txData.BlockNum)
		assert.Equal(t, exampleTx, txData.Tx)

		expectedBlock++
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		t.Fatalf("error encountered during scan, %v", err)
	}
}
