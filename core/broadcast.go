package core

import (
	"github.com/gnolang/go-tendermint/messages/types"
)

// buildProposalMessage builds a proposal message using the given proposal
func (t *Tendermint) buildProposalMessage(proposal []byte) *types.ProposalMessage {
	// TODO make thread safe
	var (
		height     = t.state.view.Height
		round      = t.state.view.Round
		validRound = t.state.validRound
	)

	// Build the proposal message (assumes the node will sign it)
	return &types.ProposalMessage{
		View: &types.View{
			Height: height,
			Round:  round,
		},
		From:          t.node.ID(),
		Proposal:      proposal,
		ProposalRound: validRound,
	}
}

// buildPrevoteMessage builds a prevote message using the given proposal identifier
func (t *Tendermint) buildPrevoteMessage(id []byte) *types.PrevoteMessage {
	// TODO make thread safe
	var (
		height = t.state.view.Height
		round  = t.state.view.Round

		processID = t.node.ID()
	)

	return &types.PrevoteMessage{
		View: &types.View{
			Height: height,
			Round:  round,
		},
		From:       processID,
		Identifier: id,
	}
}

// buildPrecommitMessage builds a precommit message using the given precommit identifier
//
//nolint:unused // Temporarily unused
func (t *Tendermint) buildPrecommitMessage(id []byte) *types.PrecommitMessage {
	// TODO make thread safe
	var (
		height = t.state.view.Height
		round  = t.state.view.Round

		processID = t.node.ID()
	)

	return &types.PrecommitMessage{
		View: &types.View{
			Height: height,
			Round:  round,
		},
		From:       processID,
		Identifier: id,
	}
}

// broadcastProposal signs and broadcasts the given proposal message
func (t *Tendermint) broadcastProposal(proposal *types.ProposalMessage) {
	message := &types.Message{
		Type:      types.MessageType_PROPOSAL,
		Signature: t.signer.Sign(proposal.Marshal()),
		Payload: &types.Message_ProposalMessage{
			ProposalMessage: proposal,
		},
	}

	// Broadcast the proposal message
	t.broadcast.Broadcast(message)
}

// broadcastPrevote signs and broadcasts the given prevote message
func (t *Tendermint) broadcastPrevote(prevote *types.PrevoteMessage) {
	message := &types.Message{
		Type:      types.MessageType_PREVOTE,
		Signature: t.signer.Sign(prevote.Marshal()),
		Payload: &types.Message_PrevoteMessage{
			PrevoteMessage: prevote,
		},
	}

	// Broadcast the prevote message
	t.broadcast.Broadcast(message)
}

// broadcastPrecommit signs and broadcasts the given precommit message
//
//nolint:unused // Temporarily unused
func (t *Tendermint) broadcastPrecommit(precommit *types.PrecommitMessage) {
	message := &types.Message{
		Type:      types.MessageType_PRECOMMIT,
		Signature: t.signer.Sign(precommit.Marshal()),
		Payload: &types.Message_PrecommitMessage{
			PrecommitMessage: precommit,
		},
	}

	// Broadcast the precommit message
	t.broadcast.Broadcast(message)
}
