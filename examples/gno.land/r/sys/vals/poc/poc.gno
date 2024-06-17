package poc

import (
	"std"

	"gno.land/r/sys/vals/types"
)

const govdaoRealm = "gno.land/r/gov/dao"

const errNotGovdao = "call not executed by govdao"

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

func (p *PoC) AddValidator(address std.Address, pubKey string) *types.Validator {
	// Validate that the operation is a valid call
	p.validateAdd(address)

	v := &types.Validator{
		Address:     address,
		PubKey:      pubKey, // TODO: in the future, verify the public key
		VotingPower: 1,      // in this PoC system, all new validators have the same voting power
	}

	// Add the validator to the set
	p.addressToValidatorIndex[v.Address] = len(p.validators)
	p.validators = append(p.validators, v)

	// Emit the validator set change
	std.Emit(types.ValidatorAddedEvent)

	return v
}

// validateAdd validates a validator add call
func (p *PoC) validateAdd(address std.Address) {
	if std.PrevRealm().PkgPath() != govdaoRealm {
		println(std.PrevRealm())
		panic(errNotGovdao)
	}

	// Check if the validator is already in the set
	if p.IsValidator(address) {
		panic(types.ErrValidatorExists)
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

	// Emit the validator set change
	std.Emit(types.ValidatorRemovedEvent)

	return validator
}

// validateRemove validates a validator remove call
func (p *PoC) validateRemove(address std.Address) {
	// Check if the method call is from a proposal execution
	if std.PrevRealm().PkgPath() != govdaoRealm {
		panic(errNotGovdao)
	}

	// Check if the address is a validator
	if !p.IsValidator(address) {
		panic(types.ErrValidatorMissing)
	}
}

func (p *PoC) IsValidator(address std.Address) bool {
	_, exists := p.addressToValidatorIndex[address]

	return exists
}

func (p *PoC) GetValidators() []*types.Validator {
	return p.validators
}
