package validators

import (
	"std"
)

func CreateValidator(name, pubkey string) {
	std.AssertOriginCall()

	if validators.Size() >= maxValidators {
		panic("validator set is full")
	}

	caller := std.GetOrigCaller()
	vkey := validatorKey(caller)
	if validators.Has(vkey) {
		panic("validator already registered")
	}

	power := std.GetOrigSend().AmountOf(nativeDenom)

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

func UpdateMaxValidators(max int) {
	std.AssertOriginCall()
	caller := std.GetOrigCaller()
	if caller != admin {
		panic("unauthorized")
	}

	maxValidators = max
}

func ValidatorSet() string {
	output := ""
	colon := ""
	validators.Iterate("", "", func(key string, value interface{}) bool {
		validator := value.(*Validator)
		output = output + colon + validator.JSON()
		colon = ","
		return false

	})

	return "[" + output + "]"
}
