package validators

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
	"gno.land/p/sys/validators"
)

var (
	vp      validators.ValsetProtocol // p is the underlying validator set protocol
	changes *avl.Tree                 // changes holds any valset changes; seqid(block number) -> []change
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
	ch := change{
		blockNum:  std.GetHeight(),
		validator: val,
	}

	saveChange(ch)

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
	ch := change{
		blockNum: std.GetHeight(),
		validator: validators.Validator{
			Address:     val.Address,
			PubKey:      val.PubKey,
			VotingPower: 0, // nullified the voting power indicates removal
		},
	}

	saveChange(ch)

	// Emit the validator set change
	std.Emit(validators.ValidatorRemovedEvent)
}

// saveChange saves the valset change
func saveChange(ch change) {
	id := getBlockID(ch.blockNum)

	setRaw, exists := changes.Get(id)
	if !exists {
		changes.Set(id, []change{ch})

		return
	}

	// Save the change
	set := setRaw.([]change)
	set = append(set, ch)

	changes.Set(id, set)
}

// getBlockID converts the block number to a sequential ID
func getBlockID(blockNum int64) string {
	return seqid.ID(uint64(blockNum)).String()
}

func Render(_ string) string {
	if changes.Size() == 0 {
		return "No valset changes to apply."
	}

	output := "Valset changes:\n"

	changes.Iterate("", "", func(_ string, value interface{}) bool {
		chs := value.([]change)

		for _, ch := range chs {
			output += ufmt.Sprintf(
				"- #%d: %s (%d)\n",
				ch.blockNum,
				ch.validator.Address.String(),
				ch.validator.VotingPower,
			)
		}

		return false
	})

	return output
}
