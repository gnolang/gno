package proposal

import (
	"errors"

	"gno.land/p/nt/ufmt/v0"
	valopers "gno.land/r/gnops/valopers"
	"gno.land/r/gov/dao"
	sysparams "gno.land/r/sys/params"
	validators "gno.land/r/sys/validators/v3"
)

var (
	ErrValidatorMissing = errors.New("the validator is missing")
	ErrSameValues       = errors.New("the valoper has the same voting power and pubkey")
)

// NewValidatorProposalRequest creates a proposal request to the GovDAO
// for adding (or removing) the given valoper to/from the validator set.
//
// Signature is preserved for historical-replay compatibility (gnoland-1
// callers call this with a single address). Body is rewired to call
// v3's operator-keyed NewValidatorProposalRequest with a single-element
// slice; v3's executor re-resolves the signing pubkey from valoperCache
// at execution time, so a mid-flight rotation publishes the current key.
func NewValidatorProposalRequest(cur realm, addr address) dao.ProposalRequest {
	var (
		valoper     = valopers.GetByAddr(addr)
		votingPower = uint64(1)
	)

	exist := validators.IsValidator(valoper.SigningAddress)

	// Determine the voting power
	if !valoper.KeepRunning {
		if !exist {
			panic(ErrValidatorMissing)
		}
		votingPower = uint64(0)
	}

	if exist {
		validator := validators.GetValidator(valoper.SigningAddress)
		if validator.VotingPower == votingPower && validator.PubKey == valoper.SigningPubKey {
			panic(ErrSameValues)
		}
	}

	// Craft the proposal title and description, framed around the
	// valoper profile. Voters see the operator identity (moniker +
	// operator address); the signing key is an implementation detail
	// resolved at execution time by v3's executor.
	title := ufmt.Sprintf(
		"Add valoper %s to the valset",
		valoper.Moniker,
	)

	description := ufmt.Sprintf("Valoper profile: [%s](/r/gnops/valopers:%s)\n\n%s",
		valoper.Moniker,
		valoper.OperatorAddress,
		valoper.Render(),
	)

	return validators.NewValidatorProposalRequest(
		[]validators.ValoperChange{
			{OperatorAddress: valoper.OperatorAddress, Power: votingPower},
		},
		title,
		description,
	)
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
// for updating the minimum fee to register a new valoper. Signature
// preserved for historical-replay compatibility (gnoland-1's
// set_minfee.gno MsgRun calls this); body now delegates to the
// generic sys/params factory so the fee lives in
// node:valoper:register_fee.
func ProposeNewMinFeeProposalRequest(cur realm, newMinFee int64) dao.ProposalRequest {
	return sysparams.NewSysParamUint64PropRequest("node", "valoper", "register_fee", uint64(newMinFee))
}
