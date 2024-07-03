package validators

import (
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/sys/validators"
)

var (
	vp      validators.ValsetProtocol // p is the underlying validator set protocol
	changes []change                  // changes holds any valset changes
)

// change represents a single valset change, tied to a specific block number
type change struct {
	blockNum  int64                // the block number associated with the valset change
	validator validators.Validator // the validator update
}

// addValidator adds a new validator to the validator set.
// If the validator is already present, the method errors out
func addValidator(validator validators.Validator) {
	val, err := vp.AddValidator(validator.Address, validator.PubKey, validator.VotingPower)
	if err != nil {
		panic(err)
	}

	// Validator added, note the change
	changes = append(changes, change{
		blockNum:  std.GetHeight(),
		validator: val,
	})

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

	// Validator removed, note the change
	changes = append(changes, change{
		blockNum: std.GetHeight(),
		validator: validators.Validator{
			Address:     val.Address,
			PubKey:      val.PubKey,
			VotingPower: 0, // nullified the voting power indicates removal
		},
	})

	// Emit the validator set change
	std.Emit(validators.ValidatorRemovedEvent)
}

func Render(_ string) string {
	if len(changes) == 0 {
		return "No valset changes to apply."
	}

	output := "Valset changes:\n"
	for _, change := range changes {
		output += ufmt.Sprintf(
			"- #%d: %s (%d)\n",
			change.blockNum,
			string(change.validator.Address),
			change.validator.VotingPower,
		)
	}

	return output
}
