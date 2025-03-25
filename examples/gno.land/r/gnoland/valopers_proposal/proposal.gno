package valopers_proposal

import (
	"errors"
	"std"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/ufmt"
	pVals "gno.land/p/sys/validators"
	valopers "gno.land/r/gnoland/valopers"
	"gno.land/r/gov/dao/bridge"
	validators "gno.land/r/sys/validators/v2"
)

var (
	ErrValidatorMissing = errors.New("the validator is missing")
	ErrSameValues       = errors.New("the valoper has the same voting power and pubkey")
)

// ProposeNewValidator creates a proposal to the GovDAO
// for adding the given valoper to the validator set.
func ProposeNewValidator(address std.Address) {
	var (
		valoper     = valopers.GetByAddr(address)
		votingPower = uint64(1)
	)

	exist := validators.IsValidator(address)

	// Determine the voting power
	if !valoper.KeepRunning {
		if !exist {
			panic(ErrValidatorMissing)
		}
		votingPower = uint64(0)
	}

	if exist {
		validator := validators.GetValidator(address)
		if validator.VotingPower == votingPower && validator.PubKey == valoper.PubKey {
			panic(ErrSameValues)
		}
	}

	changesFn := func() []pVals.Validator {
		return []pVals.Validator{
			{
				Address:     valoper.Address,
				PubKey:      valoper.PubKey,
				VotingPower: votingPower,
			},
		}
	}

	// Create the executor
	executor := validators.NewPropExecutor(changesFn)

	// Craft the proposal title
	title := ufmt.Sprintf(
		"Add valoper %s to the valset",
		valoper.Moniker,
	)

	description := ufmt.Sprintf("Valoper profile: [%s](/r/gnoland/valopers:%s)\n\n%s",
		valoper.Moniker,
		valoper.Address,
		valoper.Render(),
	)

	prop := dao.ProposalRequest{
		Title:       title,
		Description: description,
		Executor:    executor,
	}

	// Create the govdao proposal
	bridge.GovDAO().Propose(prop)
}

// ProposeNewInstructions creates a proposal to the GovDAO
// for updating the realm instructions.
func ProposeNewInstructions(newInstructions string) {
	executor := valopers.NewInstructionsExecutor(newInstructions)

	// Create a proposal
	title := "/p/gnoland/valopers: Update instructions"
	description := ufmt.Sprintf("Update the instructions to: \n\n%s", newInstructions)

	prop := dao.ProposalRequest{
		Title:       title,
		Description: description,
		Executor:    executor,
	}

	bridge.GovDAO().Propose(prop)
}

// ProposeNewMinFee creates a proposal to the GovDAO
// for updating the minimum fee to register a new valoper.
func ProposeNewMinFee(newMinFee int64) {
	executor := valopers.NewMinFeeExecutor(newMinFee)

	// Create a proposal
	title := "/p/gnoland/valopers: Update minFee"
	description := ufmt.Sprintf("Update the minimum register fee to: %d ugnot", newMinFee)

	prop := dao.ProposalRequest{
		Title:       title,
		Description: description,
		Executor:    executor,
	}

	bridge.GovDAO().Propose(prop)
}
