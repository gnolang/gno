package validators

import (
	"chain"
	"chain/runtime"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/seqid"
	"gno.land/p/nt/ufmt"
	"gno.land/p/sys/validators"
)

var (
	vp            validators.ValsetProtocol // p is the underlying validator set protocol
	changes       *avl.Tree                 // changes holds any valset changes; seqid(block number) -> []change
	allowedRealms *avl.Tree                 // allowedRealms holds the whitelist of realm paths allowed to modify the validator set
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
		blockNum:  runtime.ChainHeight(),
		validator: val,
	}

	saveChange(ch)

	// Emit the validator set change
	chain.Emit(validators.ValidatorAddedEvent)
}

// removeValidator removes the given validator from the set.
// If the validator is not present in the set, the method errors out
func removeValidator(address_XXX address) {
	val, err := vp.RemoveValidator(address_XXX)
	if err != nil {
		panic(err)
	}

	// Validator removed, note the change
	ch := change{
		blockNum: runtime.ChainHeight(),
		validator: validators.Validator{
			Address:     val.Address,
			PubKey:      val.PubKey,
			VotingPower: 0, // nullified the voting power indicates removal
		},
	}

	saveChange(ch)

	// Emit the validator set change
	chain.Emit(validators.ValidatorRemovedEvent)
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
	var (
		size       = changes.Size()
		maxDisplay = 10
	)

	if size == 0 {
		return "No valset changes to apply."
	}

	output := "Valset changes:\n"
	changes.ReverseIterateByOffset(size-maxDisplay, maxDisplay, func(_ string, value any) bool {
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
