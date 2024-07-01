package validators

import (
	"errors"
	"std"
)

// ValsetProtocol defines the validator set protocol (PoA / PoS / PoC / ?)
type ValsetProtocol interface {
	// AddValidator adds a new validator to the validator set.
	// If the validator is already present, the method should error out
	//
	// TODO: This API is not ideal -- the address should be derived from
	// the public key, and not be passed in as such, but currently Gno
	// does not support crypto address derivation
	AddValidator(address std.Address, pubKey string, power uint64) (*Validator, error)

	// RemoveValidator removes the given validator from the set.
	// If the validator is not present in the set, the method should error out
	RemoveValidator(address std.Address) (*Validator, error)

	// IsValidator returns a flag indicating if the given
	// bech32 address is part of the validator set
	IsValidator(address std.Address) bool

	// GetValidator returns the validator using the given address
	GetValidator(address std.Address) (*Validator, error)

	// GetValidators returns the currently active validator set
	GetValidators() []Validator
}

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

var (
	// ErrValidatorExists is returned when the validator is already in the set
	ErrValidatorExists = errors.New("validator already exists")

	// ErrValidatorMissing is returned when the validator is not in the set
	ErrValidatorMissing = errors.New("validator doesn't exist")
)
