package vls

import (
	"std"
	"strconv"
)

const (
	DelegateChangedEvent    = "DelegateChanged"
	VotingPowerChangedEvent = "VotingPowerChanged"
	GovernanceUpdatedEvent  = "GovernanceUpdated"
)

const (
	EventDelegatorKey = "delegator"
	EventFromKey      = "from"
	EventToKey        = "to"
	EventDelegateeKey = "delegatee"
	EventVotesKey     = "votes"
	EventOldGovKey    = "old_governance"
	EventNewGovKey    = "new_governance"
)

func emitDelegateChanged(delegator, from, to string) {
	std.Emit(
		DelegateChangedEvent,
		EventDelegatorKey, delegator,
		EventFromKey, from,
		EventToKey, to,
	)
}

func emitVotingPowerChanged(delegatee string, votes int64) {
	std.Emit(
		VotingPowerChangedEvent,
		EventDelegateeKey, delegatee,
		EventVotesKey, strconv.Itoa(int(votes)),
	)
}

func emitGovernanceUpdated(oldGov, newGov string) {
	std.Emit(
		GovernanceUpdatedEvent,
		EventOldGovKey, oldGov,
		EventNewGovKey, newGov,
	)
}
