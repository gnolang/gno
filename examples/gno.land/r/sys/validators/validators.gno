package validators

import (
	"std"
	"strconv"

	"gno.land/p/sys/validators"
)

// valset is the wrapper for the validator set protocol
type valset struct {
	p       validators.ValsetProtocol // p is the underlying validator set protocol
	changes []validators.Validator    // changes are the set changes that happened between scrapes
}

// v holds the active on-chain validator set state
var v *valset

// addValidator adds a new validator to the validator set.
// If the validator is already present, the method errors out
func addValidator(validator validators.Validator) {
	val, err := v.p.AddValidator(validator.Address, validator.PubKey, validator.VotingPower)
	if err != nil {
		panic(err)
	}

	if val != nil {
		// Validator added, note the change
		v.changes = append(v.changes, *val)

		// Emit the validator set change
		std.Emit(validators.ValidatorAddedEvent)
	}
}

// removeValidator removes the given validator from the set.
// If the validator is not present in the set, the method errors out
func removeValidator(address std.Address) {
	val, err := v.p.RemoveValidator(address)
	if err != nil {
		panic(err)
	}

	if val != nil {
		// Validator removed, note the change
		v.changes = append(v.changes, validators.Validator{
			Address:     val.Address,
			PubKey:      val.PubKey,
			VotingPower: 0, // nullified the voting power indicates removal
		})

		// Emit the validator set change
		std.Emit(validators.ValidatorRemovedEvent)
	}
}

func Render(_ string) string {
	if len(v.changes) == 0 {
		return "No valset changes to apply."
	}

	output := "Valset changes to apply:\n"
	for _, change := range v.changes {
		output += "- " + string(change.Address) + " (" + strconv.FormatUint(change.VotingPower, 10) + ")\n"
	}

	return output
}
