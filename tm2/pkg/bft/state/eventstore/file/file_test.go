package file

import (
	"bufio"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	storetypes "github.com/gnolang/gno/tm2/pkg/bft/state/eventstore/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

// generateTestTransactions generates random transaction results
func generateTestTransactions(count int) []types.TxResult {
	txs := make([]types.TxResult, count)

	for i := range count {
		txs[i] = types.TxResult{}
	}

	return txs
}

func TestTxEventStore_New(t *testing.T) {
	t.Parallel()

	t.Run("invalid file path specified", func(t *testing.T) {
		t.Parallel()

		cfg := &storetypes.Config{
			EventStoreType: "invalid",
		}

		i, err := NewTxEventStore(cfg)

		assert.Nil(t, i)
		assert.ErrorIs(t, err, errInvalidType)
	})

	t.Run("invalid file path specified", func(t *testing.T) {
		t.Parallel()

		cfg := &storetypes.Config{
			EventStoreType: EventStoreType,
			Params:         nil,
		}

		i, err := NewTxEventStore(cfg)

		assert.Nil(t, i)
		assert.ErrorIs(t, err, errMissingPath)
	})

	t.Run("valid file path specified", func(t *testing.T) {
		t.Parallel()

		headPath := "."

		cfg := &storetypes.Config{
			EventStoreType: EventStoreType,
			Params: map[string]any{
				Path: headPath,
			},
		}

		i, err := NewTxEventStore(cfg)
		if i == nil {
			t.Fatalf("unable to create event store")
		}

		assert.NoError(t, err)
		assert.Equal(t, headPath, i.headPath)
		assert.Equal(t, EventStoreType, i.GetType())
	})
}

func TestTxEventStore_Append(t *testing.T) {
	t.Parallel()

	headFile, cleanup := testutils.NewTestFile(t)
	t.Cleanup(func() {
		cleanup()
	})

	eventStore, err := NewTxEventStore(&storetypes.Config{
		EventStoreType: EventStoreType,
		Params: map[string]any{
			Path: headFile.Name(),
		},
	})
	if err != nil {
		t.Fatalf("unable to create tx event store, %v", err)
	}

	// Start the event store
	if err = eventStore.Start(); err != nil {
		t.Fatalf("unable to start event store, %v", err)
	}

	t.Cleanup(func() {
		// Stop the event store
		if err = eventStore.Stop(); err != nil {
			t.Fatalf("unable to stop event store gracefully, %v", err)
		}
	})

	numTxs := 10
	txs := generateTestTransactions(numTxs)

	for _, tx := range txs {
		if err = eventStore.Append(tx); err != nil {
			t.Fatalf("unable to store transaction, %v", err)
		}
	}

	// Make sure the file group's size is valid
	if eventStore.group.ReadGroupInfo().TotalSize == 0 {
		t.Fatalf("invalid group size")
	}

	// Open file for reading
	scanner := bufio.NewScanner(headFile)

	linesRead := 0
	for scanner.Scan() {
		line := scanner.Bytes()

		var txRes types.TxResult
		if err = amino.UnmarshalJSON(line, &txRes); err != nil {
			t.Fatalf("unable to read store line")
		}

		assert.Equal(t, txs[linesRead], txRes)

		linesRead++
	}

	assert.Equal(t, numTxs, linesRead)
}
