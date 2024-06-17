package pos

import (
	"std"

	"gno.land/p/sys/vals/types"
)

type Option func(*PoS)

// WithInitialSet sets the initial PoS validator set, with empty stakes
func WithInitialSet(validators []*types.Validator) Option {
	return func(p *PoS) {
		for index, validator := range validators {
			p.validators = append(p.validators, validator)
			p.addressToValidatorIndex[validator.Address] = index
			p.addressToStakedAmount[validator.Address] = std.Coin{"ugnot", 0} // no stake
		}
	}
}
