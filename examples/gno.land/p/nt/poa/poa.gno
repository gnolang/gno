package poa

import (
	"errors"

	"gno.land/p/nt/avl"
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
	validators *avl.Tree // address -> validators.Validator
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

func (p *PoA) AddValidator(address_XXX address, pubKey string, power uint64) (validators.Validator, error) {
	// Make sure the voting power > 0
	if power == 0 {
		return validators.Validator{}, ErrInvalidVotingPower
	}

	// If the validator already exists, update only if the voting power changed
	if existing, err := p.GetValidator(address_XXX); err == nil {
		if existing.VotingPower == power {
			return validators.Validator{}, validators.ErrValidatorExists
		}
	}

	v := validators.Validator{
		Address:     address_XXX,
		PubKey:      pubKey, // TODO: in the future, verify the public key
		VotingPower: power,
	}

	// Add or update the validator in the set
	p.validators.Set(address_XXX.String(), v)

	return v, nil
}

func (p *PoA) RemoveValidator(address_XXX address) (validators.Validator, error) {
	// Validate that the operation is a valid call
	// Fetch the validator
	validator, err := p.GetValidator(address_XXX)
	if err != nil {
		return validators.Validator{}, err
	}

	// Remove the validator from the set
	p.validators.Remove(address_XXX.String())

	return validator, nil
}

func (p *PoA) IsValidator(address_XXX address) bool {
	_, exists := p.validators.Get(address_XXX.String())

	return exists
}

func (p *PoA) GetValidator(address_XXX address) (validators.Validator, error) {
	validatorRaw, exists := p.validators.Get(address_XXX.String())
	if !exists {
		return validators.Validator{}, validators.ErrValidatorMissing
	}

	validator := validatorRaw.(validators.Validator)

	return validator, nil
}

func (p *PoA) GetValidators() []validators.Validator {
	vals := make([]validators.Validator, 0, p.validators.Size())

	p.validators.Iterate("", "", func(_ string, value any) bool {
		validator := value.(validators.Validator)
		vals = append(vals, validator)

		return false
	})

	return vals
}
