package memberstore

import (
	"chain/runtime"

	"gno.land/r/gov/dao"
)

func NewChangeTiersRequest(tiers TiersByName) dao.ProposalRequest {
	member, _ := Get().GetMember(runtime.OriginCaller())
	if member == nil {
		panic("proposer is not a member")
	}

	callback := func(cur realm) error {
		// TODO: Assigning tiers directly is not safe, could be potentially changed after proposal passes
		Tiers = tiers

		return nil
	}

	e := dao.NewSimpleExecutor(callback, "New set of tiers proposed.")

	return dao.NewProposalRequest("Change Tiers Proposal", "This proposal is looking to change the existing Tiers in memberstore", e)
}
