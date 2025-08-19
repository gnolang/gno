package events

import (
	"log"
	"reflect"
	"time"
)

// Returns a synchronous event emitter.
// You can specify initialTimeout to set the initial timeout for waiting for the event (default is 10s).
func Subscribe(evsw EventSwitch, listenerID string, initialTimeout ...time.Duration) <-chan Event {
	ch := make(chan Event) // synchronous
	return SubscribeOn(evsw, listenerID, ch, initialTimeout...)
}

// Like Subscribe, but lets the caller construct a channel.  If the capacity of
// the provided channel is 0, it will be called synchronously; otherwise, it
// will drop when the capacity is reached and a select doesn't immediately
// send.
// You can specify initialTimeout to set the initial timeout for waiting for the event (default is 10s).
func SubscribeOn(evsw EventSwitch, listenerID string, ch chan Event, initialTimeout ...time.Duration) <-chan Event {
	return SubscribeFilteredOn(evsw, listenerID, nil, ch, initialTimeout...)
}

func SubscribeToEvent(evsw EventSwitch, listenerID string, protoevent Event, initialTimeout ...time.Duration) <-chan Event {
	ch := make(chan Event) // synchronous
	return SubscribeToEventOn(evsw, listenerID, protoevent, ch, initialTimeout...)
}

func SubscribeToEventOn(evsw EventSwitch, listenerID string, protoevent Event, ch chan Event, initialTimeout ...time.Duration) <-chan Event {
	rt := reflect.TypeOf(protoevent)
	return SubscribeFilteredOn(evsw, listenerID, func(event Event) bool {
		return reflect.TypeOf(event) == rt
	}, ch, initialTimeout...)
}

type EventFilter func(Event) bool

func SubscribeFiltered(evsw EventSwitch, listenerID string, filter EventFilter, initialTimeout ...time.Duration) <-chan Event {
	ch := make(chan Event)
	return SubscribeFilteredOn(evsw, listenerID, filter, ch, initialTimeout...)
}

func SubscribeFilteredOn(evsw EventSwitch, listenerID string, filter EventFilter, ch chan Event, initialTimeout ...time.Duration) <-chan Event {
	evsw.AddListener(listenerID, func(event Event) {
		if filter != nil && !filter(event) {
			return // filter
		}

		timeout := 10 * time.Second
		if len(initialTimeout) > 0 && initialTimeout[0] > 0 {
			timeout = initialTimeout[0]
		}

		// NOTE: This callback must not block for performance.
		if cap(ch) == 0 {
		LOOP:
			for {
				select { // sync
				case ch <- event:
					break LOOP
				case <-evsw.Quit():
					close(ch)
					break LOOP
				case <-time.After(timeout):
					// After a minute, print a message for debugging.
					log.Printf("[WARN] EventSwitch subscriber %v blocked on %v for %v", listenerID, event, timeout)
					// Exponentially back off warning messages.
					timeout *= 2
				}
			}
		} else {
			select {
			case ch <- event:
			default: // async
				evsw.RemoveListener(listenerID) // TODO log
				close(ch)
			}
		}
	})
	return ch
}
