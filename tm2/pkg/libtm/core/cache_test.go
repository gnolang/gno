package core

import (
	"testing"

	"github.com/gnolang/libtm/messages/types"
	"github.com/stretchr/testify/assert"
)

func TestMessageCache_AddMessages(t *testing.T) {
	t.Parallel()

	isValidFn := func(_ *types.PrevoteMessage) bool {
		return true
	}

	t.Run("non-duplicate messages", func(t *testing.T) {
		t.Parallel()

		// Create the cache
		cache := newMessageCache[*types.PrevoteMessage](isValidFn)

		// Generate non-duplicate messages
		messages := generatePrevoteMessages(t, 10, &types.View{}, nil)

		// Add the messages
		cache.addMessages(messages)

		// Make sure all messages are added
		fetchedMessages := cache.getMessages()

		for index, message := range messages {
			assert.True(t, message.Equals(fetchedMessages[index]))
		}
	})

	t.Run("duplicate messages", func(t *testing.T) {
		t.Parallel()

		var (
			numMessages   = 10
			numDuplicates = numMessages / 2
		)

		// Create the cache
		cache := newMessageCache[*types.PrevoteMessage](isValidFn)

		// Generate non-duplicate messages
		messages := generatePrevoteMessages(t, numMessages, &types.View{}, nil)

		// Make sure some are duplicated
		for i := 0; i < numDuplicates; i++ {
			messages[i].Sender = []byte("common sender")
		}

		expectedMessages := messages[numDuplicates-1:]

		// Add the messages
		cache.addMessages(messages)

		// Make sure all messages are added
		fetchedMessages := cache.getMessages()

		assert.Len(t, fetchedMessages, len(expectedMessages))

		for index, message := range expectedMessages {
			assert.True(t, message.Equals(fetchedMessages[index]))
		}
	})
}
