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

		tm = Timeout{
			Initial: initial,
			Delta:   delta,
		}
	)

	for round := uint64(0); round < 100; round++ {
		assert.Equal(
			t,
			initial+time.Duration(round)*delta,
			tm.CalculateTimeout(round),
		)
	}
}

func TestTimeout_ScheduleTimeoutPropose(t *testing.T) {
	t.Parallel()

	var (
		capturedMessage *types.PrevoteMessage
		id              = []byte("node ID")
		signature       = []byte("signature")
		view            = &types.View{
			Height: 10,
			Round:  0,
		}

		mockBroadcast = &mockBroadcast{
			broadcastPrevoteFn: func(message *types.PrevoteMessage) {
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
		timeouts:  make(map[step]Timeout),
		broadcast: mockBroadcast,
		node:      mockNode,
		signer:    mockSigner,
	}

	// set the timeout data for the propose step
	tm.timeouts[propose] = Timeout{
		Initial: 50 * time.Millisecond,
		Delta:   50 * time.Millisecond,
	}

	// Schedule the timeout
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	var (
		callback = func() {
			tm.onTimeoutPropose(tm.state.view.Round)
		}

		timeoutPropose = tm.timeouts[propose].CalculateTimeout(tm.state.view.Round)
	)

	tm.scheduleTimeout(ctx, timeoutPropose, callback)

	// Wait for the timer to trigger
	tm.wg.Wait()

	// Validate the prevote message was sent with a NIL value
	require.NotNil(t, capturedMessage)

	assert.Equal(t, signature, capturedMessage.Signature)

	assert.Nil(t, capturedMessage.GetIdentifier())
	assert.Equal(t, id, capturedMessage.GetSender())
	assert.Equal(t, view.GetHeight(), capturedMessage.GetView().GetHeight())
	assert.Equal(t, view.GetRound(), capturedMessage.GetView().GetRound())
}
