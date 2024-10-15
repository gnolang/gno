package core

import (
	"sync/atomic"

	"github.com/gnolang/libtm/messages/types"
)

// step is the current state step
type step uint32

const (
	propose step = iota
	prevote
	precommit
)

// set updates the current step value [THREAD SAFE]
func (s *step) set(n step) {
	atomic.SwapUint32((*uint32)(s), uint32(n))
}

// get fetches the current step value [THREAD SAFE]
func (s *step) get() step {
	return step(atomic.LoadUint32((*uint32)(s)))
}

// state holds information about the current consensus state
type state struct {
	view *types.View

	acceptedProposal   []byte
	acceptedProposalID []byte

	lockedValue []byte
	validValue  []byte

	lockedRound int64
	validRound  int64

	step step
}

// newState creates a fresh state using the given view
func newState() state {
	return state{
		view: &types.View{
			Height: 0, // zero height
			Round:  0, // zero round
		},
		step:               propose,
		acceptedProposal:   nil,
		acceptedProposalID: nil,
		lockedValue:        nil,
		lockedRound:        -1,
		validValue:         nil,
		validRound:         -1,
	}
}

// getHeight fetches the current view height [THREAD SAFE]
func (s *state) getHeight() uint64 {
	return atomic.LoadUint64(&s.view.Height)
}

// getRound fetches the current view round [THREAD SAFE]
func (s *state) getRound() uint64 {
	return atomic.LoadUint64(&s.view.Round)
}

// increaseRound increases the current view round by 1 [THREAD SAFE]
func (s *state) increaseRound() {
	atomic.AddUint64(&s.view.Round, 1)
}

// setRound sets the current view round to the given value [THREAD SAFE]
func (s *state) setRound(r uint64) {
	atomic.SwapUint64(&s.view.Round, r)
}

// setHeight sets the current view height to the given value [THREAD SAFE]
func (s *state) setHeight(h uint64) {
	atomic.SwapUint64(&s.view.Height, h)
}
