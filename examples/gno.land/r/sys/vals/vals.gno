package vals

import (
	"std"
	"strconv"

	"gno.land/p/sys/vals/types"
)

// vals is the wrapper for the validator set protocol
type vals struct {
	p       types.Protocol    // p is the underlying validator set protocol
	changes []types.Validator // changes are the set changes that happened between scrapes
}

// v holds the active on-chain validator set state
var v *vals

// addValidator adds a new validator to the validator set.
// If the validator is already present, the method errors out
func addValidator(validator types.Validator) {
	if val := v.p.AddValidator(
		validator.Address,
		validator.PubKey,
		validator.VotingPower,
	); val != nil {
		// Validator added, note the change
		v.changes = append(v.changes, types.Validator{
			Address:     val.Address,
			PubKey:      val.PubKey,
			VotingPower: val.VotingPower,
		})

		// Emit the validator set change
		std.Emit(types.ValidatorAddedEvent)
	}
}

// removeValidator removes the given validator from the set.
// If the validator is not present in the set, the method errors out
func removeValidator(address std.Address) {
	if val := v.p.RemoveValidator(address); val != nil {
		// Validator removed, note the change
		v.changes = append(v.changes, types.Validator{
			Address:     val.Address,
			PubKey:      val.PubKey,
			VotingPower: 0, // nullified the voting power indicates removal
		})

		// Emit the validator set change
		std.Emit(types.ValidatorRemovedEvent)
	}
}

// getChanges returns the validator changes stored on the realm
func getChanges() []types.Validator {
	// Construct the changes
	changes := make([]types.Validator, len(v.changes))
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
