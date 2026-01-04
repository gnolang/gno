package backup

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/contribs/tx-archive/backup/client"
	"github.com/gnolang/gno/contribs/tx-archive/backup/writer/standard"
	"github.com/gnolang/gno/contribs/tx-archive/log/noop"
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

func generateMemo(blockNum, txNum uint64) string {
	return fmt.Sprintf("example transaction %d:%d", blockNum, txNum)
}

var blockTime = time.Date(1987, 0o6, 24, 6, 32, 11, 0, time.FixedZone("Europe/Madrid", 0))

func generateBlocks(t *testing.T, from, to, nTxs uint64) []*client.Block {
	t.Helper()

	// generateBlocks return only blocks containing transaction
	if nTxs == 0 {
		return nil
	}

	blocks := make([]*client.Block, to-from+1)

	for i := range blocks {
		txs := make([]std.Tx, nTxs)
		blockNum := from + uint64(i)

		for j := range txs {
			txs[j].Memo = generateMemo(blockNum, uint64(j))
		}

		blocks[i] = &client.Block{
			Height:    blockNum,
			Timestamp: blockTime.Add(time.Duration(blockNum) * time.Minute).UnixMilli(),
			Txs:       txs,
		}
	}

	return blocks
}

type testCase struct {
	name        string
	batchSize   uint
	fromBlock   uint64
	toBlock     uint64
	txsPerBlock uint64
}

var testCases = []testCase{
	// Batch 0 (should be forced to fetch by 1 by config)
	{name: "batch 0/10 blocks/3 txs", batchSize: 0, fromBlock: 1, toBlock: 10, txsPerBlock: 3},
	{name: "batch 0/10 blocks/1 tx", batchSize: 0, fromBlock: 1, toBlock: 10, txsPerBlock: 1},
	{name: "batch 0/10 blocks/0 tx", batchSize: 0, fromBlock: 1, toBlock: 10, txsPerBlock: 0},
	// Batch 1 (fetch 1 by 1)
	{name: "batch 1/10 blocks/3 txs", batchSize: 1, fromBlock: 1, toBlock: 10, txsPerBlock: 3},
	{name: "batch 1/10 blocks/1 tx", batchSize: 1, fromBlock: 1, toBlock: 10, txsPerBlock: 1},
	{name: "batch 1/10 blocks/0 tx", batchSize: 1, fromBlock: 1, toBlock: 10, txsPerBlock: 0},
	// Batch 6 (first fetch 6, then 4)
	{name: "batch 6/10 blocks/3 txs", batchSize: 6, fromBlock: 1, toBlock: 10, txsPerBlock: 3},
	{name: "batch 6/10 blocks/1 tx", batchSize: 6, fromBlock: 1, toBlock: 10, txsPerBlock: 1},
	{name: "batch 6/10 blocks/0 tx", batchSize: 6, fromBlock: 1, toBlock: 10, txsPerBlock: 0},
	// Batch 10 (fetch all blocks in 1 batch)
	{name: "batch 10/10 blocks/3 txs", batchSize: 10, fromBlock: 1, toBlock: 10, txsPerBlock: 3},
	{name: "batch 10/10 blocks/1 tx", batchSize: 10, fromBlock: 1, toBlock: 10, txsPerBlock: 1},
	{name: "batch 10/10 blocks/0 tx", batchSize: 10, fromBlock: 1, toBlock: 10, txsPerBlock: 0},
	// Batch 11 (batch size (11) bigger than block count within range (10))
	{name: "batch 11/10 blocks/3 txs", batchSize: 11, fromBlock: 1, toBlock: 10, txsPerBlock: 3},
	{name: "batch 11/10 blocks/1 tx", batchSize: 11, fromBlock: 1, toBlock: 10, txsPerBlock: 1},
	{name: "batch 11/10 blocks/0 tx", batchSize: 11, fromBlock: 1, toBlock: 10, txsPerBlock: 0},
}

func TestBackup_ExecuteBackup_FixedRange(t *testing.T) {
	t.Parallel()

	//nolint:thelper,gocritic
	testFunc := func(t *testing.T, tCase testCase) {
		t.Run(tCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				tempFile = createTempFile(t)
				cfg      = DefaultConfig()

				mockClient = &mockClient{
					getLatestBlockNumberFn: func() (uint64, error) {
						return tCase.toBlock, nil
					},
					getBlocksFn: func(_ context.Context, from, to uint64) ([]*client.Block, error) {
						// Sanity check
						if from > to {
							t.Fatal("invalid block number requested")
						}

						return generateBlocks(t, from, to, tCase.txsPerBlock), nil
					},
					getTxResultsFn: func(_ uint64) ([]*abci.ResponseDeliverTx, error) {
						txs := make([]*abci.ResponseDeliverTx, 0, tCase.txsPerBlock)

						for range tCase.txsPerBlock {
							txs = append(txs, &abci.ResponseDeliverTx{
								ResponseBase: abci.ResponseBase{
									Error: nil,
								},
							})
						}

						return txs, nil
					},
				}
			)

			// Temp file cleanup
			t.Cleanup(func() {
				require.NoError(t, tempFile.Close())
				require.NoError(t, os.Remove(tempFile.Name()))
			})

			// Set the config
			cfg.FromBlock = tCase.fromBlock
			cfg.ToBlock = &tCase.toBlock

			s := NewService(mockClient, standard.NewWriter(tempFile), WithLogger(noop.New()), WithBatchSize(tCase.batchSize))

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
			lineCount := uint64(0)

			// Iterate over each line in the file
			for ; scanner.Scan(); lineCount++ {
				var txData gnoland.TxWithMetadata

				// Unmarshal the JSON data into the Person struct
				if err := amino.UnmarshalJSON(scanner.Bytes(), &txData); err != nil {
					t.Fatalf("unable to unmarshal JSON line, %v", err)
				}

				expectedBlock := tCase.fromBlock + lineCount/tCase.txsPerBlock
				expectedTx := lineCount % tCase.txsPerBlock
				expectedTxMemo := generateMemo(expectedBlock, expectedTx)
				assert.Equal(t, expectedTxMemo, txData.Tx.Memo)
				assert.Equal(
					t,
					blockTime.Add(time.Duration(expectedBlock)*time.Minute).Local(),
					time.UnixMilli(txData.Metadata.Timestamp),
				)
			}

			// Check for errors during scanning
			if err := scanner.Err(); err != nil {
				t.Fatalf("error encountered during scan, %v", err)
			}

			// Ensure we found 1 line by expected transaction
			expectedTxCount := (tCase.toBlock - tCase.fromBlock + 1) * tCase.txsPerBlock
			assert.Equal(t, expectedTxCount, lineCount)
		})
	}

	for _, tCase := range testCases {
		testFunc(t, tCase)
	}
}

func TestBackup_ExecuteBackup_Watch(t *testing.T) {
	t.Parallel()

	//nolint:thelper,gocritic
	testFunc := func(t *testing.T, tCase testCase) {
		t.Run(tCase.name, func(t *testing.T) {
			t.Parallel()

			// Set up the context that is controlled by the test
			ctx, cancelFn := context.WithCancel(context.Background())
			defer cancelFn()

			var (
				tempFile   = createTempFile(t)
				cfg        = DefaultConfig()
				watchStart = tCase.toBlock / 2
				latest     = watchStart

				mockClient = &mockClient{
					getLatestBlockNumberFn: func() (uint64, error) {
						if latest == tCase.toBlock { // Simulate last block incrementing while in watch mode
							cancelFn()
						} else {
							latest++
						}

						return latest, nil
					},
					getBlocksFn: func(_ context.Context, from, to uint64) ([]*client.Block, error) {
						// Sanity check
						if from > to {
							t.Fatal("invalid block number requested")
						}

						switch {
						case from > tCase.toBlock:
							return nil, nil
						case from > watchStart: // Watch mode, return blocks 1 by 1
							return generateBlocks(t, from, from, tCase.txsPerBlock), nil
						case to > latest:
							return generateBlocks(t, from, latest, tCase.txsPerBlock), nil
						}

						return generateBlocks(t, from, to, tCase.txsPerBlock), nil
					},
					getTxResultsFn: func(_ uint64) ([]*abci.ResponseDeliverTx, error) {
						txs := make([]*abci.ResponseDeliverTx, 0, tCase.txsPerBlock)

						for range tCase.txsPerBlock {
							txs = append(txs, &abci.ResponseDeliverTx{
								ResponseBase: abci.ResponseBase{
									Error: nil,
								},
							})
						}

						return txs, nil
					},
				}
			)

			// Temp file cleanup
			t.Cleanup(func() {
				require.NoError(t, tempFile.Close())
				require.NoError(t, os.Remove(tempFile.Name()))
			})

			// Set the config
			cfg.FromBlock = tCase.fromBlock
			cfg.ToBlock = &tCase.toBlock
			cfg.Watch = true

			s := NewService(mockClient, standard.NewWriter(tempFile), WithLogger(noop.New()), WithBatchSize(tCase.batchSize))
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
			lineCount := uint64(0)

			// Iterate over each line in the file
			for ; scanner.Scan(); lineCount++ {
				var txData gnoland.TxWithMetadata

				// Unmarshal the JSON data into the Person struct
				if err := amino.UnmarshalJSON(scanner.Bytes(), &txData); err != nil {
					t.Fatalf("unable to unmarshal JSON line, %v", err)
				}

				expectedBlock := tCase.fromBlock + lineCount/tCase.txsPerBlock
				expectedTx := lineCount % tCase.txsPerBlock
				expectedTxMemo := generateMemo(expectedBlock, expectedTx)
				assert.Equal(t, expectedTxMemo, txData.Tx.Memo)
				assert.Equal(
					t,
					blockTime.Add(time.Duration(expectedBlock)*time.Minute).Local(),
					time.UnixMilli(txData.Metadata.Timestamp),
				)
			}

			// Check for errors during scanning
			if err := scanner.Err(); err != nil {
				t.Fatalf("error encountered during scan, %v", err)
			}

			// Ensure we found 1 line by expected transaction
			expectedTxCount := (tCase.toBlock - tCase.fromBlock + 1) * tCase.txsPerBlock
			assert.Equal(t, expectedTxCount, lineCount)
		})
	}

	for _, tCase := range testCases {
		testFunc(t, tCase)
	}
}
