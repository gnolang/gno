package file

import (
	"bufio"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/state/txindex/config"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

// generateTestTransactions generates random transaction results
func generateTestTransactions(count int) []types.TxResult {
	txs := make([]types.TxResult, count)

	for i := 0; i < count; i++ {
		txs[i] = types.TxResult{}
	}

	return txs
}

func TestTxIndexer_New(t *testing.T) {
	t.Parallel()

	t.Run("invalid file path specified", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			IndexerType: "invalid",
		}

		i, err := NewTxIndexer(cfg)

		assert.Nil(t, i)
		assert.ErrorIs(t, err, errInvalidType)
	})

	t.Run("invalid file path specified", func(t *testing.T) {
		t.Parallel()

		cfg := &config.Config{
			IndexerType: IndexerType,
			Params:      nil,
		}

		i, err := NewTxIndexer(cfg)

		assert.Nil(t, i)
		assert.ErrorIs(t, err, errMissingPath)
	})

	t.Run("valid file path specified", func(t *testing.T) {
		t.Parallel()

		headPath := "."

		cfg := &config.Config{
			IndexerType: IndexerType,
			Params: map[string]any{
				Path: headPath,
			},
		}

		i, err := NewTxIndexer(cfg)
		if i == nil {
			t.Fatalf("unable to create indexer")
		}

		assert.NoError(t, err)
		assert.Equal(t, headPath, i.headPath)
		assert.Equal(t, IndexerType, i.GetType())
	})
}

func TestTxIndexer_Index(t *testing.T) {
	t.Parallel()

	headFile, cleanup := testutils.NewTestFile(t)
	t.Cleanup(func() {
		cleanup()
	})

	indexer, err := NewTxIndexer(&config.Config{
		IndexerType: IndexerType,
		Params: map[string]any{
			Path: headFile.Name(),
		},
	})

	if err != nil {
		t.Fatalf("unable to create tx indexer, %v", err)
	}

	// Start the indexer
	if err = indexer.Start(); err != nil {
		t.Fatalf("unable to start indexer, %v", err)
	}

	t.Cleanup(func() {
		// Stop the indexer
		if err = indexer.Stop(); err != nil {
			t.Fatalf("unable to stop indexer gracefully, %v", err)
		}
	})

	numTxs := 10
	txs := generateTestTransactions(numTxs)

	for _, tx := range txs {
		if err = indexer.Index(tx); err != nil {
			t.Fatalf("unable to index transaction, %v", err)
		}
	}

	// Make sure the file group's size is valid
	if indexer.group.ReadGroupInfo().TotalSize == 0 {
		t.Fatalf("invalid group size")
	}

	// Open file for reading
	scanner := bufio.NewScanner(headFile)

	linesRead := 0
	for scanner.Scan() {
		line := scanner.Bytes()

		var txRes types.TxResult
		if err = amino.UnmarshalJSON(line, &txRes); err != nil {
			t.Fatalf("unable to read indexer line")
		}

		assert.Equal(t, txs[linesRead], txRes)

		linesRead++
	}

	assert.Equal(t, numTxs, linesRead)
}
