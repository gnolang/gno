package valoper_dao_proposal

import (
	"std"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/ufmt"
	pVals "gno.land/p/sys/validators"
	valopers "gno.land/r/gnoland/valopers"
	"gno.land/r/gov/dao/bridge"
	validators "gno.land/r/sys/validators/v2"
)

// MakeProposal creates a proposal to the GovDAO
// for adding the given valoper to the validator set.
func MakeProposal(address std.Address) {
	var (
		valoper     = valopers.GetByAddr(address)
		votingPower = uint64(1)
	)

	// Determine the voting power
	if !valoper.Active {
		votingPower = uint64(0)
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
		"Add valoper %s (Address: %s; PubKey: %s) to the valset",
		valoper.Moniker,
		valoper.Address.String(),
		valoper.PubKey,
	)

	prop := dao.ProposalRequest{
		Title:       title,
		Description: valoper.Render(),
		Executor:    executor,
	}

	// Create the govdao proposal
	bridge.GovDAO().Propose(prop)
}
