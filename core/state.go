package core

import (
	"sync/atomic"

	"github.com/gnolang/go-tendermint/messages/types"
)

// step is the current state step
type step uint32

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

func (n *step) Set(newStep step) {
	atomic.SwapUint32((*uint32)(n), uint32(newStep))
}

func (n *step) Load() step {
	s := atomic.LoadUint32((*uint32)(n))

	return step(s)
}

// state holds information about the current consensus state
// TODO make thread safe
type state struct {
	view *types.View

	acceptedProposal   *types.ProposalMessage
	acceptedProposalID []byte

	lockedValue []byte

	// no concurrent writes/reads
	// no need sync primitive
	// used in startRound()
	validValue []byte

	lockedRound int64
	validRound  int64

	step step
}

func (s *state) LoadHeight() uint64 {
	return atomic.LoadUint64(&s.view.Height)
}

func (s *state) LoadRound() uint64 {
	return atomic.LoadUint64(&s.view.Round)
}

func (s *state) LoadValidRound() int64 {
	return atomic.LoadInt64(&s.validRound)
}

func (s *state) LoadLockedRound() int64 {
	return atomic.LoadInt64(&s.lockedRound)
}

func (s *state) IncRound() {
	atomic.AddUint64(&s.view.Round, 1)
}

func (s *state) SetRound(r uint64) {
	atomic.SwapUint64(&s.view.Round, r)
}

// newState creates a fresh state using the given view
func newState(view *types.View) state {
	return state{
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
