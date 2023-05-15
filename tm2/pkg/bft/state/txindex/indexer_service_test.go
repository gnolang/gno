package txindex

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/stretchr/testify/assert"
)

// generateTxEvents generates random transaction events
func generateTxEvents(count int) []types.EventTx {
	txEvents := make([]types.EventTx, count)

	for i := 0; i < count; i++ {
		txEvents[i] = types.EventTx{
			Result: types.TxResult{},
		}
	}

	return txEvents
}

func TestIndexerService_Monitor(t *testing.T) {
	t.Parallel()

	const defaultTimeout = 5 * time.Second

	var (
		startCalled     = false
		stopCalled      = false
		receivedResults = make([]types.TxResult, 0)
		receivedSize    atomic.Int64

		cb    events.EventCallback
		cbSet atomic.Bool

		mockIndexer = &mockIndexer{
			startFn: func() error {
				startCalled = true

				return nil
			},
			stopFn: func() error {
				stopCalled = true

				return nil
			},
			indexFn: func(result types.TxResult) error {
				receivedResults = append(receivedResults, result)

				// Atomic because we are accessing this size from a routine
				receivedSize.Store(int64(len(receivedResults)))

				return nil
			},
		}
		mockEventSwitch = &mockEventSwitch{
			fireEventFn: func(event events.Event) {
				// Exec the callback on event fire
				cb(event)
			},
			addListenerFn: func(_ string, callback events.EventCallback) {
				// Attach callback
				cb = callback

				// Atomic because we are accessing this info from a routine
				cbSet.Store(true)
			},
		}
	)

	// Create a new indexer instance
	i := NewIndexerService(mockIndexer, mockEventSwitch)
	if i == nil {
		t.Fatal("unable to create indexer service")
	}

	// Start the indexer
	if err := i.OnStart(); err != nil {
		t.Fatalf("unable to start indexer, %v", err)
	}

	assert.True(t, startCalled)

	t.Cleanup(func() {
		// Stop the indexer
		i.OnStop()

		assert.True(t, stopCalled)
	})

	// Fire off the events so the indexer can catch them

	numEvents := 1000
	txEvents := generateTxEvents(numEvents)

	var wg sync.WaitGroup

	// Start a routine that asynchronously pushes events
	wg.Add(1)
	go func() {
		defer wg.Done()

		timeout := time.After(defaultTimeout)

		for {
			select {
			case <-timeout:
				return
			default:
				// If the callback is set, fire the events
				if !cbSet.Load() {
					// Listener not set yet
					continue
				}

				for _, event := range txEvents {
					mockEventSwitch.FireEvent(event)
				}

				return
			}
		}
	}()

	// Start a routine that monitors received results
	wg.Add(1)
	go func() {
		defer wg.Done()

		timeout := time.After(defaultTimeout)

		for {
			select {
			case <-timeout:
				return
			default:
				if int(receivedSize.Load()) == numEvents {
					return
				}
			}
		}
	}()

	wg.Wait()

	// Make sure all results were received
	if len(receivedResults) != numEvents {
		t.Fatalf("invalid number of results received, %d", len(receivedResults))
	}

	// Make sure all results match
	for index, event := range txEvents {
		assert.Equal(t, event.Result, receivedResults[index])
	}
}
