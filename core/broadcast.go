package core

import (
	"github.com/gnolang/go-tendermint/messages/types"
)

// buildProposalMessage builds a proposal message using the given proposal
func (t *Tendermint) buildProposalMessage(proposal []byte) *types.ProposalMessage {
	var (
		height     = t.state.LoadHeight()
		round      = t.state.LoadRound()
		validRound = t.state.LoadValidRound()
	)

	// Build the proposal message (assumes the node will sign it)
	message := &types.ProposalMessage{
		View: &types.View{
			Height: height,
			Round:  round,
		},
		Sender:        t.node.ID(),
		Proposal:      proposal,
		ProposalRound: validRound,
	}

	// Sign the message
	message.Signature = t.signer.Sign(message.GetSignaturePayload())

	return message
}

// buildPrevoteMessage builds a prevote message using the given proposal identifier
func (t *Tendermint) buildPrevoteMessage(id []byte) *types.PrevoteMessage {
	var (
		height = t.state.LoadHeight()
		round  = t.state.LoadRound()

		processID = t.node.ID()
	)

	message := &types.PrevoteMessage{
		View: &types.View{
			Height: height,
			Round:  round,
		},
		Sender:     processID,
		Identifier: id,
	}

	// Sign the message
	message.Signature = t.signer.Sign(message.GetSignaturePayload())

	return message
}

// buildPrecommitMessage builds a precommit message using the given precommit identifier
//
//nolint:unused // Temporarily unused
func (t *Tendermint) buildPrecommitMessage(id []byte) *types.PrecommitMessage {
	var (
		height = t.state.LoadHeight()
		round  = t.state.LoadRound()

		processID = t.node.ID()
	)

	message := &types.PrecommitMessage{
		View: &types.View{
			Height: height,
			Round:  round,
		},
		Sender:     processID,
		Identifier: id,
	}

	// Sign the message
	message.Signature = t.signer.Sign(message.GetSignaturePayload())

	return message
}
