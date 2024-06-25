package messages

import (
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/libtm/messages/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateProposalMessages generates dummy proposal messages
// for the given view and type
func generateProposalMessages(
	t *testing.T,
	count int,
	view *types.View,
) []*types.ProposalMessage {
	t.Helper()

	messages := make([]*types.ProposalMessage, 0, count)

	for index := 0; index < count; index++ {
		message := &types.ProposalMessage{
			Sender: []byte(strconv.Itoa(index)),
			View:   view,
		}

		messages = append(messages, message)
	}

	return messages
}

// generatePrevoteMessages generates dummy prevote messages
// for the given view and type
func generatePrevoteMessages(
	t *testing.T,
	count int,
	view *types.View,
) []*types.PrevoteMessage {
	t.Helper()

	messages := make([]*types.PrevoteMessage, 0, count)

	for index := 0; index < count; index++ {
		message := &types.PrevoteMessage{
			Sender: []byte(strconv.Itoa(index)),
			View:   view,
		}

		messages = append(messages, message)
	}

	return messages
}

// generatePrevoteMessages generates dummy prevote messages
// for the given view and type
func generatePrecommitMessages(
	t *testing.T,
	count int,
	view *types.View,
) []*types.PrecommitMessage {
	t.Helper()

	messages := make([]*types.PrecommitMessage, 0, count)

	for index := 0; index < count; index++ {
		message := &types.PrecommitMessage{
			Sender: []byte(strconv.Itoa(index)),
			View:   view,
		}

		messages = append(messages, message)
	}

	return messages
}

func TestCollector_AddMessage(t *testing.T) {
	t.Parallel()

	t.Run("empty message queue", func(t *testing.T) {
		t.Parallel()

		// Create the collector
		c := NewCollector[types.ProposalMessage]()

		// Fetch the messages
		messages := c.GetMessages()

		require.NotNil(t, messages)
		assert.Len(t, messages, 0)
	})

	t.Run("valid PROPOSAL messages fetched", func(t *testing.T) {
		t.Parallel()

		var (
			count       = 5
			initialView = &types.View{
				Height: 1,
				Round:  0,
			}
		)

		// Create the collector
		c := NewCollector[types.ProposalMessage]()

		generatedMessages := generateProposalMessages(
			t,
			count,
			initialView,
		)

		expectedMessages := make([]*types.ProposalMessage, 0, count)

		for _, proposal := range generatedMessages {
			c.AddMessage(proposal.View, proposal.Sender, proposal)

			expectedMessages = append(expectedMessages, proposal)
		}

		// Sort the messages for the test
		sort.SliceStable(expectedMessages, func(i, j int) bool {
			return string(expectedMessages[i].Sender) < string(expectedMessages[j].Sender)
		})

		// Get the messages Sender the store
		messages := c.GetMessages()

		// Sort the messages for the test
		sort.SliceStable(messages, func(i, j int) bool {
			return string(messages[i].Sender) < string(messages[j].Sender)
		})

		// Make sure the messages match
		assert.Equal(t, expectedMessages, messages)
	})

	t.Run("valid PREVOTE messages fetched", func(t *testing.T) {
		t.Parallel()

		var (
			count       = 5
			initialView = &types.View{
				Height: 1,
				Round:  0,
			}
		)

		// Create the collector
		c := NewCollector[types.PrevoteMessage]()

		generatedMessages := generatePrevoteMessages(
			t,
			count,
			initialView,
		)

		expectedMessages := make([]*types.PrevoteMessage, 0, count)

		for _, prevote := range generatedMessages {
			c.AddMessage(prevote.View, prevote.Sender, prevote)

			expectedMessages = append(expectedMessages, prevote)
		}

		// Sort the messages for the test
		sort.SliceStable(expectedMessages, func(i, j int) bool {
			return string(expectedMessages[i].Sender) < string(expectedMessages[j].Sender)
		})

		// Get the messages from the store
		messages := c.GetMessages()

		// Sort the messages for the test
		sort.SliceStable(messages, func(i, j int) bool {
			return string(messages[i].Sender) < string(messages[j].Sender)
		})

		// Make sure the messages match
		assert.Equal(t, expectedMessages, messages)
	})

	t.Run("valid PRECOMMIT messages fetched", func(t *testing.T) {
		t.Parallel()

		var (
			count       = 5
			initialView = &types.View{
				Height: 1,
				Round:  0,
			}
		)

		// Create the collector
		c := NewCollector[types.PrecommitMessage]()

		generatedMessages := generatePrecommitMessages(
			t,
			count,
			initialView,
		)

		expectedMessages := make([]*types.PrecommitMessage, 0, count)

		for _, precommit := range generatedMessages {
			c.AddMessage(precommit.View, precommit.Sender, precommit)

			expectedMessages = append(expectedMessages, precommit)
		}

		// Sort the messages for the test
		sort.SliceStable(expectedMessages, func(i, j int) bool {
			return string(expectedMessages[i].Sender) < string(expectedMessages[j].Sender)
		})

		// Get the messages Sender the store
		messages := c.GetMessages()

		// Sort the messages for the test
		sort.SliceStable(messages, func(i, j int) bool {
			return string(messages[i].Sender) < string(messages[j].Sender)
		})

		// Make sure the messages match
		assert.Equal(t, expectedMessages, messages)
	})
}

func TestCollector_AddDuplicateMessages(t *testing.T) {
	t.Parallel()

	var (
		count        = 5
		commonSender = []byte("sender 1")
		view         = &types.View{
			Height: 1,
			Round:  1,
		}
	)

	// Create the collector
	c := NewCollector[types.PrevoteMessage]()

	generatedMessages := generatePrevoteMessages(
		t,
		count,
		view,
	)

	for _, prevote := range generatedMessages {
		// Make sure each message is from the same sender
		prevote.Sender = commonSender

		c.AddMessage(prevote.View, prevote.Sender, prevote)
	}

	// Check that only 1 message has been added
	assert.Len(t, c.GetMessages(), 1)
}

func TestCollector_Subscribe(t *testing.T) {
	t.Parallel()

	t.Run("subscribe with pre-existing messages", func(t *testing.T) {
		t.Parallel()

		var (
			count = 100
			view  = &types.View{
				Height: 1,
				Round:  0,
			}
		)

		// Create the collector
		c := NewCollector[types.PrevoteMessage]()

		generatedMessages := generatePrevoteMessages(
			t,
			count,
			view,
		)

		expectedMessages := make([]*types.PrevoteMessage, 0, count)

		for _, prevote := range generatedMessages {
			c.AddMessage(prevote.View, prevote.Sender, prevote)

			expectedMessages = append(expectedMessages, prevote)
		}

		// Create a subscription
		notifyCh, unsubscribeFn := c.Subscribe()
		defer unsubscribeFn()

		var messages []*types.PrevoteMessage

		select {
		case callback := <-notifyCh:
			messages = callback()
		case <-time.After(5 * time.Second):
		}

		// Sort the messages for the test
		sort.SliceStable(expectedMessages, func(i, j int) bool {
			return string(expectedMessages[i].Sender) < string(expectedMessages[j].Sender)
		})

		// Sort the messages for the test
		sort.SliceStable(messages, func(i, j int) bool {
			return string(messages[i].Sender) < string(messages[j].Sender)
		})

		// Make sure the messages match
		assert.Equal(t, expectedMessages, messages)
	})

	t.Run("subscribe with no pre-existing messages", func(t *testing.T) {
		t.Parallel()

		var (
			count = 100
			view  = &types.View{
				Height: 1,
				Round:  0,
			}
		)

		// Create the collector
		c := NewCollector[types.PrevoteMessage]()

		generatedMessages := generatePrevoteMessages(
			t,
			count,
			view,
		)

		expectedMessages := make([]*types.PrevoteMessage, 0, count)

		// Create a subscription
		notifyCh, unsubscribeFn := c.Subscribe()
		defer unsubscribeFn()

		for _, prevote := range generatedMessages {
			c.AddMessage(prevote.View, prevote.Sender, prevote)

			expectedMessages = append(expectedMessages, prevote)
		}

		var (
			messages []*types.PrevoteMessage

			wg sync.WaitGroup
		)

		wg.Add(1)

		go func() {
			defer wg.Done()

			select {
			case callback := <-notifyCh:
				messages = callback()
			case <-time.After(5 * time.Second):
			}
		}()

		wg.Wait()

		// Sort the messages for the test
		sort.SliceStable(expectedMessages, func(i, j int) bool {
			return string(expectedMessages[i].Sender) < string(expectedMessages[j].Sender)
		})

		// Sort the messages for the test
		sort.SliceStable(messages, func(i, j int) bool {
			return string(messages[i].Sender) < string(messages[j].Sender)
		})

		// Make sure the messages match
		assert.Equal(t, expectedMessages, messages)
	})
}

func TestCollector_DropMessages(t *testing.T) {
	t.Parallel()

	var (
		count = 5
		view  = &types.View{
			Height: 10,
			Round:  5,
		}
		earlierView = &types.View{
			Height: view.Height,
			Round:  view.Round - 1,
		}
	)

	// Create the collector
	c := NewCollector[types.PrevoteMessage]()

	// Generate latest round messages
	latestRoundMessages := generatePrevoteMessages(
		t,
		count,
		view,
	)

	// Generate earlier round messages
	earlierRoundMessages := generatePrevoteMessages(
		t,
		count,
		earlierView,
	)

	for _, message := range latestRoundMessages {
		c.AddMessage(message.GetView(), message.GetSender(), message)
	}

	for _, message := range earlierRoundMessages {
		c.AddMessage(message.GetView(), message.GetSender(), message)
	}

	// Drop the older messages
	c.DropMessages(view)

	// Make sure the messages were dropped
	fetchedMessages := c.GetMessages()

	require.Len(t, fetchedMessages, len(latestRoundMessages))
	assert.ElementsMatch(t, fetchedMessages, latestRoundMessages)
}
