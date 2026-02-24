package validators

import (
	"errors"
)

// ValsetProtocol defines the validator set protocol (PoA / PoS / PoC / ?)
type ValsetProtocol interface {
	// AddValidator adds a new validator to the validator set.
	// If the validator is already present *AND* has the same voting power,
	// the method should error out, otherwise update the validator's voting power
	//
	// TODO: This API is not ideal -- the address should be derived from
	// the public key, and not be passed in as such, but currently Gno
	// does not support crypto address derivation
	AddValidator(address_XXX address, pubKey string, power uint64) (Validator, error)

	// RemoveValidator removes the given validator from the set.
	// If the validator is not present in the set, the method should error out
	RemoveValidator(address_XXX address) (Validator, error)

	// IsValidator returns a flag indicating if the given
	// bech32 address is part of the validator set
	IsValidator(address_XXX address) bool

	// GetValidator returns the validator using the given address
	GetValidator(address_XXX address) (Validator, error)

	// GetValidators returns the currently active validator set
	GetValidators() []Validator
}

// Validator represents a single chain validator
type Validator struct {
	Address     address // bech32 address
	PubKey      string  // bech32 representation of the public key
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
