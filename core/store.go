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
func newStore() store {
	return store{
		proposeMessages:   messages.NewCollector[types.ProposalMessage](),
		prevoteMessages:   messages.NewCollector[types.PrevoteMessage](),
		precommitMessages: messages.NewCollector[types.PrecommitMessage](),
	}
}

// AddProposalMessage adds a proposal message to the store
func (s *store) AddProposalMessage(proposal *types.ProposalMessage) {
	s.proposeMessages.AddMessage(proposal.View, proposal.Sender, proposal)
}

// AddPrevoteMessage adds a prevote message to the store
func (s *store) AddPrevoteMessage(prevote *types.PrevoteMessage) {
	s.prevoteMessages.AddMessage(prevote.View, prevote.Sender, prevote)
}

// AddPrecommitMessage adds a precommit message to the store
func (s *store) AddPrecommitMessage(precommit *types.PrecommitMessage) {
	s.precommitMessages.AddMessage(precommit.View, precommit.Sender, precommit)
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
