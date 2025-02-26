package memberstore

import (
	"std"

	"gno.land/r/gov/dao/proxy"
)

type ChangeTiersMetadata struct {
	Tiers TiersByName
}

func (a *ChangeTiersMetadata) IsMetadata() {}
func (a *ChangeTiersMetadata) String() string {
	return "New set of tiers proposed."
}

func changeTiersCallback(m proxy.Metadata) error {
	amm, ok := m.(*ChangeTiersMetadata)
	if !ok {
		panic("metadata type for changing tiers is wrong")
	}

	Tiers = amm.Tiers

	return nil
}

func NewChangeTiersRequest(tiers TiersByName) proxy.ProposalRequest {
	member, _ := Get().GetMember(std.GetOrigCaller())
	if member == nil {
		panic("proposer is not a member")
	}

	return proxy.ProposalRequest{
		Title:       "Change Tiers Proposal",
		Description: "This prpoposal is looking to change the existing Tiers in memberstore",
		Callback:    changeTiersCallback,
		Metadata: &ChangeTiersMetadata{
			Tiers: tiers,
		},
	}
}
