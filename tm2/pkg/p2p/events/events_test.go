package events

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateEvents generates p2p events
func generateEvents(count int) []Event {
	events := make([]Event, 0, count)

	for i := range count {
		var event Event

		if i%2 == 0 {
			event = PeerConnectedEvent{
				PeerID: types.ID(fmt.Sprintf("peer-%d", i)),
			}
		} else {
			event = PeerDisconnectedEvent{
				PeerID: types.ID(fmt.Sprintf("peer-%d", i)),
			}
		}

		events = append(events, event)
	}

	return events
}

func TestEvents_Subscribe(t *testing.T) {
	t.Parallel()

	var (
		capturedEvents []Event

		events = generateEvents(10)
		subFn  = func(e Event) bool {
			return e.Type() == PeerDisconnected
		}
	)

	// Create the events manager
	e := New()

	// Subscribe to events
	ch, unsubFn := e.Subscribe(subFn)
	defer unsubFn()

	// Listen for the events
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		timeout := time.After(5 * time.Second)

		for {
			select {
			case ev := <-ch:
				capturedEvents = append(capturedEvents, ev)

				if len(capturedEvents) == len(events)/2 {
					return
				}
			case <-timeout:
				return
			}
		}
	}()

	// Send out the events
	for _, ev := range events {
		e.Notify(ev)
	}

	wg.Wait()

	// Make sure the events were captured
	// and filtered properly
	require.Len(t, capturedEvents, len(events)/2)

	for _, ev := range capturedEvents {
		assert.Equal(t, ev.Type(), PeerDisconnected)
	}
}
