package core

import "github.com/gnolang/go-tendermint/messages/types"

// step is the current state step
type step uint8

const (
	propose step = iota
	prevote
	precommit
)

func (n step) String() string {
	switch n {
	case propose:
		return "propose"
	case prevote:
		return "prevote"
	case precommit:
		return "precommit"
	}

	return ""
}

// state holds information about the current consensus state
// TODO make thread safe
type state struct {
	view *types.View
	step step

	acceptedProposal   *types.ProposalMessage
	acceptedProposalID []byte

	lockedValue []byte
	lockedRound int64

	validValue []byte
	validRound int64
}

// newState creates a fresh state using the given view
func newState(view *types.View) *state {
	return &state{
		view:               view,
		step:               propose,
		acceptedProposal:   nil,
		acceptedProposalID: nil,
		lockedValue:        nil,
		lockedRound:        -1,
		validValue:         nil,
		validRound:         -1,
	}
}
