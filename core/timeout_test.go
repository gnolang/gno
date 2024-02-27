package core

import (
	"context"
	"testing"
	"time"

	"github.com/gnolang/go-tendermint/messages/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeout_CalculateTimeout(t *testing.T) {
	t.Parallel()

	var (
		initial = 10 * time.Second
		delta   = 200 * time.Millisecond

		timeout = timeout{
			initial: initial,
			delta:   delta,
		}
	)

	for round := uint64(0); round < 100; round++ {
		assert.Equal(
			t,
			initial+time.Duration(round)*delta,
			timeout.calculateTimeout(round),
		)
	}
}

func TestTimeout_ScheduleTimeoutPropose(t *testing.T) {
	t.Parallel()

	var (
		capturedMessage *types.Message
		id              = []byte("node ID")
		signature       = []byte("signature")
		view            = &types.View{
			Height: 10,
			Round:  0,
		}

		mockBroadcast = &mockBroadcast{
			broadcastFn: func(message *types.Message) {
				capturedMessage = message
			},
		}

		mockNode = &mockNode{
			idFn: func() []byte {
				return id
			},
		}

		mockSigner = &mockSigner{
			signFn: func(_ []byte) []byte {
				return signature
			},
		}
	)

	tm := &Tendermint{
		state:     newState(view),
		timeouts:  make(map[step]timeout),
		broadcast: mockBroadcast,
		node:      mockNode,
		signer:    mockSigner,
	}

	// Set the timeout data for the propose step
	tm.timeouts[propose] = timeout{
		initial: 50 * time.Millisecond,
		delta:   50 * time.Millisecond,
	}

	// Schedule the timeout
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	tm.scheduleTimeoutPropose(ctx)

	// Wait for the timer to trigger
	tm.wg.Wait()

	// Validate the prevote message was sent with a NIL value
	require.NotNil(t, capturedMessage)

	require.Equal(t, capturedMessage.Type, types.MessageType_PREVOTE)

	message, ok := capturedMessage.Payload.(*types.Message_PrevoteMessage)
	require.True(t, ok)

	assert.Equal(t, signature, capturedMessage.Signature)

	assert.Nil(t, message.PrevoteMessage.Identifier)
	assert.Equal(t, id, message.PrevoteMessage.From)
	assert.Equal(t, view.GetHeight(), message.PrevoteMessage.View.GetHeight())
	assert.Equal(t, view.GetRound(), message.PrevoteMessage.View.GetRound())
}
