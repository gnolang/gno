package backup

import (
	"bufio"
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/tx-archive/backup/client"
	"github.com/gnolang/tx-archive/backup/writer/standard"
	"github.com/gnolang/tx-archive/log/noop"
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

func TestBackup_ExecuteBackup_FixedRange(t *testing.T) {
	t.Parallel()

	var (
		tempFile = createTempFile(t)

		fromBlock uint64 = 10
		toBlock          = fromBlock + 10

		exampleTx = std.Tx{
			Memo: "example transaction",
		}

		cfg = DefaultConfig()

		blockTime = time.Date(1987, 0o6, 24, 6, 32, 11, 0, time.FixedZone("Europe/Madrid", 0))

		mockClient = &mockClient{
			getLatestBlockNumberFn: func() (uint64, error) {
				return toBlock, nil
			},
			getBlockFn: func(blockNum uint64) (*client.Block, error) {
				// Sanity check
				if blockNum < fromBlock && blockNum > toBlock {
					t.Fatal("invalid block number requested")
				}

				return &client.Block{
					Height:    blockNum,
					Timestamp: blockTime.Add(time.Duration(blockNum) * time.Minute).UnixMilli(),
					Txs:       []std.Tx{exampleTx},
				}, nil // 1 tx per block
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
		var txData gnoland.TxWithMetadata

		// Unmarshal the JSON data into the Person struct
		if err := amino.UnmarshalJSON(scanner.Bytes(), &txData); err != nil {
			t.Fatalf("unable to unmarshal JSON line, %v", err)
		}

		assert.Equal(t, exampleTx, txData.Tx)
		assert.Equal(
			t,
			blockTime.Add(time.Duration(expectedBlock)*time.Minute).Local(),
			time.UnixMilli(txData.Metadata.Timestamp),
		)

		expectedBlock++
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		t.Fatalf("error encountered during scan, %v", err)
	}
}

func TestBackup_ExecuteBackup_Watch(t *testing.T) {
	t.Parallel()

	// Set up the context that is controlled by the test
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	var (
		tempFile = createTempFile(t)

		fromBlock uint64 = 10
		toBlock          = fromBlock + 10

		requestToBlock = toBlock / 2

		exampleTx = std.Tx{
			Memo: "example transaction",
		}

		cfg = DefaultConfig()

		blockTime = time.Date(1987, 0o6, 24, 6, 32, 11, 0, time.FixedZone("Europe/Madrid", 0))

		mockClient = &mockClient{
			getLatestBlockNumberFn: func() (uint64, error) {
				return toBlock, nil
			},
			getBlockFn: func(blockNum uint64) (*client.Block, error) {
				// Sanity check
				if blockNum < fromBlock && blockNum > toBlock {
					t.Fatal("invalid block number requested")
				}

				if blockNum == toBlock {
					// End of the road, close the watch process
					cancelFn()
				}

				return &client.Block{
					Height:    blockNum,
					Timestamp: blockTime.Add(time.Duration(blockNum) * time.Minute).UnixMilli(),
					Txs:       []std.Tx{exampleTx},
				}, nil // 1 tx per block
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
	cfg.ToBlock = &requestToBlock
	cfg.Watch = true

	s := NewService(mockClient, standard.NewWriter(tempFile), WithLogger(noop.New()))
	s.watchInterval = 10 * time.Millisecond // make the interval almost instant for the test

	// Run the backup procedure
	require.NoError(
		t,
		s.ExecuteBackup(
			ctx,
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
		var txData gnoland.TxWithMetadata

		// Unmarshal the JSON data into the Person struct
		if err := amino.UnmarshalJSON(scanner.Bytes(), &txData); err != nil {
			t.Fatalf("unable to unmarshal JSON line, %v", err)
		}

		assert.Equal(t, exampleTx, txData.Tx)
		assert.Equal(
			t,
			blockTime.Add(time.Duration(expectedBlock)*time.Minute).Local(),
			time.UnixMilli(txData.Metadata.Timestamp),
		)

		expectedBlock++
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		t.Fatalf("error encountered during scan, %v", err)
	}
}
