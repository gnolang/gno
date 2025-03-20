package validators

var vp validators.ValsetProtocol // vp is the underlying validator set protocol

// addValidator adds a new validator to the validator set.
// If the validator is already present, the method errors out
func addValidator(validator validators.Validator) {
	val, err := vp.AddValidator(validator.Address, validator.PubKey, validator.VotingPower)
	if err != nil {
		panic(err)
	}

	// Emit the validator set change
	std.Emit(validators.ValidatorAddedEvent)
}

// removeValidator removes the given validator from the set.
// If the validator is not present in the set, the method errors out
func removeValidator(address std.Address) {
	val, err := vp.RemoveValidator(address)
	if err != nil {
		panic(err)
	}

	// Emit the validator set change
	std.Emit(validators.ValidatorRemovedEvent)
}
