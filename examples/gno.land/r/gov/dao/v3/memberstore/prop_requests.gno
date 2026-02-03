package memberstore

import (
	"chain/runtime"

	"gno.land/p/nt/avl"

	"gno.land/r/gov/dao"
)

func NewChangeTiersRequest(tiers map[string]Tier) dao.ProposalRequest {
	if len(tiers) == 0 {
		panic("tiers list is empty")
	}

	member, _ := Get().GetMember(runtime.OriginCaller())
	if member == nil {
		panic("proposer is not a member")
	}

	newTiers := TiersByName{avl.NewTree()}
	for name, tier := range tiers {
		newTiers.Set(name, tier)
	}

	callback := func(cur realm) error {
		Tiers = newTiers

		return nil
	}

	e := dao.NewSimpleExecutor(callback, "New set of tiers proposed.")

	return dao.NewProposalRequest("Change Tiers Proposal", "This proposal is looking to change the existing Tiers in memberstore", e)
}
