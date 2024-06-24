package core

import (
	"github.com/gnolang/libtm/messages/types"
)

// msgType is the combined message type interface
type msgType interface {
	*types.ProposalMessage | *types.PrevoteMessage | *types.PrecommitMessage
}

type cacheMessage interface {
	msgType
	Message
}

// messageCache contains filtered messages
// added in by the calling context
type messageCache[T cacheMessage] struct {
	isValidFn        func(T) bool
	seenMap          map[string]struct{}
	filteredMessages []T
}

// newMessageCache creates a new incoming message cache
func newMessageCache[T cacheMessage](isValidFn func(T) bool) messageCache[T] {
	return messageCache[T]{
		isValidFn:        isValidFn,
		filteredMessages: make([]T, 0),
		seenMap:          make(map[string]struct{}),
	}
}

// addMessages pushes a new message list that is filtered
// and parsed by the cache
func (c *messageCache[T]) addMessages(messages []T) {
	for _, message := range messages {
		sender := message.GetSender()

		// Check if the message has been seen in the past
		_, seen := c.seenMap[string(sender)]
		if seen {
			continue
		}

		// Filter the message
		if !c.isValidFn(message) {
			continue
		}

		// Mark the message as seen
		c.seenMap[string(sender)] = struct{}{}

		// Save the message as it's
		// been filtered, and doesn't exist in the cache
		c.filteredMessages = append(c.filteredMessages, message)
	}
}

// getMessages returns the filtered out messages from the cache
func (c *messageCache[T]) getMessages() []T {
	return c.filteredMessages
}
