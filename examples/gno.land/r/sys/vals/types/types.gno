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

const (
	// ErrCallerNotUserAccount is returned when the caller is not a user account
	ErrCallerNotUserAccount = "caller not user account"

	// ErrValidatorExists is returned when the validator is already in the set
	ErrValidatorExists = "validator already exists"

	// ErrValidatorMissing is returned when the validator is not in the set
	ErrValidatorMissing = "validator doesn't exist"
)
