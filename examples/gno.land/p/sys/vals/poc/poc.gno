package poc

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/sys/vals/types"
)

const errInvalidVotingPower = "invalid voting power"

// PoC specifies the Proof of Contribution validator set.
// In order to become part of the set, users to be voted into
// the validator set by a govdao proposal
type PoC struct {
	validators *avl.Tree // std.Address -> *types.Validator
}

// NewPoC creates a new empty Proof of Contribution validator set
func NewPoC(opts ...Option) *PoC {
	// Create the empty set
	p := &PoC{
		validators: avl.NewTree(),
	}

	// Apply the options
	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *PoC) AddValidator(address std.Address, pubKey string, power uint64) *types.Validator {
	// Validate that the operation is a valid call.
	// Check if the validator is already in the set
	if p.IsValidator(address) {
		panic(types.ErrValidatorExists)
	}

	// Make sure the voting power > 0
	if power == 0 {
		panic(errInvalidVotingPower)
	}

	v := &types.Validator{
		Address:     address,
		PubKey:      pubKey, // TODO: in the future, verify the public key
		VotingPower: power,
	}

	// Add the validator to the set
	p.validators.Set(address.String(), v)

	return v
}

func (p *PoC) RemoveValidator(address std.Address) *types.Validator {
	// Validate that the operation is a valid call
	// Check if the address is a validator
	if !p.IsValidator(address) {
		panic(types.ErrValidatorMissing)
	}

	// Fetch the validator
	validator := p.GetValidator(address)

	// Remove the validator from the set
	p.validators.Remove(address.String())

	return &validator
}

func (p *PoC) IsValidator(address std.Address) bool {
	_, exists := p.validators.Get(address.String())

	return exists
}

func (p *PoC) GetValidator(address std.Address) types.Validator {
	validatorRaw, exists := p.validators.Get(address.String())
	if !exists {
		panic(types.ErrValidatorMissing)
	}

	return *(validatorRaw.(*types.Validator))
}

func (p *PoC) GetValidators() []*types.Validator {
	validators := make([]*types.Validator, 0, p.validators.Size())

	p.validators.Iterate("", "", func(_ string, value interface{}) bool {
		validator := value.(*types.Validator)
		validators = append(validators, validator)

		return false
	})

	return validators
}
