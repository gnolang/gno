package poa

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/sys/vals"
)

var ErrInvalidVotingPower = errors.New("invalid voting power")

// PoA specifies the Proof of Authority validator set, with simple add / remove constraints.
//
// To add:
// - proposed validator must not be part of the set already
// - proposed validator voting power must be > 0
//
// To remove:
// - proposed validator must be part of the set already
type PoA struct {
	validators *avl.Tree // std.Address -> vals.Validator
}

// NewPoA creates a new empty Proof of Authority validator set
func NewPoA(opts ...Option) *PoA {
	// Create the empty set
	p := &PoA{
		validators: avl.NewTree(),
	}

	// Apply the options
	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *PoA) AddValidator(address std.Address, pubKey string, power uint64) (*vals.Validator, error) {
	// Validate that the operation is a valid call.
	// Check if the validator is already in the set
	if p.IsValidator(address) {
		return nil, vals.ErrValidatorExists
	}

	// Make sure the voting power > 0
	if power == 0 {
		return nil, ErrInvalidVotingPower
	}

	v := vals.Validator{
		Address:     address,
		PubKey:      pubKey, // TODO: in the future, verify the public key
		VotingPower: power,
	}

	// Add the validator to the set
	p.validators.Set(address.String(), v)

	return &v, nil
}

func (p *PoA) RemoveValidator(address std.Address) (*vals.Validator, error) {
	// Validate that the operation is a valid call
	// Fetch the validator
	validator, err := p.GetValidator(address)
	if err != nil {
		return nil, err
	}

	// Remove the validator from the set
	p.validators.Remove(address.String())

	return validator, nil
}

func (p *PoA) IsValidator(address std.Address) bool {
	_, exists := p.validators.Get(address.String())

	return exists
}

func (p *PoA) GetValidator(address std.Address) (*vals.Validator, error) {
	validatorRaw, exists := p.validators.Get(address.String())
	if !exists {
		return nil, vals.ErrValidatorMissing
	}

	validator := validatorRaw.(vals.Validator)

	return &validator, nil
}

func (p *PoA) GetValidators() []vals.Validator {
	validators := make([]vals.Validator, 0, p.validators.Size())

	p.validators.Iterate("", "", func(_ string, value interface{}) bool {
		validator := value.(vals.Validator)
		validators = append(validators, validator)

		return false
	})

	return validators
}
