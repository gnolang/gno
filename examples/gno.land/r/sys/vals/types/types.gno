package types

import (
	"std"
)

// Validator represents a single chain validator
type Validator struct {
	Address     std.Address // bech32 address
	PubKey      string      // bech32 representation of the public key
	VotingPower uint64
}

const (
	ValidatorAddedEvent   = "ValidatorAdded"   // emitted when a validator was added to the set
	ValidatorRemovedEvent = "ValidatorRemoved" // emitted when a validator was removed from the set
)
