package core

import (
	"github.com/gnolang/go-tendermint/messages"
	"github.com/gnolang/go-tendermint/messages/types"
)

// store is the message store
type store struct {
	proposeMessages   *messages.Collector[types.ProposalMessage]
	prevoteMessages   *messages.Collector[types.PrevoteMessage]
	precommitMessages *messages.Collector[types.PrecommitMessage]
}

// newStore creates a new message store
func newStore() *store {
	return &store{
		proposeMessages:   messages.NewCollector[types.ProposalMessage](),
		prevoteMessages:   messages.NewCollector[types.PrevoteMessage](),
		precommitMessages: messages.NewCollector[types.PrecommitMessage](),
	}
}

// AddMessage adds a new message to the store
func (s *store) AddMessage(message *types.Message) {
	switch message.Type {
	case types.MessageType_PROPOSAL:
		// Parse the propose message
		wrappedMessage, ok := message.Payload.(*types.Message_ProposalMessage)
		if !ok {
			return
		}

		// Get the proposal
		proposal := wrappedMessage.ProposalMessage

		s.proposeMessages.AddMessage(proposal.View, proposal.From, proposal)
	case types.MessageType_PREVOTE:
		// Parse the prevote message
		wrappedMessage, ok := message.Payload.(*types.Message_PrevoteMessage)
		if !ok {
			return
		}

		// Get the prevote
		prevote := wrappedMessage.PrevoteMessage

		s.prevoteMessages.AddMessage(prevote.View, prevote.From, prevote)
	case types.MessageType_PRECOMMIT:
		// Parse the precommit message
		wrappedMessage, ok := message.Payload.(*types.Message_PrecommitMessage)
		if !ok {
			return
		}

		// Get the precommit
		precommit := wrappedMessage.PrecommitMessage

		s.precommitMessages.AddMessage(precommit.View, precommit.From, precommit)
	}
}

// SubscribeToPropose subscribes to incoming PROPOSE messages
func (s *store) SubscribeToPropose() (<-chan func() []*types.ProposalMessage, func()) {
	return s.proposeMessages.Subscribe()
}

// SubscribeToPrevote subscribes to incoming PREVOTE messages
func (s *store) SubscribeToPrevote() (<-chan func() []*types.PrevoteMessage, func()) {
	return s.prevoteMessages.Subscribe()
}

// SubscribeToPrecommit subscribes to incoming PRECOMMIT messages
func (s *store) SubscribeToPrecommit() (<-chan func() []*types.PrecommitMessage, func()) {
	return s.precommitMessages.Subscribe()
}
