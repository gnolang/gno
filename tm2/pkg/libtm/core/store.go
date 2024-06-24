package core

import (
	"github.com/gnolang/libtm/messages"
	"github.com/gnolang/libtm/messages/types"
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

// addProposalMessage adds a proposal message to the store
func (s *store) addProposalMessage(proposal *types.ProposalMessage) {
	s.proposeMessages.AddMessage(proposal.View, proposal.Sender, proposal)
}

// addPrevoteMessage adds a prevote message to the store
func (s *store) addPrevoteMessage(prevote *types.PrevoteMessage) {
	s.prevoteMessages.AddMessage(prevote.View, prevote.Sender, prevote)
}

// addPrecommitMessage adds a precommit message to the store
func (s *store) addPrecommitMessage(precommit *types.PrecommitMessage) {
	s.precommitMessages.AddMessage(precommit.View, precommit.Sender, precommit)
}

// subscribeToPropose subscribes to incoming PROPOSE messages
func (s *store) subscribeToPropose() (<-chan func() []*types.ProposalMessage, func()) {
	return s.proposeMessages.Subscribe()
}

// subscribeToPrevote subscribes to incoming PREVOTE messages
func (s *store) subscribeToPrevote() (<-chan func() []*types.PrevoteMessage, func()) {
	return s.prevoteMessages.Subscribe()
}

// subscribeToPrecommit subscribes to incoming PRECOMMIT messages
func (s *store) subscribeToPrecommit() (<-chan func() []*types.PrecommitMessage, func()) {
	return s.precommitMessages.Subscribe()
}

// dropMessages drops all messages from the store that are
// less than the given view (earlier)
func (s *store) dropMessages(view *types.View) {
	// Clean up the propose messages
	s.proposeMessages.DropMessages(view)

	// Clean up the prevote messages
	s.prevoteMessages.DropMessages(view)

	// Clean up the precommit messages
	s.precommitMessages.DropMessages(view)
}
