package proposal

import (
	"errors"

	"gno.land/p/nt/ufmt/v0"
	pVals "gno.land/p/sys/validators"
	valopers "gno.land/r/gnops/valopers"
	"gno.land/r/gov/dao"
	validators "gno.land/r/sys/validators/v2"
)

var (
	ErrValidatorMissing = errors.New("the validator is missing")
	ErrSameValues       = errors.New("the valoper has the same voting power and pubkey")
)

// NewValidatorProposalRequest creates a proposal request to the GovDAO
// for adding the given valoper to the validator set.
func NewValidatorProposalRequest(cur realm, address_XXX address) dao.ProposalRequest {
	var (
		valoper     = valopers.GetByAddr(address_XXX)
		votingPower = uint64(1)
	)

	exist := validators.IsValidator(address_XXX)

	// Determine the voting power
	if !valoper.KeepRunning {
		if !exist {
			panic(ErrValidatorMissing)
		}
		votingPower = uint64(0)
	}

	if exist {
		validator := validators.GetValidator(address_XXX)
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

	// Craft the proposal title
	title := ufmt.Sprintf(
		"Add valoper %s to the valset",
		valoper.Moniker,
	)

	description := ufmt.Sprintf("Valoper profile: [%s](/r/gnops/valopers:%s)\n\n%s",
		valoper.Moniker,
		valoper.Address,
		valoper.Render(),
	)

	// Create the request
	return validators.NewPropRequest(changesFn, title, description)
}

// ProposeNewInstructionsProposalRequest creates a proposal to the GovDAO
// for updating the realm instructions.
func ProposeNewInstructionsProposalRequest(cur realm, newInstructions string) dao.ProposalRequest {
	cb := valopers.NewInstructionsProposalCallback(newInstructions)
	// Create a proposal
	title := "/p/gnops/valopers: Update instructions"
	description := ufmt.Sprintf("Update the instructions to: \n\n%s", newInstructions)

	e := dao.NewSimpleExecutor(cb, "")

	return dao.NewProposalRequest(title, description, e)
}

// ProposeNewMinFeeProposalRequest creates a proposal to the GovDAO
// for updating the minimum fee to register a new valoper.
func ProposeNewMinFeeProposalRequest(cur realm, newMinFee int64) dao.ProposalRequest {
	cb := valopers.NewMinFeeProposalCallback(newMinFee)
	// Create a proposal
	title := "/p/gnops/valopers: Update minFee"
	description := ufmt.Sprintf("Update the minimum register fee to: %d ugnot", newMinFee)

	e := dao.NewSimpleExecutor(cb, "")

	return dao.NewProposalRequest(title, description, e)
}
