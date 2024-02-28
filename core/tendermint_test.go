package core

import (
	"testing"
	"time"

	"github.com/gnolang/go-tendermint/messages/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTendermint_AddMessage_Invalid(t *testing.T) {
	t.Parallel()

	t.Run("empty message", func(t *testing.T) {
		t.Parallel()

		tm := &Tendermint{
			store: newStore(),
		}

		assert.ErrorIs(
			t,
			tm.AddMessage(nil),
			ErrMessageNotSet,
		)
	})

	t.Run("empty payload", func(t *testing.T) {
		t.Parallel()

		message := &types.Message{
			Payload: nil,
		}

		tm := &Tendermint{
			store: newStore(),
		}

		assert.ErrorIs(
			t,
			tm.AddMessage(message),
			ErrMessagePayloadNotSet,
		)
	})

	t.Run("invalid signature payload", func(t *testing.T) {
		t.Parallel()

		// message that has a type / payload mismatch
		message := &types.Message{
			Type:    types.MessageType_PROPOSAL,
			Payload: &types.Message_PrevoteMessage{},
		}

		tm := &Tendermint{
			store: newStore(),
		}

		assert.ErrorIs(
			t,
			tm.AddMessage(message),
			types.ErrInvalidMessagePayload,
		)
	})

	t.Run("invalid signature", func(t *testing.T) {
		t.Parallel()

		var (
			signature = []byte("invalid signature")
			message   = &types.Message{
				Type: types.MessageType_PREVOTE,
				Payload: &types.Message_PrevoteMessage{
					PrevoteMessage: &types.PrevoteMessage{},
				},
				Signature: signature,
			}

			signer = &mockSigner{
				isValidSignatureFn: func(_ []byte, sig []byte) bool {
					require.Equal(t, signature, sig)

					return false
				},
			}
		)

		tm := &Tendermint{
			store:  newStore(),
			signer: signer,
		}

		assert.ErrorIs(
			t,
			tm.AddMessage(message),
			ErrInvalidMessageSignature,
		)
	})

	t.Run("sender is not a validator", func(t *testing.T) {
		t.Parallel()

		var (
			signature = []byte("valid signature")
			sender    = []byte("sender")

			message = &types.Message{
				Type: types.MessageType_PREVOTE,
				Payload: &types.Message_PrevoteMessage{
					PrevoteMessage: &types.PrevoteMessage{
						View: &types.View{},
						From: sender,
					},
				},
				Signature: signature,
			}

			signer = &mockSigner{
				isValidSignatureFn: func(_ []byte, sig []byte) bool {
					require.Equal(t, signature, sig)

					return true
				},
			}
			verifier = &mockVerifier{
				isValidatorFn: func(from []byte) bool {
					require.Equal(t, sender, from)

					return false
				},
			}
		)

		tm := &Tendermint{
			store:    newStore(),
			signer:   signer,
			verifier: verifier,
		}

		assert.ErrorIs(
			t,
			tm.AddMessage(message),
			ErrMessageFromNonValidator,
		)
	})

	t.Run("message is for an earlier height", func(t *testing.T) {
		t.Parallel()

		var (
			currentView = &types.View{
				Height: 10,
				Round:  0,
			}
			signature = []byte("valid signature")
			sender    = []byte("sender")

			message = &types.Message{
				Type: types.MessageType_PREVOTE,
				Payload: &types.Message_PrevoteMessage{
					PrevoteMessage: &types.PrevoteMessage{
						View: &types.View{
							Height: currentView.Height - 1, // earlier height
							Round:  currentView.Round,
						},
						From: sender,
					},
				},
				Signature: signature,
			}

			signer = &mockSigner{
				isValidSignatureFn: func(_ []byte, sig []byte) bool {
					require.Equal(t, signature, sig)

					return true
				},
			}
			verifier = &mockVerifier{
				isValidatorFn: func(from []byte) bool {
					require.Equal(t, sender, from)

					return true
				},
			}
		)

		tm := &Tendermint{
			state:    newState(currentView),
			store:    newStore(),
			signer:   signer,
			verifier: verifier,
		}

		assert.ErrorIs(
			t,
			tm.AddMessage(message),
			ErrEarlierHeightMessage,
		)
	})

	t.Run("message is for an earlier round", func(t *testing.T) {
		t.Parallel()

		var (
			currentView = &types.View{
				Height: 1,
				Round:  10,
			}
			signature = []byte("valid signature")
			sender    = []byte("sender")

			message = &types.Message{
				Type: types.MessageType_PREVOTE,
				Payload: &types.Message_PrevoteMessage{
					PrevoteMessage: &types.PrevoteMessage{
						View: &types.View{
							Height: currentView.Height,
							Round:  currentView.Round - 1, // earlier round
						},
						From: sender,
					},
				},
				Signature: signature,
			}

			signer = &mockSigner{
				isValidSignatureFn: func(_ []byte, sig []byte) bool {
					require.Equal(t, signature, sig)

					return true
				},
			}
			verifier = &mockVerifier{
				isValidatorFn: func(from []byte) bool {
					require.Equal(t, sender, from)

					return true
				},
			}
		)

		tm := &Tendermint{
			state:    newState(currentView),
			store:    newStore(),
			signer:   signer,
			verifier: verifier,
		}

		assert.ErrorIs(
			t,
			tm.AddMessage(message),
			ErrEarlierRoundMessage,
		)
	})

	t.Run("invalid proposal message payload", func(t *testing.T) {
		t.Parallel()

		var (
			currentView = &types.View{
				Height: 1,
				Round:  10,
			}
			signature = []byte("valid signature")
			sender    = []byte("sender")

			message = &types.Message{
				Type: types.MessageType_PROPOSAL,
				Payload: &types.Message_ProposalMessage{
					ProposalMessage: &types.ProposalMessage{
						View: nil, // invalid view
						From: sender,
					},
				},
				Signature: signature,
			}

			signer = &mockSigner{
				isValidSignatureFn: func(_ []byte, sig []byte) bool {
					require.Equal(t, signature, sig)

					return true
				},
			}
			verifier = &mockVerifier{
				isValidatorFn: func(from []byte) bool {
					require.Equal(t, sender, from)

					return true
				},
			}
		)

		tm := &Tendermint{
			state:    newState(currentView),
			store:    newStore(),
			signer:   signer,
			verifier: verifier,
		}

		assert.ErrorIs(
			t,
			tm.AddMessage(message),
			types.ErrInvalidMessagePayload,
		)
	})

	t.Run("invalid prevote message payload", func(t *testing.T) {
		t.Parallel()

		var (
			currentView = &types.View{
				Height: 1,
				Round:  10,
			}
			signature = []byte("valid signature")
			sender    = []byte("sender")

			message = &types.Message{
				Type: types.MessageType_PREVOTE,
				Payload: &types.Message_PrevoteMessage{
					PrevoteMessage: &types.PrevoteMessage{
						View: nil, // invalid view
						From: sender,
					},
				},
				Signature: signature,
			}

			signer = &mockSigner{
				isValidSignatureFn: func(_ []byte, sig []byte) bool {
					require.Equal(t, signature, sig)

					return true
				},
			}
			verifier = &mockVerifier{
				isValidatorFn: func(from []byte) bool {
					require.Equal(t, sender, from)

					return true
				},
			}
		)

		tm := &Tendermint{
			state:    newState(currentView),
			store:    newStore(),
			signer:   signer,
			verifier: verifier,
		}

		assert.ErrorIs(
			t,
			tm.AddMessage(message),
			types.ErrInvalidMessagePayload,
		)
	})

	t.Run("invalid precommit message payload", func(t *testing.T) {
		t.Parallel()

		var (
			currentView = &types.View{
				Height: 1,
				Round:  10,
			}
			signature = []byte("valid signature")
			sender    = []byte("sender")

			message = &types.Message{
				Type: types.MessageType_PRECOMMIT,
				Payload: &types.Message_PrecommitMessage{
					PrecommitMessage: &types.PrecommitMessage{
						View: nil, // invalid view
						From: sender,
					},
				},
				Signature: signature,
			}

			signer = &mockSigner{
				isValidSignatureFn: func(_ []byte, sig []byte) bool {
					require.Equal(t, signature, sig)

					return true
				},
			}
			verifier = &mockVerifier{
				isValidatorFn: func(from []byte) bool {
					require.Equal(t, sender, from)

					return true
				},
			}
		)

		tm := &Tendermint{
			state:    newState(currentView),
			store:    newStore(),
			signer:   signer,
			verifier: verifier,
		}

		assert.ErrorIs(
			t,
			tm.AddMessage(message),
			types.ErrInvalidMessagePayload,
		)
	})
}

func TestTendermint_AddMessage_Valid(t *testing.T) {
	t.Parallel()

	t.Run("valid proposal message", func(t *testing.T) {
		t.Parallel()

		var (
			currentView = &types.View{
				Height: 1,
				Round:  10,
			}

			signature = []byte("valid signature")
			sender    = []byte("sender")
			proposal  = []byte("proposal")

			message = &types.Message{
				Type: types.MessageType_PROPOSAL,
				Payload: &types.Message_ProposalMessage{
					ProposalMessage: &types.ProposalMessage{
						View:          currentView,
						From:          sender,
						Proposal:      proposal,
						ProposalRound: -1,
					},
				},
				Signature: signature,
			}

			signer = &mockSigner{
				isValidSignatureFn: func(_ []byte, sig []byte) bool {
					require.Equal(t, signature, sig)

					return true
				},
			}
			verifier = &mockVerifier{
				isValidatorFn: func(from []byte) bool {
					require.Equal(t, sender, from)

					return true
				},
			}
		)

		tm := &Tendermint{
			state:    newState(currentView),
			store:    newStore(),
			signer:   signer,
			verifier: verifier,
		}

		sub, unsubFn := tm.store.SubscribeToPropose()
		defer unsubFn()

		// Make sure the message is added
		require.NoError(t, tm.AddMessage(message))

		// Make sure the message is present in the store
		var messages []*types.ProposalMessage
		select {
		case getMessages := <-sub:
			messages = getMessages()
		case <-time.After(5 * time.Second):
		}

		require.Len(t, messages, 1)

		storeMessage := messages[0]

		assert.Equal(
			t,
			message.GetProposalMessage().GetProposal(),
			storeMessage.GetProposal(),
		)
		assert.Equal(
			t,
			message.GetProposalMessage().GetProposalRound(),
			storeMessage.GetProposalRound(),
		)
		assert.Equal(
			t,
			message.GetProposalMessage().GetView(),
			storeMessage.GetView(),
		)
		assert.Equal(
			t,
			message.GetProposalMessage().GetFrom(),
			storeMessage.GetFrom(),
		)
	})

	t.Run("valid prevote message", func(t *testing.T) {
		t.Parallel()

		var (
			currentView = &types.View{
				Height: 1,
				Round:  10,
			}

			signature = []byte("valid signature")
			sender    = []byte("sender")
			id        = []byte("prevote ID")

			message = &types.Message{
				Type: types.MessageType_PREVOTE,
				Payload: &types.Message_PrevoteMessage{
					PrevoteMessage: &types.PrevoteMessage{
						View:       currentView,
						From:       sender,
						Identifier: id,
					},
				},
				Signature: signature,
			}

			signer = &mockSigner{
				isValidSignatureFn: func(_ []byte, sig []byte) bool {
					require.Equal(t, signature, sig)

					return true
				},
			}
			verifier = &mockVerifier{
				isValidatorFn: func(from []byte) bool {
					require.Equal(t, sender, from)

					return true
				},
			}
		)

		tm := &Tendermint{
			state:    newState(currentView),
			store:    newStore(),
			signer:   signer,
			verifier: verifier,
		}

		sub, unsubFn := tm.store.SubscribeToPrevote()
		defer unsubFn()

		// Make sure the message is added
		require.NoError(t, tm.AddMessage(message))

		// Make sure the message is present in the store
		var messages []*types.PrevoteMessage
		select {
		case getMessages := <-sub:
			messages = getMessages()
		case <-time.After(5 * time.Second):
		}

		require.Len(t, messages, 1)

		storeMessage := messages[0]

		assert.Equal(
			t,
			message.GetPrevoteMessage().GetIdentifier(),
			storeMessage.GetIdentifier(),
		)
		assert.Equal(
			t,
			message.GetPrevoteMessage().GetView(),
			storeMessage.GetView(),
		)
		assert.Equal(
			t,
			message.GetPrevoteMessage().GetFrom(),
			storeMessage.GetFrom(),
		)
	})

	t.Run("valid precommit message", func(t *testing.T) {
		t.Parallel()

		var (
			currentView = &types.View{
				Height: 1,
				Round:  10,
			}

			signature = []byte("valid signature")
			sender    = []byte("sender")
			id        = []byte("precommit ID")

			message = &types.Message{
				Type: types.MessageType_PRECOMMIT,
				Payload: &types.Message_PrecommitMessage{
					PrecommitMessage: &types.PrecommitMessage{
						View:       currentView,
						From:       sender,
						Identifier: id,
					},
				},
				Signature: signature,
			}

			signer = &mockSigner{
				isValidSignatureFn: func(_ []byte, sig []byte) bool {
					require.Equal(t, signature, sig)

					return true
				},
			}
			verifier = &mockVerifier{
				isValidatorFn: func(from []byte) bool {
					require.Equal(t, sender, from)

					return true
				},
			}
		)

		tm := &Tendermint{
			state:    newState(currentView),
			store:    newStore(),
			signer:   signer,
			verifier: verifier,
		}

		sub, unsubFn := tm.store.SubscribeToPrecommit()
		defer unsubFn()

		// Make sure the message is added
		require.NoError(t, tm.AddMessage(message))

		// Make sure the message is present in the store
		var messages []*types.PrecommitMessage
		select {
		case getMessages := <-sub:
			messages = getMessages()
		case <-time.After(5 * time.Second):
		}

		require.Len(t, messages, 1)

		storeMessage := messages[0]

		assert.Equal(
			t,
			message.GetPrecommitMessage().GetIdentifier(),
			storeMessage.GetIdentifier(),
		)
		assert.Equal(
			t,
			message.GetPrecommitMessage().GetView(),
			storeMessage.GetView(),
		)
		assert.Equal(
			t,
			message.GetPrecommitMessage().GetFrom(),
			storeMessage.GetFrom(),
		)
	})
}
