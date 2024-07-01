package poc

import "gno.land/p/sys/vals/types"

type Option func(*PoC)

// WithInitialSet sets the initial PoA validator set
func WithInitialSet(validators []*types.Validator) Option {
	return func(p *PoC) {
		for _, validator := range validators {
			p.validators.Set(validator.Address.String(), validator)
		}
	}
}
