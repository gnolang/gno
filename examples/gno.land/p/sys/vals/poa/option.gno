package poa

import "gno.land/p/sys/vals/types"

type Option func(*PoA)

// WithInitialSet sets the initial PoA validator set
func WithInitialSet(validators []types.Validator) Option {
	return func(p *PoA) {
		for _, validator := range validators {
			p.validators.Set(validator.Address.String(), validator)
		}
	}
}
