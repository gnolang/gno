package gnoland

import (
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/rs/xid"
)

// filterFn is the filter method for incoming events
type filterFn[T any] func(events.Event) []T

// collector is the generic in-memory event collector
type collector[T any] struct {
	events []T         // temporary event storage
	filter filterFn[T] // method used for filtering events
}

// newCollector creates a new event collector
func newCollector[T any](
	evsw events.EventSwitch,
	filter filterFn[T],
) *collector[T] {
	c := &collector[T]{
		events: make([]T, 0),
		filter: filter,
	}

	// Register the listener
	evsw.AddListener(xid.New().String(), func(e events.Event) {
		c.updateWith(e)
	})

	return c
}

// updateWith updates the collector with the given event
func (c *collector[T]) updateWith(event events.Event) {
	if extracted := c.filter(event); extracted != nil {
		c.events = append(c.events, extracted...)
	}
}

// getEvents returns the filtered events,
// and resets the collector store
func (c *collector[T]) getEvents() []T {
	capturedEvents := make([]T, len(c.events))
	copy(capturedEvents, c.events)

	c.events = c.events[:0]

	return capturedEvents
}
