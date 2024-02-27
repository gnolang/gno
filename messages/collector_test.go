package messages

import (
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/go-tendermint/messages/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateMessages generates dummy messages
// for the given view and type
func generateMessages(
	t *testing.T,
	count int,
	view *types.View,
	messageTypes ...types.MessageType,
) []*types.Message {
	t.Helper()

	messages := make([]*types.Message, 0, count)

	for index := 0; index < count; index++ {
		for _, messageType := range messageTypes {
			message := &types.Message{
				Type: messageType,
			}

			switch messageType {
			case types.MessageType_PROPOSAL:
				message.Payload = &types.Message_ProposalMessage{
					ProposalMessage: &types.ProposalMessage{
						From: []byte(strconv.Itoa(index)),
						View: view,
					},
				}
			case types.MessageType_PREVOTE:
				message.Payload = &types.Message_PrevoteMessage{
					PrevoteMessage: &types.PrevoteMessage{
						From: []byte(strconv.Itoa(index)),
						View: view,
					},
				}
			case types.MessageType_PRECOMMIT:
				message.Payload = &types.Message_PrecommitMessage{
					PrecommitMessage: &types.PrecommitMessage{
						From: []byte(strconv.Itoa(index)),
						View: view,
					},
				}
			}

			messages = append(messages, message)
		}
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

		generatedMessages := generateMessages(
			t,
			count,
			initialView,
			types.MessageType_PROPOSAL,
		)

		expectedMessages := make([]*types.ProposalMessage, 0, count)

		for _, message := range generatedMessages {
			proposal, ok := message.Payload.(*types.Message_ProposalMessage)
			require.True(t, ok)

			c.AddMessage(proposal.ProposalMessage.View, proposal.ProposalMessage.From, proposal.ProposalMessage)

			expectedMessages = append(expectedMessages, proposal.ProposalMessage)
		}

		// Sort the messages for the test
		sort.SliceStable(expectedMessages, func(i, j int) bool {
			return string(expectedMessages[i].From) < string(expectedMessages[j].From)
		})

		// Get the messages from the store
		messages := c.GetMessages()

		// Sort the messages for the test
		sort.SliceStable(messages, func(i, j int) bool {
			return string(messages[i].From) < string(messages[j].From)
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

		generatedMessages := generateMessages(
			t,
			count,
			initialView,
			types.MessageType_PREVOTE,
		)

		expectedMessages := make([]*types.PrevoteMessage, 0, count)

		for _, message := range generatedMessages {
			prevote, ok := message.Payload.(*types.Message_PrevoteMessage)
			require.True(t, ok)

			c.AddMessage(prevote.PrevoteMessage.View, prevote.PrevoteMessage.From, prevote.PrevoteMessage)

			expectedMessages = append(expectedMessages, prevote.PrevoteMessage)
		}

		// Sort the messages for the test
		sort.SliceStable(expectedMessages, func(i, j int) bool {
			return string(expectedMessages[i].From) < string(expectedMessages[j].From)
		})

		// Get the messages from the store
		messages := c.GetMessages()

		// Sort the messages for the test
		sort.SliceStable(messages, func(i, j int) bool {
			return string(messages[i].From) < string(messages[j].From)
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

		generatedMessages := generateMessages(
			t,
			count,
			initialView,
			types.MessageType_PRECOMMIT,
		)

		expectedMessages := make([]*types.PrecommitMessage, 0, count)

		for _, message := range generatedMessages {
			precommit, ok := message.Payload.(*types.Message_PrecommitMessage)
			require.True(t, ok)

			c.AddMessage(precommit.PrecommitMessage.View, precommit.PrecommitMessage.From, precommit.PrecommitMessage)

			expectedMessages = append(expectedMessages, precommit.PrecommitMessage)
		}

		// Sort the messages for the test
		sort.SliceStable(expectedMessages, func(i, j int) bool {
			return string(expectedMessages[i].From) < string(expectedMessages[j].From)
		})

		// Get the messages from the store
		messages := c.GetMessages()

		// Sort the messages for the test
		sort.SliceStable(messages, func(i, j int) bool {
			return string(messages[i].From) < string(messages[j].From)
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
		commonType   = types.MessageType_PREVOTE
		view         = &types.View{
			Height: 1,
			Round:  1,
		}
	)

	// Create the collector
	c := NewCollector[types.PrevoteMessage]()

	generatedMessages := generateMessages(
		t,
		count,
		view,
		commonType,
	)

	for _, message := range generatedMessages {
		prevote, ok := message.Payload.(*types.Message_PrevoteMessage)
		require.True(t, ok)

		// Make sure each message is from the same sender
		prevote.PrevoteMessage.From = commonSender

		c.AddMessage(prevote.PrevoteMessage.View, prevote.PrevoteMessage.From, prevote.PrevoteMessage)
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

		generatedMessages := generateMessages(
			t,
			count,
			view,
			types.MessageType_PREVOTE,
		)

		expectedMessages := make([]*types.PrevoteMessage, 0, count)

		for _, message := range generatedMessages {
			prevote, ok := message.Payload.(*types.Message_PrevoteMessage)
			require.True(t, ok)

			c.AddMessage(prevote.PrevoteMessage.View, prevote.PrevoteMessage.From, prevote.PrevoteMessage)

			expectedMessages = append(expectedMessages, prevote.PrevoteMessage)
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
			return string(expectedMessages[i].From) < string(expectedMessages[j].From)
		})

		// Sort the messages for the test
		sort.SliceStable(messages, func(i, j int) bool {
			return string(messages[i].From) < string(messages[j].From)
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

		generatedMessages := generateMessages(
			t,
			count,
			view,
			types.MessageType_PREVOTE,
		)

		expectedMessages := make([]*types.PrevoteMessage, 0, count)

		// Create a subscription
		notifyCh, unsubscribeFn := c.Subscribe()
		defer unsubscribeFn()

		for _, message := range generatedMessages {
			prevote, ok := message.Payload.(*types.Message_PrevoteMessage)
			require.True(t, ok)

			c.AddMessage(prevote.PrevoteMessage.View, prevote.PrevoteMessage.From, prevote.PrevoteMessage)

			expectedMessages = append(expectedMessages, prevote.PrevoteMessage)
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
			return string(expectedMessages[i].From) < string(expectedMessages[j].From)
		})

		// Sort the messages for the test
		sort.SliceStable(messages, func(i, j int) bool {
			return string(messages[i].From) < string(messages[j].From)
		})

		// Make sure the messages match
		assert.Equal(t, expectedMessages, messages)
	})
}
