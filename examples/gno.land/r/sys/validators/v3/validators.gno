package validators

import (
	"gno.land/p/sys/validators"
)

var vp validators.ValsetProtocol // vp is the underlying validator set protocol

// addValidator adds a new validator to the validator set.
// If the validator is already present, the method panics.
func addValidator(validator validators.Validator) {
	if _, err := vp.AddValidator(validator.Address, validator.PubKey, validator.VotingPower); err != nil {
		panic(err)
	}
}

// removeValidator removes the given validator from the set.
// If the validator is not present in the set, the method panics.
func removeValidator(addr address) {
	if _, err := vp.RemoveValidator(addr); err != nil {
		panic(err)
	}
}
