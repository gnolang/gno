package messages

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/gnolang/libtm/messages/types"
)

// msgType is the combined message type interface,
// for easy reference and type safety
type msgType interface {
	types.ProposalMessage | types.PrevoteMessage | types.PrecommitMessage
}

// this is because Go doesn't support covariance on slices
// []*T -> []I does not work
func ConvertToInterface[T msgType](msgs []*T, convertFunc func(m *T)) {
	for _, msg := range msgs {
		convertFunc(msg)
	}
}

type (
	// collection are the actual received messages.
	// Maps a unique identifier -> their message (of a specific type) to avoid duplicates.
	// Identifiers are derived from <sender ID, height, round>.
	// Each validator in the consensus needs to send at most 1 message of every type
	// (minus the PROPOSAL, which is only sent by the proposer),
	// so the message system needs to keep track of only 1 message per type, per validator, per view
	collection[T msgType] map[string]*T
)

// Collector is a single message type collector
type Collector[T msgType] struct {
	collection    collection[T]    // the message storage
	subscriptions subscriptions[T] // the active message subscriptions

	collectionMux    sync.RWMutex
	subscriptionsMux sync.RWMutex
}

// NewCollector creates a new message collector
func NewCollector[T msgType]() *Collector[T] {
	return &Collector[T]{
		collection:    make(collection[T]),
		subscriptions: make(subscriptions[T]),
	}
}

// Subscribe creates a new collector subscription.
// Returns the channel for receiving messages,
// as well as the unsubscribe method
func (c *Collector[T]) Subscribe() (<-chan func() []*T, func()) {
	c.subscriptionsMux.Lock()
	defer c.subscriptionsMux.Unlock()

	// Create a new subscription
	id, ch := c.subscriptions.add()

	// Create the unsubscribe callback
	unsubscribeFn := func() {
		c.subscriptionsMux.Lock()
		defer c.subscriptionsMux.Unlock()

		c.subscriptions.remove(id)
	}

	// Notify the subscription immediately,
	// since there can be existing messages in the collection.
	// This action assumes the channel is not blocking (created with initial size),
	// since the calling context does not have access to it yet at this point
	notifySubscription(ch, c.GetMessages)

	return ch, unsubscribeFn
}

// GetMessages returns the currently present messages in the collector
func (c *Collector[T]) GetMessages() []*T {
	c.collectionMux.RLock()
	defer c.collectionMux.RUnlock()

	// Fetch the messages in the collection
	return c.collection.getMessages()
}

// getMessages fetches the messages in the collection
func (c *collection[T]) getMessages() []*T {
	messages := make([]*T, 0, len(*c))

	for _, senderMessage := range *c {
		messages = append(messages, senderMessage)
	}

	return messages
}

// DropMessages drops all messages from the collection if their view
// is less than the given view (earlier)
func (c *Collector[T]) DropMessages(view *types.View) {
	c.collectionMux.RLock()
	defer c.collectionMux.RUnlock()

	// Filter out messages from the collection
	shouldStay := func(key string) bool {
		// Construct the view from the message
		messageView := getViewFromKey(key)

		// Only messages who are >= the current view stay
		return messageView.Height >= view.Height &&
			messageView.Round >= view.Round
	}

	// Filter out the messages
	c.collection.dropMessages(shouldStay)
}

// dropMessages drops messages from the collection using the given filter
func (c *collection[T]) dropMessages(shouldStay func(string) bool) {
	for key := range *c {
		if shouldStay(key) {
			continue
		}

		delete(*c, key)
	}
}

// AddMessage adds a new message to the collector
func (c *Collector[T]) AddMessage(view *types.View, from []byte, message *T) {
	c.collectionMux.Lock()

	// Add the message
	c.collection.addMessage(
		getCollectionKey(from, view),
		message,
	)

	c.collectionMux.Unlock()

	// Notify the subscriptions
	c.subscriptionsMux.RLock()
	defer c.subscriptionsMux.RUnlock()

	c.subscriptions.notify(c.GetMessages)
}

// addMessage adds a new message to the collection
func (c *collection[T]) addMessage(key string, message *T) {
	(*c)[key] = message
}

// getCollectionKey constructs a key based on the
// message sender and view information.
// This key guarantees uniqueness in the message store
func getCollectionKey(from []byte, view *types.View) string {
	return fmt.Sprintf("%s_%d_%d", from, view.Height, view.Round)
}

// getViewFromKey constructs the view information,
// based on the collection key
func getViewFromKey(key string) *types.View {
	// Split the key
	keyParts := strings.Split(key, "_")

	// Parse the view values
	height, _ := strconv.ParseUint(keyParts[1], 10, 64) //nolint:errcheck // Key is valid
	round, _ := strconv.ParseUint(keyParts[2], 10, 64)  //nolint:errcheck // Key is valid

	return &types.View{
		Height: height,
		Round:  round,
	}
}
