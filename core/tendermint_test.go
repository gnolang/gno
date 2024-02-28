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

	t.Run("invalid signature", func(t *testing.T) {
		t.Parallel()

		var (
			signature = []byte("invalid signature")
			message   = &types.PrevoteMessage{
				View:      &types.View{},
				Sender:    []byte{},
				Signature: signature,
			}

			signer = &mockSigner{
				isValidSignatureFn: func(_ []byte, sig []byte) bool {
					require.Equal(t, signature, sig)

					return false
				},
			}
			verifier = &mockVerifier{
				isValidatorFn: func(_ []byte) bool {
					return true
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
			tm.AddPrevoteMessage(message),
			ErrInvalidMessageSignature,
		)
	})

	t.Run("sender is not a validator", func(t *testing.T) {
		t.Parallel()

		var (
			signature = []byte("valid signature")
			sender    = []byte("sender")

			message = &types.PrevoteMessage{
				View:      &types.View{},
				Sender:    sender,
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
			tm.AddPrevoteMessage(message),
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

			message = &types.PrevoteMessage{
				View: &types.View{
					Height: currentView.Height - 1, // earlier height
					Round:  currentView.Round,
				},
				Sender:    sender,
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
			tm.AddPrevoteMessage(message),
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

			message = &types.PrevoteMessage{
				View: &types.View{
					Height: currentView.Height,
					Round:  currentView.Round - 1, // earlier round
				},
				Sender:    sender,
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
			tm.AddPrevoteMessage(message),
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

			message = &types.ProposalMessage{
				View:      nil, // invalid view
				Sender:    sender,
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
			tm.AddProposalMessage(message),
			types.ErrInvalidMessageView,
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

			message = &types.PrevoteMessage{
				View:      nil, // invalid view
				Sender:    sender,
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
			tm.AddPrevoteMessage(message),
			types.ErrInvalidMessageView,
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

			message = &types.PrecommitMessage{
				View:      nil, // invalid view
				Sender:    sender,
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
			tm.AddPrecommitMessage(message),
			types.ErrInvalidMessageView,
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

			message = &types.ProposalMessage{
				View:          currentView,
				Sender:        sender,
				Proposal:      proposal,
				ProposalRound: -1,
				Signature:     signature,
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
		require.NoError(t, tm.AddProposalMessage(message))

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
			message.GetProposal(),
			storeMessage.GetProposal(),
		)
		assert.Equal(
			t,
			message.GetProposalRound(),
			storeMessage.GetProposalRound(),
		)
		assert.Equal(
			t,
			message.GetView(),
			storeMessage.GetView(),
		)
		assert.Equal(
			t,
			message.GetSender(),
			storeMessage.GetSender(),
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

			message = &types.PrevoteMessage{
				View:       currentView,
				Sender:     sender,
				Identifier: id,
				Signature:  signature,
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
		require.NoError(t, tm.AddPrevoteMessage(message))

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
			message.GetIdentifier(),
			storeMessage.GetIdentifier(),
		)
		assert.Equal(
			t,
			message.GetView(),
			storeMessage.GetView(),
		)
		assert.Equal(
			t,
			message.GetSender(),
			storeMessage.GetSender(),
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

			message = &types.PrecommitMessage{
				View:       currentView,
				Sender:     sender,
				Identifier: id,
				Signature:  signature,
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
		require.NoError(t, tm.AddPrecommitMessage(message))

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
			message.GetIdentifier(),
			storeMessage.GetIdentifier(),
		)
		assert.Equal(
			t,
			message.GetView(),
			storeMessage.GetView(),
		)
		assert.Equal(
			t,
			message.GetSender(),
			storeMessage.GetSender(),
		)
	})
}
