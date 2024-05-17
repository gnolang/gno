package poa

import "gno.land/r/sys/vals/types"

type Option func(*PoA)

// WithInitialSet sets the initial PoA validator set
func WithInitialSet(validators []*types.Validator) Option {
	return func(p *PoA) {
		for index, validator := range validators {
			p.validators = append(p.validators, validator)
			p.addressToValidatorIndex[validator.Address] = index

			p.totalVotingPower += validator.VotingPower
		}

		p.majorityPower = (2 * p.totalVotingPower) / 3
	}
}
