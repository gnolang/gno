package messages

import "github.com/rs/xid"

type (
	// MsgCallback is the callback that returns all given messages
	MsgCallback[T msgType] func() []*T

	// subscriptions is the subscription store,
	// maps subscription id -> notification channel.
	// Usage of this type is NOT thread safe
	subscriptions[T msgType] map[string]chan func() []*T
)

// add adds a new subscription to the subscription map.
// Returns the subscription ID, and update channel
func (s *subscriptions[T]) add() (string, chan func() []*T) {
	var (
		id = xid.New().String()
		ch = make(chan func() []*T, 1)
	)

	(*s)[id] = ch

	return id, ch
}

// remove removes the given subscription
func (s *subscriptions[T]) remove(id string) {
	if ch := (*s)[id]; ch != nil {
		// Close the notification channel
		close(ch)
	}

	// Delete the subscription
	delete(*s, id)
}

// notify notifies all subscription listeners
func (s *subscriptions[T]) notify(callback func() []*T) {
	// Notify the listeners
	for _, ch := range *s {
		notifySubscription(ch, callback)
	}
}

// notifySubscription alerts the notification channel
// about a callback. This function is pure syntactic sugar
func notifySubscription[T msgType](
	ch chan func() []*T,
	callback MsgCallback[T],
) {
	select {
	case ch <- callback:
	default:
	}
}
