package poa

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/sys/validators"
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
	validators *avl.Tree // std.Address -> validators.Validator
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

func (p *PoA) AddValidator(address std.Address, pubKey string, power uint64) (validators.Validator, error) {
	// Validate that the operation is a valid call.
	// Check if the validator is already in the set
	if p.IsValidator(address) {
		return validators.Validator{}, validators.ErrValidatorExists
	}

	// Make sure the voting power > 0
	if power == 0 {
		return validators.Validator{}, ErrInvalidVotingPower
	}

	v := validators.Validator{
		Address:     address,
		PubKey:      pubKey, // TODO: in the future, verify the public key
		VotingPower: power,
	}

	// Add the validator to the set
	p.validators.Set(address.String(), v)

	return v, nil
}

func (p *PoA) RemoveValidator(address std.Address) (validators.Validator, error) {
	// Validate that the operation is a valid call
	// Fetch the validator
	validator, err := p.GetValidator(address)
	if err != nil {
		return validators.Validator{}, err
	}

	// Remove the validator from the set
	p.validators.Remove(address.String())

	return validator, nil
}

func (p *PoA) IsValidator(address std.Address) bool {
	_, exists := p.validators.Get(address.String())

	return exists
}

func (p *PoA) GetValidator(address std.Address) (validators.Validator, error) {
	validatorRaw, exists := p.validators.Get(address.String())
	if !exists {
		return validators.Validator{}, validators.ErrValidatorMissing
	}

	validator := validatorRaw.(validators.Validator)

	return validator, nil
}

func (p *PoA) GetValidators() []validators.Validator {
	vals := make([]validators.Validator, 0, p.validators.Size())

	p.validators.Iterate("", "", func(_ string, value interface{}) bool {
		validator := value.(validators.Validator)
		vals = append(vals, validator)

		return false
	})

	return vals
}
