package vls

import (
	"std"
)

const (
	GovernanceUpdatedEvent = "GovernanceUpdated"
)

const (
	EventOldGovKey = "old_governance"
	EventNewGovKey = "new_governance"
)

func emitGovernanceUpdated(oldGov, newGov string) {
	std.Emit(
		GovernanceUpdatedEvent,
		EventOldGovKey, oldGov,
		EventNewGovKey, newGov,
	)
}
