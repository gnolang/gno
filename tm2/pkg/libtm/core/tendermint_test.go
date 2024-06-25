package core

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gnolang/libtm/messages/types"
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

		tm := NewTendermint(
			verifier,
			nil,
			nil,
			signer,
		)

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

		tm := NewTendermint(
			verifier,
			nil,
			nil,
			signer,
		)

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

		tm := NewTendermint(
			verifier,
			nil,
			nil,
			signer,
		)
		tm.state.setHeight(currentView.Height)
		tm.state.setRound(currentView.Round)

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

		tm := NewTendermint(
			verifier,
			nil,
			nil,
			signer,
		)
		tm.state.setHeight(currentView.Height)
		tm.state.setRound(currentView.Round)

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

		tm := NewTendermint(
			verifier,
			nil,
			nil,
			signer,
		)
		tm.state.setHeight(currentView.Height)
		tm.state.setRound(currentView.Round)

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

		tm := NewTendermint(
			verifier,
			nil,
			nil,
			signer,
		)
		tm.state.setHeight(currentView.Height)
		tm.state.setRound(currentView.Round)

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

		tm := NewTendermint(
			verifier,
			nil,
			nil,
			signer,
		)
		tm.state.setHeight(currentView.Height)
		tm.state.setRound(currentView.Round)

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

		tm := NewTendermint(
			verifier,
			nil,
			nil,
			signer,
		)
		tm.state.setHeight(currentView.Height)
		tm.state.setRound(currentView.Round)

		sub, unsubFn := tm.store.subscribeToPropose()
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

		tm := NewTendermint(
			verifier,
			nil,
			nil,
			signer,
		)
		tm.state.setHeight(currentView.Height)
		tm.state.setRound(currentView.Round)

		sub, unsubFn := tm.store.subscribeToPrevote()
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

		tm := NewTendermint(
			verifier,
			nil,
			nil,
			signer,
		)
		tm.state.setHeight(currentView.Height)
		tm.state.setRound(currentView.Round)

		sub, unsubFn := tm.store.subscribeToPrecommit()
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

func TestTendermint_FinalizeProposal_Propose(t *testing.T) {
	t.Parallel()

	t.Run("validator is the proposer", func(t *testing.T) {
		t.Parallel()

		t.Run("validator builds new proposal", func(t *testing.T) {
			t.Parallel()

			// Create the execution context
			ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelFn()

			var (
				id        = []byte("node ID")
				hash      = []byte("hash")
				signature = []byte("signature")
				view      = &types.View{
					Height: 10,
					Round:  0,
				}
				proposal = []byte("proposal")

				broadcastPropose *types.ProposalMessage
				broadcastPrevote *types.PrevoteMessage

				mockVerifier = &mockVerifier{
					isProposerFn: func(nodeID []byte, h uint64, r uint64) bool {
						require.Equal(t, id, nodeID)
						require.EqualValues(t, view.GetHeight(), h)
						require.EqualValues(t, view.GetRound(), r)

						return true
					},
				}
				mockNode = &mockNode{
					idFn: func() []byte {
						return id
					},
					buildProposalFn: func(h uint64) []byte {
						require.Equal(t, view.GetHeight(), h)

						return proposal
					},
					hashFn: func(p []byte) []byte {
						require.Equal(t, proposal, p)

						return hash
					},
				}
				mockSigner = &mockSigner{
					signFn: func(b []byte) []byte {
						require.NotNil(t, b)

						return signature
					},
				}
				mockBroadcast = &mockBroadcast{
					broadcastProposeFn: func(proposalMessage *types.ProposalMessage) {
						broadcastPropose = proposalMessage
					},
					broadcastPrevoteFn: func(prevoteMessage *types.PrevoteMessage) {
						broadcastPrevote = prevoteMessage

						// Stop the execution
						cancelFn()
					},
				}
			)

			// Create the tendermint instance
			tm := NewTendermint(
				mockVerifier,
				mockNode,
				mockBroadcast,
				mockSigner,
			)
			tm.state.setHeight(view.Height)
			tm.state.setRound(view.Round)

			// Run through the states
			tm.finalizeProposal(ctx)

			tm.wg.Wait()

			// Make sure the correct state was updated
			require.Equal(t, prevote, tm.state.step)
			assert.True(t, view.Equals(tm.state.view))
			assert.Equal(t, hash, tm.state.acceptedProposalID)

			// Make sure the broadcast propose was valid
			require.NotNil(t, broadcastPropose)
			require.NotNil(t, tm.state.acceptedProposal)
			assert.Equal(t, broadcastPropose.GetProposal(), tm.state.acceptedProposal)

			assert.True(t, view.Equals(broadcastPropose.GetView()))
			assert.Equal(t, id, broadcastPropose.GetSender())
			assert.Equal(t, signature, broadcastPropose.GetSignature())
			assert.Equal(t, proposal, broadcastPropose.GetProposal())
			assert.EqualValues(t, -1, broadcastPropose.GetProposalRound())

			// Make sure the broadcast prevote was valid
			require.NotNil(t, broadcastPrevote)
			assert.True(t, view.Equals(broadcastPrevote.GetView()))
			assert.Equal(t, id, broadcastPrevote.GetSender())
			assert.Equal(t, signature, broadcastPrevote.GetSignature())
			assert.Equal(t, hash, broadcastPrevote.GetIdentifier())
		})

		t.Run("validator uses old proposal", func(t *testing.T) {
			t.Parallel()

			// Create the execution context
			ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelFn()

			var (
				id        = []byte("node ID")
				hash      = []byte("hash")
				signature = []byte("signature")
				view      = &types.View{
					Height: 10,
					Round:  10,
				}
				proposal      = []byte("old proposal")
				proposalRound = int64(5)

				broadcastPropose *types.ProposalMessage
				broadcastPrevote *types.PrevoteMessage

				mockVerifier = &mockVerifier{
					isProposerFn: func(nodeID []byte, h uint64, r uint64) bool {
						require.Equal(t, id, nodeID)
						require.EqualValues(t, view.GetHeight(), h)
						require.EqualValues(t, view.GetRound(), r)

						return true
					},
				}
				mockNode = &mockNode{
					idFn: func() []byte {
						return id
					},
					buildProposalFn: func(_ uint64) []byte {
						t.FailNow()

						return nil
					},
					hashFn: func(p []byte) []byte {
						require.Equal(t, proposal, p)

						return hash
					},
				}
				mockSigner = &mockSigner{
					signFn: func(b []byte) []byte {
						require.NotNil(t, b)

						return signature
					},
				}
				mockBroadcast = &mockBroadcast{
					broadcastProposeFn: func(proposalMessage *types.ProposalMessage) {
						broadcastPropose = proposalMessage
					},
					broadcastPrevoteFn: func(prevoteMessage *types.PrevoteMessage) {
						broadcastPrevote = prevoteMessage

						// Stop the execution
						cancelFn()
					},
				}
			)

			// Create the tendermint instance
			tm := NewTendermint(
				mockVerifier,
				mockNode,
				mockBroadcast,
				mockSigner,
			)
			tm.state.setHeight(view.Height)
			tm.state.setRound(view.Round)

			// Set the old proposal
			tm.state.validValue = proposal
			tm.state.validRound = proposalRound

			// Run through the states
			tm.finalizeProposal(ctx)

			tm.wg.Wait()

			// Make sure the correct state was updated
			require.Equal(t, prevote, tm.state.step)
			assert.True(t, view.Equals(tm.state.view))
			assert.Equal(t, hash, tm.state.acceptedProposalID)

			// Make sure the broadcast propose was valid
			require.NotNil(t, broadcastPropose)
			require.NotNil(t, tm.state.acceptedProposal)
			assert.Equal(t, broadcastPropose.GetProposal(), tm.state.acceptedProposal)

			assert.True(t, view.Equals(broadcastPropose.GetView()))
			assert.Equal(t, id, broadcastPropose.GetSender())
			assert.Equal(t, signature, broadcastPropose.GetSignature())
			assert.Equal(t, proposal, broadcastPropose.GetProposal())
			assert.EqualValues(t, proposalRound, broadcastPropose.GetProposalRound())

			// Make sure the broadcast prevote was valid
			require.NotNil(t, broadcastPrevote)
			assert.True(t, view.Equals(broadcastPrevote.GetView()))
			assert.Equal(t, id, broadcastPrevote.GetSender())
			assert.Equal(t, signature, broadcastPrevote.GetSignature())
			assert.Equal(t, hash, broadcastPrevote.GetIdentifier())
		})
	})

	t.Run("validator is not the proposer", func(t *testing.T) {
		t.Parallel()

		t.Run("validator receives valid fresh proposal", func(t *testing.T) {
			t.Parallel()

			// Create the execution context
			ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelFn()

			var (
				proposerID = []byte("proposer ID")
				id         = []byte("node ID")
				hash       = []byte("hash")
				signature  = []byte("signature")
				view       = &types.View{
					Height: 10,
					Round:  0,
				}
				proposal = []byte("proposal")

				proposalMessage = &types.ProposalMessage{
					View:          view,
					Sender:        proposerID,
					Signature:     signature,
					Proposal:      proposal,
					ProposalRound: -1,
				}

				broadcastPrevote *types.PrevoteMessage

				mockVerifier = &mockVerifier{
					isProposerFn: func(nodeID []byte, h uint64, r uint64) bool {
						if bytes.Equal(id, nodeID) {
							return false
						}

						require.Equal(t, proposerID, nodeID)
						require.EqualValues(t, view.GetHeight(), h)
						require.EqualValues(t, view.GetRound(), r)

						return true
					},
					isValidatorFn: func(id []byte) bool {
						require.Equal(t, proposerID, id)

						return true
					},
					isValidProposalFn: func(p []byte, h uint64) bool {
						require.Equal(t, proposal, p)
						require.EqualValues(t, view.GetHeight(), h)

						return true
					},
				}
				mockNode = &mockNode{
					idFn: func() []byte {
						return id
					},
					hashFn: func(p []byte) []byte {
						require.Equal(t, proposal, p)

						return hash
					},
				}
				mockSigner = &mockSigner{
					signFn: func(b []byte) []byte {
						require.NotNil(t, b)

						return signature
					},
					isValidSignatureFn: func(raw []byte, signed []byte) bool {
						require.Equal(t, proposalMessage.GetSignaturePayload(), raw)
						require.Equal(t, signature, signed)

						return true
					},
				}
				mockBroadcast = &mockBroadcast{
					broadcastPrevoteFn: func(prevoteMessage *types.PrevoteMessage) {
						broadcastPrevote = prevoteMessage

						// Stop the execution
						cancelFn()
					},
				}
			)

			// Create the tendermint instance
			tm := NewTendermint(mockVerifier, mockNode, mockBroadcast, mockSigner)
			tm.state.setHeight(view.Height)
			tm.state.setRound(view.Round)

			// Add in the proposal message
			require.NoError(t, tm.AddProposalMessage(proposalMessage))

			// Run through the states
			tm.finalizeProposal(ctx)

			tm.wg.Wait()

			// Make sure the correct state was updated
			require.Equal(t, prevote, tm.state.step)
			assert.True(t, view.Equals(tm.state.view))
			assert.Equal(t, hash, tm.state.acceptedProposalID)

			// Make sure the correct proposal was accepted
			assert.Equal(t, proposalMessage.GetProposal(), tm.state.acceptedProposal)
			assert.Equal(t, hash, tm.state.acceptedProposalID)

			// Make sure the broadcast prevote was valid
			require.NotNil(t, broadcastPrevote)
			assert.True(t, view.Equals(broadcastPrevote.GetView()))
			assert.Equal(t, id, broadcastPrevote.GetSender())
			assert.Equal(t, signature, broadcastPrevote.GetSignature())
			assert.Equal(t, hash, broadcastPrevote.GetIdentifier())
		})

		t.Run("validator receives valid locked proposal", func(t *testing.T) {
			t.Parallel()

			// Create the execution context
			ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelFn()

			var (
				proposerID = []byte("proposer ID")
				id         = []byte("node ID")
				hash       = []byte("hash")
				signature  = []byte("signature")
				view       = &types.View{
					Height: 10,
					Round:  10,
				}
				proposal      = []byte("proposal")
				proposalRound = int64(5)

				proposalMessage = &types.ProposalMessage{
					View:          view,
					Sender:        proposerID,
					Signature:     signature,
					Proposal:      proposal,
					ProposalRound: proposalRound,
				}

				broadcastPrevote *types.PrevoteMessage

				mockVerifier = &mockVerifier{
					isProposerFn: func(nodeID []byte, h uint64, r uint64) bool {
						if bytes.Equal(id, nodeID) {
							return false
						}

						require.Equal(t, proposerID, nodeID)
						require.EqualValues(t, view.GetHeight(), h)
						require.EqualValues(t, view.GetRound(), r)

						return true
					},
					isValidatorFn: func(id []byte) bool {
						require.Equal(t, proposerID, id)

						return true
					},
					isValidProposalFn: func(p []byte, h uint64) bool {
						require.Equal(t, proposal, p)
						require.EqualValues(t, view.GetHeight(), h)

						return true
					},
				}
				mockNode = &mockNode{
					idFn: func() []byte {
						return id
					},
					hashFn: func(p []byte) []byte {
						require.Equal(t, proposal, p)

						return hash
					},
				}
				mockSigner = &mockSigner{
					signFn: func(b []byte) []byte {
						require.NotNil(t, b)

						return signature
					},
					isValidSignatureFn: func(raw []byte, signed []byte) bool {
						require.Equal(t, proposalMessage.GetSignaturePayload(), raw)
						require.Equal(t, signature, signed)

						return true
					},
				}
				mockBroadcast = &mockBroadcast{
					broadcastPrevoteFn: func(prevoteMessage *types.PrevoteMessage) {
						broadcastPrevote = prevoteMessage

						// Stop the execution
						cancelFn()
					},
				}
			)

			// Create the tendermint instance
			tm := NewTendermint(mockVerifier, mockNode, mockBroadcast, mockSigner)
			tm.state.setHeight(view.Height)
			tm.state.setRound(view.Round)

			// Add in the proposal message
			require.NoError(t, tm.AddProposalMessage(proposalMessage))

			// Run through the states
			tm.finalizeProposal(ctx)

			tm.wg.Wait()

			// Make sure the correct state was updated
			require.Equal(t, prevote, tm.state.step)
			assert.True(t, view.Equals(tm.state.view))
			assert.Equal(t, hash, tm.state.acceptedProposalID)

			// Make sure the correct proposal was accepted
			assert.Equal(t, proposalMessage.GetProposal(), tm.state.acceptedProposal)
			assert.Equal(t, hash, tm.state.acceptedProposalID)

			// Make sure the broadcast prevote was valid
			require.NotNil(t, broadcastPrevote)
			assert.True(t, view.Equals(broadcastPrevote.GetView()))
			assert.Equal(t, id, broadcastPrevote.GetSender())
			assert.Equal(t, signature, broadcastPrevote.GetSignature())
			assert.Equal(t, hash, broadcastPrevote.GetIdentifier())
		})

		t.Run("validator receives invalid fresh proposal", func(t *testing.T) {
			t.Parallel()

			// Create the execution context
			ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelFn()

			var (
				proposerID = []byte("proposer ID")
				id         = []byte("node ID")
				hash       = []byte("hash")
				signature  = []byte("signature")
				view       = &types.View{
					Height: 10,
					Round:  5,
				}
				proposal = []byte("proposal")

				lockedRound = int64(view.Round - 1) // earlier round

				proposalMessage = &types.ProposalMessage{
					View:          view,
					Sender:        proposerID,
					Signature:     signature,
					Proposal:      proposal,
					ProposalRound: -1,
				}

				broadcastPrevote *types.PrevoteMessage

				mockVerifier = &mockVerifier{
					isProposerFn: func(nodeID []byte, h uint64, r uint64) bool {
						if bytes.Equal(id, nodeID) {
							return false
						}

						require.Equal(t, proposerID, nodeID)
						require.EqualValues(t, view.GetHeight(), h)
						require.EqualValues(t, view.GetRound(), r)

						return true
					},
					isValidatorFn: func(id []byte) bool {
						require.Equal(t, proposerID, id)

						return true
					},
					isValidProposalFn: func(p []byte, h uint64) bool {
						require.Equal(t, proposal, p)
						require.EqualValues(t, view.GetHeight(), h)

						return true
					},
				}
				mockNode = &mockNode{
					idFn: func() []byte {
						return id
					},
					hashFn: func(p []byte) []byte {
						require.Equal(t, proposal, p)

						return hash
					},
				}
				mockSigner = &mockSigner{
					signFn: func(b []byte) []byte {
						require.NotNil(t, b)

						return signature
					},
					isValidSignatureFn: func(raw []byte, signed []byte) bool {
						require.Equal(t, proposalMessage.GetSignaturePayload(), raw)
						require.Equal(t, signature, signed)

						return true
					},
				}
				mockBroadcast = &mockBroadcast{
					broadcastPrevoteFn: func(prevoteMessage *types.PrevoteMessage) {
						broadcastPrevote = prevoteMessage

						// Stop the execution
						cancelFn()
					},
				}
			)

			// Create the tendermint instance
			tm := NewTendermint(mockVerifier, mockNode, mockBroadcast, mockSigner)
			tm.state.setHeight(view.Height)
			tm.state.setRound(view.Round)

			// Set the locked round
			tm.state.lockedRound = lockedRound

			// Add in the proposal message
			require.NoError(t, tm.AddProposalMessage(proposalMessage))

			// Run through the states
			tm.finalizeProposal(ctx)

			tm.wg.Wait()

			// Make sure the correct state was updated
			require.Equal(t, prevote, tm.state.step)
			assert.True(t, view.Equals(tm.state.view))

			// Make sure the correct proposal was not accepted
			assert.Nil(t, tm.state.acceptedProposal)
			assert.Nil(t, tm.state.acceptedProposalID)

			// Make sure the broadcast prevote was NIL
			require.NotNil(t, broadcastPrevote)
			assert.True(t, view.Equals(broadcastPrevote.GetView()))
			assert.Equal(t, id, broadcastPrevote.GetSender())
			assert.Equal(t, signature, broadcastPrevote.GetSignature())
			assert.Nil(t, broadcastPrevote.GetIdentifier())
		})

		t.Run("validator receives locked proposal from an invalid round", func(t *testing.T) {
			t.Parallel()

			// Create the execution context
			ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelFn()

			var (
				proposerID = []byte("proposer ID")
				id         = []byte("node ID")
				hash       = []byte("hash")
				signature  = []byte("signature")
				view       = &types.View{
					Height: 10,
					Round:  5,
				}
				proposal = []byte("proposal")

				lockedRound = int64(view.Round - 1) // earlier round

				proposalMessage = &types.ProposalMessage{
					View:          view,
					Sender:        proposerID,
					Signature:     signature,
					Proposal:      proposal,
					ProposalRound: lockedRound + 1, // invalid proposal round
				}

				broadcastPrevote *types.PrevoteMessage

				mockVerifier = &mockVerifier{
					isProposerFn: func(nodeID []byte, h uint64, r uint64) bool {
						if bytes.Equal(id, nodeID) {
							return false
						}

						require.Equal(t, proposerID, nodeID)
						require.EqualValues(t, view.GetHeight(), h)
						require.EqualValues(t, view.GetRound(), r)

						return true
					},
					isValidatorFn: func(id []byte) bool {
						require.Equal(t, proposerID, id)

						return true
					},
					isValidProposalFn: func(p []byte, h uint64) bool {
						require.Equal(t, proposal, p)
						require.EqualValues(t, view.GetHeight(), h)

						return true
					},
				}
				mockNode = &mockNode{
					idFn: func() []byte {
						return id
					},
					hashFn: func(p []byte) []byte {
						require.Equal(t, proposal, p)

						return hash
					},
				}
				mockSigner = &mockSigner{
					signFn: func(b []byte) []byte {
						require.NotNil(t, b)

						return signature
					},
					isValidSignatureFn: func(raw []byte, signed []byte) bool {
						require.Equal(t, proposalMessage.GetSignaturePayload(), raw)
						require.Equal(t, signature, signed)

						return true
					},
				}
				mockBroadcast = &mockBroadcast{
					broadcastPrevoteFn: func(prevoteMessage *types.PrevoteMessage) {
						broadcastPrevote = prevoteMessage

						// Stop the execution
						cancelFn()
					},
				}
			)

			// Create the tendermint instance
			tm := NewTendermint(mockVerifier, mockNode, mockBroadcast, mockSigner)
			tm.state.setHeight(view.Height)
			tm.state.setRound(view.Round)

			// Set the locked round
			tm.state.lockedRound = lockedRound

			// Add in the proposal message
			require.NoError(t, tm.AddProposalMessage(proposalMessage))

			// Run through the states
			tm.finalizeProposal(ctx)

			tm.wg.Wait()

			// Make sure the correct state was updated
			require.Equal(t, prevote, tm.state.step)
			assert.True(t, view.Equals(tm.state.view))

			// Make sure the correct proposal was not accepted
			assert.Nil(t, tm.state.acceptedProposal)
			assert.Nil(t, tm.state.acceptedProposalID)

			// Make sure the broadcast prevote was NIL
			require.NotNil(t, broadcastPrevote)
			assert.True(t, view.Equals(broadcastPrevote.GetView()))
			assert.Equal(t, id, broadcastPrevote.GetSender())
			assert.Equal(t, signature, broadcastPrevote.GetSignature())
			assert.Nil(t, broadcastPrevote.GetIdentifier())
		})

		t.Run("validator receives invalid locked proposal (round mismatch)", func(t *testing.T) {
			t.Parallel()

			// Create the execution context
			ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelFn()

			var (
				proposerID = []byte("proposer ID")
				id         = []byte("node ID")
				hash       = []byte("hash")
				signature  = []byte("signature")
				view       = &types.View{
					Height: 10,
					Round:  5,
				}
				proposal = []byte("proposal")

				lockedRound = int64(view.Round - 1) // earlier round

				proposalMessage = &types.ProposalMessage{
					View:          view,
					Sender:        proposerID,
					Signature:     signature,
					Proposal:      proposal,
					ProposalRound: lockedRound - 1, // invalid proposal round
				}

				broadcastPrevote *types.PrevoteMessage

				mockVerifier = &mockVerifier{
					isProposerFn: func(nodeID []byte, h uint64, r uint64) bool {
						if bytes.Equal(id, nodeID) {
							return false
						}

						require.Equal(t, proposerID, nodeID)
						require.EqualValues(t, view.GetHeight(), h)
						require.EqualValues(t, view.GetRound(), r)

						return true
					},
					isValidatorFn: func(id []byte) bool {
						require.Equal(t, proposerID, id)

						return true
					},
					isValidProposalFn: func(p []byte, h uint64) bool {
						require.Equal(t, proposal, p)
						require.EqualValues(t, view.GetHeight(), h)

						return true
					},
				}
				mockNode = &mockNode{
					idFn: func() []byte {
						return id
					},
					hashFn: func(p []byte) []byte {
						require.Equal(t, proposal, p)

						return hash
					},
				}
				mockSigner = &mockSigner{
					signFn: func(b []byte) []byte {
						require.NotNil(t, b)

						return signature
					},
					isValidSignatureFn: func(raw []byte, signed []byte) bool {
						require.Equal(t, proposalMessage.GetSignaturePayload(), raw)
						require.Equal(t, signature, signed)

						return true
					},
				}
				mockBroadcast = &mockBroadcast{
					broadcastPrevoteFn: func(prevoteMessage *types.PrevoteMessage) {
						broadcastPrevote = prevoteMessage

						// Stop the execution
						cancelFn()
					},
				}
			)

			// Create the tendermint instance
			tm := NewTendermint(mockVerifier, mockNode, mockBroadcast, mockSigner)
			tm.state.setHeight(view.Height)
			tm.state.setRound(view.Round)

			// Set the locked round
			tm.state.lockedRound = lockedRound

			// Add in the proposal message
			require.NoError(t, tm.AddProposalMessage(proposalMessage))

			// Run through the states
			tm.finalizeProposal(ctx)

			tm.wg.Wait()

			// Make sure the correct state was updated
			require.Equal(t, prevote, tm.state.step)
			assert.True(t, view.Equals(tm.state.view))

			// Make sure the correct proposal was not accepted
			assert.Nil(t, tm.state.acceptedProposal)
			assert.Nil(t, tm.state.acceptedProposalID)

			// Make sure the broadcast prevote was NIL
			require.NotNil(t, broadcastPrevote)
			assert.True(t, view.Equals(broadcastPrevote.GetView()))
			assert.Equal(t, id, broadcastPrevote.GetSender())
			assert.Equal(t, signature, broadcastPrevote.GetSignature())
			assert.Nil(t, broadcastPrevote.GetIdentifier())
		})

		t.Run("validator receives a proposal that is not valid", func(t *testing.T) {
			t.Parallel()

			// Create the execution context
			ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelFn()

			var (
				proposerID = []byte("proposer ID")
				id         = []byte("node ID")
				hash       = []byte("hash")
				signature  = []byte("signature")
				view       = &types.View{
					Height: 10,
					Round:  0,
				}
				proposal = []byte("proposal")

				proposalMessage = &types.ProposalMessage{
					View:          view,
					Sender:        proposerID,
					Signature:     signature,
					Proposal:      proposal,
					ProposalRound: -1,
				}

				broadcastPrevote *types.PrevoteMessage

				mockVerifier = &mockVerifier{
					isProposerFn: func(nodeID []byte, h uint64, r uint64) bool {
						if bytes.Equal(id, nodeID) {
							return false
						}

						require.Equal(t, proposerID, nodeID)
						require.EqualValues(t, view.GetHeight(), h)
						require.EqualValues(t, view.GetRound(), r)

						return true
					},
					isValidatorFn: func(id []byte) bool {
						require.Equal(t, proposerID, id)

						return true
					},
					isValidProposalFn: func(p []byte, h uint64) bool {
						require.Equal(t, proposal, p)
						require.EqualValues(t, view.GetHeight(), h)

						return false // invalid proposal
					},
				}
				mockNode = &mockNode{
					idFn: func() []byte {
						return id
					},
					hashFn: func(p []byte) []byte {
						require.Equal(t, proposal, p)

						return hash
					},
				}
				mockSigner = &mockSigner{
					signFn: func(b []byte) []byte {
						require.NotNil(t, b)

						return signature
					},
					isValidSignatureFn: func(raw []byte, signed []byte) bool {
						require.Equal(t, proposalMessage.GetSignaturePayload(), raw)
						require.Equal(t, signature, signed)

						return true
					},
				}
				mockBroadcast = &mockBroadcast{
					broadcastPrevoteFn: func(prevoteMessage *types.PrevoteMessage) {
						broadcastPrevote = prevoteMessage

						// Stop the execution
						cancelFn()
					},
				}
			)

			// Create the tendermint instance
			tm := NewTendermint(mockVerifier, mockNode, mockBroadcast, mockSigner)
			tm.state.setHeight(view.Height)
			tm.state.setRound(view.Round)

			// Add in the proposal message
			require.NoError(t, tm.AddProposalMessage(proposalMessage))

			// Run through the states
			tm.finalizeProposal(ctx)

			tm.wg.Wait()

			// Make sure the correct state was updated
			require.Equal(t, prevote, tm.state.step)
			assert.True(t, view.Equals(tm.state.view))

			// Make sure the correct proposal was not accepted
			assert.Nil(t, tm.state.acceptedProposal)
			assert.Nil(t, tm.state.acceptedProposalID)

			// Make sure the broadcast prevote was NIL
			require.NotNil(t, broadcastPrevote)
			assert.True(t, view.Equals(broadcastPrevote.GetView()))
			assert.Equal(t, id, broadcastPrevote.GetSender())
			assert.Equal(t, signature, broadcastPrevote.GetSignature())
			assert.Nil(t, broadcastPrevote.GetIdentifier())
		})

		t.Run("validator does not receive a proposal in time", func(t *testing.T) {
			t.Parallel()

			// Create the execution context
			ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelFn()

			var (
				proposerID = []byte("proposer ID")
				id         = []byte("node ID")
				signature  = []byte("signature")
				view       = &types.View{
					Height: 10,
					Round:  0,
				}

				timeout = Timeout{
					Initial: 100 * time.Millisecond,
					Delta:   0,
				}

				broadcastPrevote *types.PrevoteMessage

				mockVerifier = &mockVerifier{
					isProposerFn: func(nodeID []byte, h uint64, r uint64) bool {
						if bytes.Equal(id, nodeID) {
							return false
						}

						require.Equal(t, proposerID, nodeID)
						require.EqualValues(t, view.GetHeight(), h)
						require.EqualValues(t, view.GetRound(), r)

						return true
					},
					isValidatorFn: func(id []byte) bool {
						require.Equal(t, proposerID, id)

						return true
					},
				}
				mockNode = &mockNode{
					idFn: func() []byte {
						return id
					},
				}
				mockSigner = &mockSigner{
					signFn: func(b []byte) []byte {
						require.NotNil(t, b)

						return signature
					},
				}
				mockBroadcast = &mockBroadcast{
					broadcastPrevoteFn: func(prevoteMessage *types.PrevoteMessage) {
						broadcastPrevote = prevoteMessage

						// Stop the execution
						cancelFn()
					},
				}
			)

			// Create the tendermint instance
			tm := NewTendermint(
				mockVerifier,
				mockNode,
				mockBroadcast,
				mockSigner,
				WithProposeTimeout(timeout),
			)
			tm.state.setHeight(view.Height)
			tm.state.setRound(view.Round)

			// Run through the states
			tm.finalizeProposal(ctx)

			tm.wg.Wait()

			// Make sure the correct state was updated
			require.Equal(t, prevote, tm.state.step)
			assert.True(t, view.Equals(tm.state.view))

			// Make sure the correct proposal was not accepted
			assert.Nil(t, tm.state.acceptedProposal)
			assert.Nil(t, tm.state.acceptedProposalID)

			// Make sure the broadcast prevote was NIL
			require.NotNil(t, broadcastPrevote)
			assert.True(t, view.Equals(broadcastPrevote.GetView()))
			assert.Equal(t, id, broadcastPrevote.GetSender())
			assert.Equal(t, signature, broadcastPrevote.GetSignature())
			assert.Nil(t, broadcastPrevote.GetIdentifier())
		})
	})
}

// generatePrevoteMessages generates basic prevote messages
// using the given view and ID
func generatePrevoteMessages(
	t *testing.T,
	count int,
	view *types.View,
	id []byte,
) []*types.PrevoteMessage {
	t.Helper()

	messages := make([]*types.PrevoteMessage, count)

	for i := 0; i < count; i++ {
		messages[i] = &types.PrevoteMessage{
			View:       view,
			Sender:     []byte(fmt.Sprintf("sender %d", i)),
			Signature:  []byte("signature"),
			Identifier: id,
		}
	}

	return messages
}

// generatePrecommitMessages generates basic precommit messages
// using the given view and ID
func generatePrecommitMessages(
	t *testing.T,
	count int,
	view *types.View,
	id []byte,
) []*types.PrecommitMessage {
	t.Helper()

	messages := make([]*types.PrecommitMessage, count)

	for i := 0; i < count; i++ {
		messages[i] = &types.PrecommitMessage{
			View:       view,
			Sender:     []byte(fmt.Sprintf("sender %d", i)),
			Signature:  []byte("signature"),
			Identifier: id,
		}
	}

	return messages
}

func TestTendermint_FinalizeProposal_Prevote(t *testing.T) {
	t.Parallel()

	t.Run("validator received 2F+1 PREVOTEs with a valid ID", func(t *testing.T) {
		t.Parallel()

		// Create the execution context
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		var (
			id        = []byte("node ID")
			signature = []byte("signature")
			view      = &types.View{
				Height: 10,
				Round:  0,
			}
			proposalID = []byte("proposal ID")
			proposal   = []byte("proposal")

			proposalMessage = &types.ProposalMessage{
				View:          view,
				Sender:        []byte("proposer"),
				Signature:     []byte("proposer signature"),
				Proposal:      proposal,
				ProposalRound: -1,
			}

			numPrevotes     = 10
			prevoteMessages = generatePrevoteMessages(t, numPrevotes, view, proposalID)

			broadcastPrecommit *types.PrecommitMessage

			mockVerifier = &mockVerifier{
				getTotalVotingPowerFn: func(h uint64) uint64 {
					require.EqualValues(t, view.Height, h)

					return uint64(numPrevotes)
				},
				getSumVotingPowerFn: func(messages []Message) uint64 {
					return uint64(len(messages))
				},
				isValidatorFn: func(_ []byte) bool {
					return true
				},
			}

			mockNode = &mockNode{
				idFn: func() []byte {
					return id
				},
			}
			mockSigner = &mockSigner{
				signFn: func(b []byte) []byte {
					require.NotNil(t, b)

					return signature
				},
				isValidSignatureFn: func(_ []byte, _ []byte) bool {
					return true
				},
			}
			mockBroadcast = &mockBroadcast{
				broadcastPrecommitFn: func(precommitMessage *types.PrecommitMessage) {
					broadcastPrecommit = precommitMessage

					// Stop the execution
					cancelFn()
				},
			}
		)

		// Create the tendermint instance
		tm := NewTendermint(mockVerifier, mockNode, mockBroadcast, mockSigner)
		tm.state.setHeight(view.Height)
		tm.state.setRound(view.Round)
		tm.state.step = prevote
		tm.state.acceptedProposal = proposal
		tm.state.acceptedProposalID = proposalID

		// Add in 2F+1 non-NIL prevote messages
		for _, prevoteMessage := range prevoteMessages {
			require.NoError(t, tm.AddPrevoteMessage(prevoteMessage))
		}

		// Run through the states
		tm.finalizeProposal(ctx)

		tm.wg.Wait()

		// Make sure the correct state was updated
		require.Equal(t, precommit, tm.state.step)
		assert.True(t, view.Equals(tm.state.view))
		assert.EqualValues(t, view.Round, tm.state.lockedRound)
		assert.Equal(t, proposalMessage.GetProposal(), tm.state.lockedValue)
		assert.Equal(t, proposalMessage.GetProposal(), tm.state.validValue)
		assert.EqualValues(t, view.Round, tm.state.validRound)

		// Make sure the broadcast precommit was valid
		require.NotNil(t, broadcastPrecommit)
		assert.True(t, view.Equals(broadcastPrecommit.GetView()))
		assert.Equal(t, id, broadcastPrecommit.GetSender())
		assert.Equal(t, signature, broadcastPrecommit.GetSignature())
		assert.Equal(t, proposalID, broadcastPrecommit.GetIdentifier())
	})

	t.Run("validator received 2F+1 PREVOTEs with a NIL ID", func(t *testing.T) {
		t.Parallel()

		// Create the execution context
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		var (
			id        = []byte("node ID")
			signature = []byte("signature")
			view      = &types.View{
				Height: 10,
				Round:  0,
			}
			proposalID = []byte("proposal ID")
			proposal   = []byte("proposal")

			numPrevotes     = 10
			prevoteMessages = generatePrevoteMessages(t, numPrevotes, view, nil)

			broadcastPrecommit *types.PrecommitMessage

			mockVerifier = &mockVerifier{
				getTotalVotingPowerFn: func(h uint64) uint64 {
					require.EqualValues(t, view.Height, h)

					return uint64(numPrevotes)
				},
				getSumVotingPowerFn: func(messages []Message) uint64 {
					return uint64(len(messages))
				},
				isValidatorFn: func(_ []byte) bool {
					return true
				},
			}

			mockNode = &mockNode{
				idFn: func() []byte {
					return id
				},
			}
			mockSigner = &mockSigner{
				signFn: func(b []byte) []byte {
					require.NotNil(t, b)

					return signature
				},
				isValidSignatureFn: func(_ []byte, _ []byte) bool {
					return true
				},
			}
			mockBroadcast = &mockBroadcast{
				broadcastPrecommitFn: func(precommitMessage *types.PrecommitMessage) {
					broadcastPrecommit = precommitMessage

					// Stop the execution
					cancelFn()
				},
			}
		)

		// Create the tendermint instance
		tm := NewTendermint(mockVerifier, mockNode, mockBroadcast, mockSigner)
		tm.state.setHeight(view.Height)
		tm.state.setRound(view.Round)
		tm.state.step = prevote
		tm.state.acceptedProposal = proposal
		tm.state.acceptedProposalID = proposalID

		// Add in 2F+1 non-NIL prevote messages
		for _, prevoteMessage := range prevoteMessages {
			require.NoError(t, tm.AddPrevoteMessage(prevoteMessage))
		}

		// Run through the states
		tm.finalizeProposal(ctx)

		tm.wg.Wait()

		// Make sure the correct state was updated
		require.Equal(t, precommit, tm.state.step)
		assert.True(t, view.Equals(tm.state.view))

		assert.EqualValues(t, -1, tm.state.lockedRound)
		assert.Nil(t, tm.state.lockedValue)
		assert.Nil(t, tm.state.validValue)
		assert.EqualValues(t, -1, tm.state.validRound)

		// Make sure the broadcast precommit was valid
		require.NotNil(t, broadcastPrecommit)
		assert.True(t, view.Equals(broadcastPrecommit.GetView()))
		assert.Equal(t, id, broadcastPrecommit.GetSender())
		assert.Equal(t, signature, broadcastPrecommit.GetSignature())
		assert.Nil(t, broadcastPrecommit.GetIdentifier())
	})

	t.Run("validator does not receive quorum PREVOTEs in time", func(t *testing.T) {
		t.Parallel()

		// Create the execution context
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		var (
			id        = []byte("node ID")
			signature = []byte("signature")
			view      = &types.View{
				Height: 10,
				Round:  0,
			}
			proposalID = []byte("proposal ID")
			proposal   = []byte("proposal")

			totalPrevoteCount     = 10
			nilPrevoteMessages    = generatePrevoteMessages(t, totalPrevoteCount/2, view, nil)
			nonNilPrevoteMessages = generatePrevoteMessages(t, totalPrevoteCount/2, view, proposalID)

			timeout = Timeout{
				Initial: 100 * time.Millisecond,
				Delta:   0,
			}

			broadcastPrecommit *types.PrecommitMessage

			mockVerifier = &mockVerifier{
				getTotalVotingPowerFn: func(h uint64) uint64 {
					require.EqualValues(t, view.Height, h)

					return uint64(totalPrevoteCount)
				},
				getSumVotingPowerFn: func(messages []Message) uint64 {
					return uint64(len(messages))
				},
				isValidatorFn: func(_ []byte) bool {
					return true
				},
			}

			mockNode = &mockNode{
				idFn: func() []byte {
					return id
				},
			}
			mockSigner = &mockSigner{
				signFn: func(b []byte) []byte {
					require.NotNil(t, b)

					return signature
				},
				isValidSignatureFn: func(_ []byte, _ []byte) bool {
					return true
				},
			}
			mockBroadcast = &mockBroadcast{
				broadcastPrecommitFn: func(precommitMessage *types.PrecommitMessage) {
					broadcastPrecommit = precommitMessage

					// Stop the execution
					cancelFn()
				},
			}
		)

		// Create the tendermint instance
		tm := NewTendermint(
			mockVerifier,
			mockNode,
			mockBroadcast,
			mockSigner,
			WithPrevoteTimeout(timeout),
		)
		tm.state.setHeight(view.Height)
		tm.state.setRound(view.Round)
		tm.state.step = prevote
		tm.state.acceptedProposal = proposal
		tm.state.acceptedProposalID = proposalID

		// Add in non-NIL prevote messages
		for index, prevoteMessage := range nonNilPrevoteMessages {
			// Change the senders for the non-NIL prevote messages.
			// The reason the senders need to be changed is, so we can simulate the following scenario:
			// 1/2 of the total voting power sent in non-NIL prevote messages
			// 1/2 of the total voting power sent in NIL prevote messages
			// In turn, there is a super majority when their voting powers are summed (non-NIL and NIL)
			prevoteMessage.Sender = []byte(fmt.Sprintf("sender %d", index))

			require.NoError(t, tm.AddPrevoteMessage(prevoteMessage))
		}

		// Add in NIL prevote messages
		for index, prevoteMessage := range nilPrevoteMessages {
			// Change the senders for the NIL prevote messages
			prevoteMessage.Sender = []byte(fmt.Sprintf("sender %d", index+len(nonNilPrevoteMessages)))

			require.NoError(t, tm.AddPrevoteMessage(prevoteMessage))
		}

		// Run through the states
		tm.finalizeProposal(ctx)

		tm.wg.Wait()

		// Make sure the correct state was updated
		require.Equal(t, precommit, tm.state.step)
		assert.True(t, view.Equals(tm.state.view))

		assert.EqualValues(t, -1, tm.state.lockedRound)
		assert.Nil(t, tm.state.lockedValue)
		assert.Nil(t, tm.state.validValue)
		assert.EqualValues(t, -1, tm.state.validRound)

		// Make sure the broadcast precommit was valid
		require.NotNil(t, broadcastPrecommit)
		assert.True(t, view.Equals(broadcastPrecommit.GetView()))
		assert.Equal(t, id, broadcastPrecommit.GetSender())
		assert.Equal(t, signature, broadcastPrecommit.GetSignature())
		assert.Nil(t, broadcastPrecommit.GetIdentifier())
	})
}

func TestTendermint_FinalizeProposal_Precommit(t *testing.T) {
	t.Parallel()

	t.Run("validator received 2F+1 PRECOMMITs with a valid ID", func(t *testing.T) {
		t.Parallel()

		// Create the execution context
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		var (
			id        = []byte("node ID")
			signature = []byte("signature")
			view      = &types.View{
				Height: 10,
				Round:  0,
			}
			proposalID = []byte("proposal ID")
			proposal   = []byte("proposal")

			proposalMessage = &types.ProposalMessage{
				Proposal: proposal,
			}

			numPrecommits     = 10
			precommitMessages = generatePrecommitMessages(t, numPrecommits, view, proposalID)

			mockVerifier = &mockVerifier{
				getTotalVotingPowerFn: func(h uint64) uint64 {
					require.EqualValues(t, view.Height, h)

					return uint64(numPrecommits)
				},
				getSumVotingPowerFn: func(messages []Message) uint64 {
					return uint64(len(messages))
				},
				isValidatorFn: func(_ []byte) bool {
					return true
				},
			}

			mockNode = &mockNode{
				idFn: func() []byte {
					return id
				},
			}
			mockSigner = &mockSigner{
				signFn: func(b []byte) []byte {
					require.NotNil(t, b)

					return signature
				},
				isValidSignatureFn: func(_ []byte, _ []byte) bool {
					return true
				},
			}
		)

		// Create the tendermint instance
		tm := NewTendermint(mockVerifier, mockNode, &mockBroadcast{}, mockSigner)
		tm.state.setHeight(view.Height)
		tm.state.setRound(view.Round)
		tm.state.step = precommit
		tm.state.acceptedProposal = proposal
		tm.state.acceptedProposalID = proposalID

		// Add in 2F+1 non-NIL precommit messages
		for _, precommitMessage := range precommitMessages {
			require.NoError(t, tm.AddPrecommitMessage(precommitMessage))
		}

		// Run through the states
		finalizedProposalCh := tm.finalizeProposal(ctx)

		// Get the finalized proposal
		finalizedProposal := <-finalizedProposalCh

		cancelFn()
		tm.wg.Wait()

		// Make sure the finalized proposal is valid
		require.NotNil(t, finalizedProposal)
		assert.Equal(t, proposalMessage.Proposal, finalizedProposal.Data)
	})

	t.Run("validator does not receive quorum PRECOMMITs in time", func(t *testing.T) {
		t.Parallel()

		// Create the execution context
		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		var (
			id        = []byte("node ID")
			signature = []byte("signature")
			view      = &types.View{
				Height: 10,
				Round:  0,
			}
			proposalID = []byte("proposal ID")
			proposal   = []byte("proposal")

			totalPrecommitCount     = 10
			nilPrecommitMessages    = generatePrecommitMessages(t, totalPrecommitCount/2, view, nil)
			nonNilPrecommitMessages = generatePrecommitMessages(t, totalPrecommitCount/2, view, proposalID)

			timeout = Timeout{
				Initial: 100 * time.Millisecond,
				Delta:   0,
			}

			mockVerifier = &mockVerifier{
				getTotalVotingPowerFn: func(h uint64) uint64 {
					require.EqualValues(t, view.Height, h)

					return uint64(totalPrecommitCount)
				},
				getSumVotingPowerFn: func(messages []Message) uint64 {
					return uint64(len(messages))
				},
				isValidatorFn: func(_ []byte) bool {
					return true
				},
			}

			mockNode = &mockNode{
				idFn: func() []byte {
					return id
				},
			}
			mockSigner = &mockSigner{
				signFn: func(b []byte) []byte {
					require.NotNil(t, b)

					return signature
				},
				isValidSignatureFn: func(_ []byte, _ []byte) bool {
					return true
				},
			}
		)

		// Create the tendermint instance
		tm := NewTendermint(
			mockVerifier,
			mockNode,
			&mockBroadcast{},
			mockSigner,
			WithPrecommitTimeout(timeout),
		)
		tm.state.setHeight(view.Height)
		tm.state.setRound(view.Round)
		tm.state.step = precommit
		tm.state.acceptedProposal = proposal
		tm.state.acceptedProposalID = proposalID

		// Add in non-NIL precommit messages
		for index, precommitMessage := range nonNilPrecommitMessages {
			// Change the senders for the non-NIL precommit messages.
			// The reason the senders need to be changed is, so we can simulate the following scenario:
			// 1/2 of the total voting power sent in non-NIL precommit messages
			// 1/2 of the total voting power sent in NIL precommit messages
			// In turn, there is a super majority when their voting powers are summed (non-NIL and NIL)
			precommitMessage.Sender = []byte(fmt.Sprintf("sender %d", index))

			require.NoError(t, tm.AddPrecommitMessage(precommitMessage))
		}

		// Add in NIL precommit messages
		for index, precommitMessage := range nilPrecommitMessages {
			// Change the senders for the NIL precommit messages
			precommitMessage.Sender = []byte(fmt.Sprintf("sender %d", index+len(nonNilPrecommitMessages)))

			require.NoError(t, tm.AddPrecommitMessage(precommitMessage))
		}

		// Run through the states
		finalizedProposal := <-tm.finalizeProposal(ctx)

		cancelFn()
		tm.wg.Wait()

		// Make sure the finalized proposal is NIL
		assert.Nil(t, finalizedProposal)
	})
}

func TestTendermint_WatchForFutureRounds(t *testing.T) {
	t.Parallel()

	var (
		processView = &types.View{
			Height: 10,
			Round:  5,
		}
		view = &types.View{
			Height: processView.Height,
			Round:  processView.Round + 5, // higher round
		}

		totalMessagesPerType = 10
		prevoteMessages      = generatePrevoteMessages(t, totalMessagesPerType/4, view, []byte("proposal ID"))
		precommitMessages    = generatePrecommitMessages(t, totalMessagesPerType, view, []byte("proposal ID"))

		mockVerifier = &mockVerifier{
			getTotalVotingPowerFn: func(h uint64) uint64 {
				require.EqualValues(t, view.Height, h)

				return uint64(len(prevoteMessages) + len(precommitMessages))
			},
			getSumVotingPowerFn: func(messages []Message) uint64 {
				return uint64(len(messages))
			},
			isValidatorFn: func(_ []byte) bool {
				return true
			},
		}
		mockSigner = &mockSigner{
			signFn: func(b []byte) []byte {
				require.NotNil(t, b)

				return []byte("signature")
			},
			isValidSignatureFn: func(_ []byte, _ []byte) bool {
				return true
			},
		}
	)

	// Create the tendermint instance
	tm := NewTendermint(
		mockVerifier,
		nil,
		nil,
		mockSigner,
	)

	// Set the process view
	tm.state.setHeight(processView.GetHeight())
	tm.state.setRound(processView.GetRound())

	// Make sure F+1 messages are added to the
	// message queue, with a higher round
	for _, message := range prevoteMessages {
		require.NoError(t, tm.AddPrevoteMessage(message))
	}

	for _, message := range precommitMessages {
		require.NoError(t, tm.AddPrecommitMessage(message))
	}

	// Set up the wait context
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	// Wait for future round jumps
	nextRound := <-tm.watchForRoundJumps(ctx)

	tm.wg.Wait()

	// Make sure the correct round was returned
	assert.Equal(t, view.GetRound(), nextRound)
}
