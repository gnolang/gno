package memberstore

import (
	"std"

	"gno.land/r/gov/dao"
)

type ChangeTiersMetadata struct {
	Tiers TiersByName
}

func (a *ChangeTiersMetadata) IsMetadata() {}
func (a *ChangeTiersMetadata) String() string {
	return "New set of tiers proposed."
}

func changeTiersCallback(m dao.Metadata) error {
	amm, ok := m.(*ChangeTiersMetadata)
	if !ok {
		panic("metadata type for changing tiers is wrong")
	}

	Tiers = amm.Tiers

	return nil
}

func NewChangeTiersRequest(tiers TiersByName) dao.ProposalRequest {
	member, _ := Get().GetMember(std.OriginCaller())
	if member == nil {
		panic("proposer is not a member")
	}

	return dao.
		NewProposalRequestBuilder("Change Tiers Proposal").
		WithDescription("This proposal is looking to change the existing Tiers in memberstore").
		WithCallback(changeTiersCallback).
		WithMetadata(&ChangeTiersMetadata{
			Tiers: tiers,
		}).
		Build()
}
