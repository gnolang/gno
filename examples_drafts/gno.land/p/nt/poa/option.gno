package poa

import "gno.land/p/sys/validators"

type Option func(*PoA)

// WithInitialSet sets the initial PoA validator set
func WithInitialSet(validators []validators.Validator) Option {
	return func(p *PoA) {
		for _, validator := range validators {
			p.validators.Set(validator.Address.String(), validator)
		}
	}
}
