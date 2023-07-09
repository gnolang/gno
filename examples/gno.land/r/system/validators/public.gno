package validators

import (
	"std"

	"gno.land/p/demo/ufmt"
)

// CreateValidator creates a new validator with the given name and public key
func CreateValidator(name, pubkey string) {
	std.AssertOriginCall()

	caller := std.GetOrigCaller()
	vkey := validatorKey(caller)
	if validators.Has(vkey) {
		panic("validator already registered")
	}

	var power int64 = 0 // when created validator, power is 0, admin must confirm this validator is active or not
	validator := newValidator(name, caller, pubkey, power)
	validators.Set(vkey, validator)
}

func RemoveValidator() {
	std.AssertOriginCall()

	vkey := validatorKey(std.GetOrigCaller())
	if !validators.Has(vkey) {
		panic("validator not found")
	}

	validators.Remove(vkey)
}

// UpdateValidatorPower updates the power of the given validator by admins
func UpdateValidatorPower(addr std.Address, power int64) {
	assertIsAdmin()
	vkey := validatorKey(addr)

	v, ok := validators.Get(vkey)
	if !ok {
		panic("validator not found")
	}
	v.(*Validator).power = power
}

// ValidatorSet returns the current active validator set
func ValidatorSet() string {
	output := ""
	colon := ""
	validators.Iterate("", "", func(key string, value interface{}) bool {
		validator := value.(*Validator)
		if validator.power == 0 {
			return false
		}

		output = output + colon + ufmt.Sprintf("%s:%d", validator.pubkey, validator.power)
		colon = ","
		return false

	})
	return output
}
