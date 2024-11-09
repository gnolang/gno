package proposal

import (
	"std"

	"gno.land/p/demo/dao"
	pVals "gno.land/p/sys/validators"
	govdao "gno.land/r/gov/dao/v2"
	validators "gno.land/r/sys/validators/v2"
)

func init() {
	changesFn := func() []pVals.Validator {
		return []pVals.Validator{
			{
				Address:     std.Address("g1x482v9p3hwetx2ayt2km7yntj4wsmqnd5at0lf"),                                               // change with your address
				PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpk9kmexf83hesgs0glzmtnpvc356h8vvrru209dmcxfghujzzzrrknpxl8", // change with your public key
				VotingPower: 1,                                                                                                     // leave this at 1
			},
		}
	}

	// Grab the executor for the valset changes
	executor := validators.NewPropExecutor(changesFn)

	prop := dao.ProposalRequest{
		Title:    "Leon prop 3",
		Executor: executor,
	}

	govdao.Propose(prop)
}
