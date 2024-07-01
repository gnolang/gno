package vals

import (
	"std"
	"strconv"

	"gno.land/p/sys/vals"
)

// valset is the wrapper for the validator set protocol
type valset struct {
	p       vals.ValsetProtocol // p is the underlying validator set protocol
	changes []vals.Validator    // changes are the set changes that happened between scrapes
}

// v holds the active on-chain validator set state
var v *valset

// addValidator adds a new validator to the validator set.
// If the validator is already present, the method errors out
func addValidator(validator vals.Validator) {
	val, err := v.p.AddValidator(validator.Address, validator.PubKey, validator.VotingPower)
	if err != nil {
		panic(err)
	}

	if val != nil {
		// Validator added, note the change
		v.changes = append(v.changes, *val)

		// Emit the validator set change
		std.Emit(vals.ValidatorAddedEvent)
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
		v.changes = append(v.changes, vals.Validator{
			Address:     val.Address,
			PubKey:      val.PubKey,
			VotingPower: 0, // nullified the voting power indicates removal
		})

		// Emit the validator set change
		std.Emit(vals.ValidatorRemovedEvent)
	}
}

// getChanges returns the validator changes stored on the realm
func getChanges() []vals.Validator {
	// Construct the changes
	changes := make([]vals.Validator, len(v.changes))
	copy(changes, v.changes)

	// Reset the changes set
	v.changes = v.changes[:0]

	return changes
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
