package events

import (
	"sync"

	"github.com/rs/xid"
)

// EventFilter is the filter function used to
// filter incoming p2p events. A false flag will
// consider the event as irrelevant
type EventFilter func(Event) bool

// Events is the p2p event switch
type Events struct {
	subs             subscriptions
	subscriptionsMux sync.RWMutex
}

// New creates a new event subscription manager
func New() *Events {
	return &Events{
		subs: make(subscriptions),
	}
}

// Subscribe registers a new filtered event listener
func (es *Events) Subscribe(filterFn EventFilter) (<-chan Event, func()) {
	es.subscriptionsMux.Lock()
	defer es.subscriptionsMux.Unlock()

	// Create a new subscription
	id, ch := es.subs.add(filterFn)

	// Create the unsubscribe callback
	unsubscribeFn := func() {
		es.subscriptionsMux.Lock()
		defer es.subscriptionsMux.Unlock()

		es.subs.remove(id)
	}

	return ch, unsubscribeFn
}

// Notify notifies all subscribers of an incoming event [BLOCKING]
func (es *Events) Notify(event Event) {
	es.subscriptionsMux.RLock()
	defer es.subscriptionsMux.RUnlock()

	es.subs.notify(event)
}

type (
	// subscriptions holds the corresponding subscription information
	subscriptions map[string]subscription // subscription ID -> subscription

	// subscription wraps the subscription notification channel,
	// and the event filter
	subscription struct {
		ch       chan Event
		filterFn EventFilter
	}
)

// add adds a new subscription to the subscription map.
// Returns the subscription ID, and update channel
func (s *subscriptions) add(filterFn EventFilter) (string, chan Event) {
	var (
		id = xid.New().String()
		// Since the event stream is non-blocking,
		// the event buffer should be sufficiently
		// large for most use-cases. Subscribers can
		// handle large event load caller-side to mitigate
		// events potentially being missed
		ch = make(chan Event, 100)
	)

	(*s)[id] = subscription{
		ch:       ch,
		filterFn: filterFn,
	}

	return id, ch
}

// remove removes the given subscription
func (s *subscriptions) remove(id string) {
	if sub, exists := (*s)[id]; exists {
		// Close the notification channel
		close(sub.ch)
	}

	// Delete the subscription
	delete(*s, id)
}

// notify notifies all subscription listeners,
// if their filters pass
func (s *subscriptions) notify(event Event) {
	// Notify the listeners
	for _, sub := range *s {
		if !sub.filterFn(event) {
			continue
		}

		select {
		case sub.ch <- event:
		default: // non-blocking
		}
	}
}
