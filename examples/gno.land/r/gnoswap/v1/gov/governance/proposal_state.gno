package governance

import (
	"gno.land/p/demo/ufmt"
)

// ProposalStatus represents the current status of a proposal
type ProposalStatus int

const (
	_ ProposalStatus = iota
	StatusUpcoming
	StatusActive
	StatusPassed
	StatusRejected
	StatusExecuted
	StatusExpired
	StatusCanceled
)

// String returns the string representation of ProposalStatus
func (s ProposalStatus) String() string {
	switch s {
	case StatusUpcoming:
		return "upcoming"
	case StatusActive:
		return "active"
	case StatusPassed:
		return "passed"
	case StatusRejected:
		return "rejected"
	case StatusExecuted:
		return "executed"
	case StatusExpired:
		return "expired"
	case StatusCanceled:
		return "canceled"
	default:
		return "unknown"
	}
}

// IsTerminated returns true if the proposal is in a terminated state
func (s ProposalStatus) IsTerminated() bool {
	switch s {
	case StatusExecuted, StatusRejected, StatusCanceled, StatusExpired:
		return true
	default:
		return false
	}
}

// CanTransitionTo checks if transition to target status is allowed
func (s ProposalStatus) CanTransitionTo(target ProposalStatus) bool {
	validTransitions := map[ProposalStatus][]ProposalStatus{
		StatusUpcoming: {StatusActive, StatusCanceled},
		StatusActive:   {StatusPassed, StatusRejected, StatusCanceled},
		StatusPassed:   {StatusExecuted, StatusExpired, StatusCanceled},
		StatusRejected: {}, // Terminal state
		StatusExecuted: {}, // Terminal state
		StatusExpired:  {}, // Terminal state
		StatusCanceled: {}, // Terminal state
	}

	allowedTargets := validTransitions[s]
	for _, allowed := range allowedTargets {
		if allowed == target {
			return true
		}
	}
	return false
}

// ProposalPhase represents the current phase of a proposal lifecycle
type ProposalPhase int

const (
	PhaseCreated ProposalPhase = iota
	PhaseVoting
	PhaseVotingEnded
	PhaseExecution
	PhaseExecutionEnded
)

// proposalStateUpdater handles the state transitions of a proposal.
type proposalStateUpdater struct {
	proposal *ProposalInfo
	config   Config
	now      uint64
}

// StateTransition represents a single state transition with its conditions and actions
type StateTransition struct {
	name      string
	from      ProposalStatus
	to        ProposalStatus
	condition func() bool
	action    func()
}

// newProposalStateUpdater creates a new proposalStateUpdater.
func newProposalStateUpdater(proposal *ProposalInfo, now uint64) *proposalStateUpdater {
	return &proposalStateUpdater{
		proposal: proposal,
		config:   GetConfigVersion(proposal.ConfigVersion),
		now:      now,
	}
}

// getCurrentStatus returns the current status based on the proposal state
func (u *proposalStateUpdater) getCurrentStatus() ProposalStatus {
	state := u.proposal.State

	// Check terminal states first
	if state.Executed {
		return StatusExecuted
	}
	if state.Expired {
		return StatusExpired
	}
	if state.Canceled {
		return StatusCanceled
	}
	if state.Rejected {
		return StatusRejected
	}

	// Check active states
	if state.Passed {
		return StatusPassed
	}
	if state.Active {
		return StatusActive
	}
	if state.Upcoming {
		return StatusUpcoming
	}

	// Fallback (shouldn't happen with proper state management)
	return StatusUpcoming
}

// getCurrentPhase returns the current phase based on time
func (u *proposalStateUpdater) getCurrentPhase() ProposalPhase {
	votingStart, votingEnd := u.getVotingPeriod()
	executionStart, executionEnd := u.getExecutionPeriod()

	switch {
	case u.now < votingStart:
		return PhaseCreated
	case u.now >= votingStart && u.now <= votingEnd:
		return PhaseVoting
	case u.now > votingEnd && u.now < executionStart:
		return PhaseVotingEnded
	case u.now >= executionStart && u.now <= executionEnd:
		return PhaseExecution
	default:
		return PhaseExecutionEnded
	}
}

// shouldUpdate determines if the proposal state should be updated.
func (u *proposalStateUpdater) shouldUpdate() bool {
	currentStatus := u.getCurrentStatus()
	return !currentStatus.IsTerminated()
}

// update processes all possible state transitions in order
func (u *proposalStateUpdater) update() {
	transitions := u.getStateTransitions()
	currentStatus := u.getCurrentStatus()

	for _, transition := range transitions {
		if transition.from == currentStatus && transition.condition() {
			// Validate transition is allowed
			if !currentStatus.CanTransitionTo(transition.to) {
				panic(ufmt.Sprintf("Invalid state transition from %s to %s",
					currentStatus.String(), transition.to.String()))
			}
			transition.action()
			break // Only one transition per update cycle
		}
	}
}

// getStateTransitions returns all possible state transitions
func (u *proposalStateUpdater) getStateTransitions() []StateTransition {
	return []StateTransition{
		{
			name:      "upcoming_to_active",
			from:      StatusUpcoming,
			to:        StatusActive,
			condition: u.isVotingPhase,
			action:    u.startVoting,
		},
		{
			name:      "active_to_passed",
			from:      StatusActive,
			to:        StatusPassed,
			condition: u.isVotingEndedAndPassed,
			action:    u.markAsPassed,
		},
		{
			name:      "active_to_rejected",
			from:      StatusActive,
			to:        StatusRejected,
			condition: u.isVotingEndedAndRejected,
			action:    u.markAsRejected,
		},
		{
			name:      "passed_to_expired",
			from:      StatusPassed,
			to:        StatusExpired,
			condition: u.isExecutionExpired,
			action:    u.markAsExpired,
		},
	}
}

// Time period calculations
func (u *proposalStateUpdater) getVotingPeriod() (start, end uint64) {
	start = u.proposal.State.CreatedAt + u.config.VotingStartDelay
	end = start + u.config.VotingPeriod
	return
}

func (u *proposalStateUpdater) getExecutionPeriod() (start, end uint64) {
	_, votingEnd := u.getVotingPeriod()
	start = votingEnd + u.config.ExecutionDelay
	end = start + u.config.ExecutionWindow
	return
}

// Phase-based condition checks (safer than boolean logic)
func (u *proposalStateUpdater) isVotingPhase() bool {
	return u.getCurrentPhase() == PhaseVoting
}

func (u *proposalStateUpdater) isVotingEndedAndPassed() bool {
	phase := u.getCurrentPhase()
	if phase != PhaseVotingEnded && phase != PhaseExecution && phase != PhaseExecutionEnded {
		return false
	}

	yeaVotes := u.proposal.Yea.Uint64()
	nayVotes := u.proposal.Nay.Uint64()
	return u.isVotingPassed(yeaVotes, nayVotes)
}

func (u *proposalStateUpdater) isVotingEndedAndRejected() bool {
	phase := u.getCurrentPhase()
	if phase != PhaseVotingEnded && phase != PhaseExecution && phase != PhaseExecutionEnded {
		return false
	}

	yeaVotes := u.proposal.Yea.Uint64()
	nayVotes := u.proposal.Nay.Uint64()
	return !u.isVotingPassed(yeaVotes, nayVotes)
}

func (u *proposalStateUpdater) isExecutionExpired() bool {
	// Only non-text proposals can expire
	if u.proposal.ProposalType == Text {
		return false
	}

	return u.getCurrentPhase() == PhaseExecutionEnded
}

// State transition actions
func (u *proposalStateUpdater) startVoting() {
	u.proposal.State.Upcoming = false
	u.proposal.State.Active = true
}

func (u *proposalStateUpdater) markAsPassed() {
	u.proposal.State.Passed = true
	u.proposal.State.PassedAt = u.now
	u.proposal.State.Upcoming = false
	u.proposal.State.Active = false
}

func (u *proposalStateUpdater) markAsRejected() {
	u.proposal.State.Rejected = true
	u.proposal.State.RejectedAt = u.now
	u.proposal.State.Upcoming = false
	u.proposal.State.Active = false
}

func (u *proposalStateUpdater) markAsExpired() {
	u.proposal.State.Expired = true
	u.proposal.State.ExpiredAt = u.now
}

// Helper methods
func (u *proposalStateUpdater) isVotingPassed(yeaVotes, nayVotes uint64) bool {
	return yeaVotes >= u.proposal.QuorumAmount && yeaVotes > nayVotes
}

// isProposalTerminated checks if a proposal is in a terminated state (legacy compatibility)
func isProposalTerminated(proposal ProposalInfo) bool {
	updater := &proposalStateUpdater{proposal: &proposal}
	return updater.getCurrentStatus().IsTerminated()
}
