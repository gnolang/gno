package validators

import (
	"std"
	"strconv"

	"gno.land/p/sys/validators"
)

var (
	vp      validators.ValsetProtocol // p is the underlying validator set protocol
	changes []validators.Validator    // changes are the set changes that happened between scrapes
)

// addValidator adds a new validator to the validator set.
// If the validator is already present, the method errors out
func addValidator(validator validators.Validator) {
	val, err := vp.AddValidator(validator.Address, validator.PubKey, validator.VotingPower)
	if err != nil {
		panic(err)
	}

	if val != nil {
		// Validator added, note the change
		changes = append(changes, *val)

		// Emit the validator set change
		std.Emit(validators.ValidatorAddedEvent)
	}
}

// removeValidator removes the given validator from the set.
// If the validator is not present in the set, the method errors out
func removeValidator(address std.Address) {
	val, err := vp.RemoveValidator(address)
	if err != nil {
		panic(err)
	}

	if val != nil {
		// Validator removed, note the change
		changes = append(changes, validators.Validator{
			Address:     val.Address,
			PubKey:      val.PubKey,
			VotingPower: 0, // nullified the voting power indicates removal
		})

		// Emit the validator set change
		std.Emit(validators.ValidatorRemovedEvent)
	}
}

func Render(_ string) string {
	if len(changes) == 0 {
		return "No valset changes to apply."
	}

	output := "Valset changes to apply:\n"
	for _, change := range changes {
		output += "- " + string(change.Address) + " (" + strconv.FormatUint(change.VotingPower, 10) + ")\n"
	}

	return output
}
