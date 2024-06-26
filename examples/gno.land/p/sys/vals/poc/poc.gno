package poc

import (
	"std"

	"gno.land/p/sys/vals/types"
)

const errInvalidVotingPower = "invalid voting power"

// PoC specifies the Proof of Contribution validator set.
// In order to become part of the set, users to be voted into
// the validator set by a govdao proposal
type PoC struct {
	// validators holds the current validator set.
	// This slice can never practically grow more than ~150 elements,
	// due to Tendermint's quadratic network complexity
	validators              []*types.Validator
	addressToValidatorIndex map[std.Address]int // address -> index
}

// NewPoC creates a new empty Proof of Contribution validator set
func NewPoC(opts ...Option) *PoC {
	// Create the empty set
	p := &PoC{
		validators:              make([]*types.Validator, 0),
		addressToValidatorIndex: make(map[std.Address]int),
	}

	// Apply the options
	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *PoC) AddValidator(address std.Address, pubKey string, power uint64) *types.Validator {
	// Validate that the operation is a valid call
	p.validateAdd(address, power)

	v := &types.Validator{
		Address:     address,
		PubKey:      pubKey, // TODO: in the future, verify the public key
		VotingPower: power,
	}

	// Add the validator to the set
	p.addressToValidatorIndex[v.Address] = len(p.validators)
	p.validators = append(p.validators, v)

	return v
}

// validateAdd validates a validator add call
func (p *PoC) validateAdd(address std.Address, power uint64) {
	// Check if the validator is already in the set
	if p.IsValidator(address) {
		panic(types.ErrValidatorExists)
	}

	// Make sure the voting power > 0
	if power == 0 {
		panic(errInvalidVotingPower)
	}
}

func (p *PoC) RemoveValidator(address std.Address) *types.Validator {
	// Validate that the operation is a valid call
	p.validateRemove(address)

	// Fetch the validator index
	index := p.addressToValidatorIndex[address]

	// Remove the validator from the set
	validator := p.validators[index]
	p.validators = append(p.validators[:index], p.validators[index+1:]...)

	delete(p.addressToValidatorIndex, address)

	return validator
}

// validateRemove validates a validator remove call
func (p *PoC) validateRemove(address std.Address) {
	// Check if the address is a validator
	if !p.IsValidator(address) {
		panic(types.ErrValidatorMissing)
	}
}

func (p *PoC) IsValidator(address std.Address) bool {
	_, exists := p.addressToValidatorIndex[address]

	return exists
}

func (p *PoC) GetValidator(address std.Address) types.Validator {
	valIndex, exists := p.addressToValidatorIndex[address]
	if !exists {
		panic(types.ErrValidatorMissing)
	}

	return *p.validators[valIndex]
}

func (p *PoC) GetValidators() []*types.Validator {
	return p.validators
}
