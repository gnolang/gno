package governor

import (
	"gno.land/r/demo/groups"
	"std"
)

type Membership interface {
	AddMember(addr string, weight int, meta string)
	UpdateWeightCheckPoints(addr std.Address, amount int32, numTiers int32)
	TotalVotingPower() int32                                     // use int32 corresponding to maths.Rat for now
	TotalVotingPowerAt(height int64) int32                       // use int32 corresponding to maths.Rat for now
	VotingPowerFromMember(addr std.Address) int32                // this get weight, not direct tokens held
	VotingPowerFromMemberAt(addr std.Address, block int64) int32 // this get weight, not direct tokens held
	RenderMembership() string
}

type VotingPowerType int

const (
	VP_Fixed_1 VotingPowerType = iota // voting power set to fixed value per member
	VP_Grc20                          // voting power determined by grc20 token
	VP_Grc721                         // voting power determined by grc721 token
	VP_Native                         // voting power determined by native token, worx?
)

type GroupAdapter struct {
	groupID groups.GroupID // where voting power comes from
	// grc* token, native token, etc. determines voting power as a factor, one per dao
	// xxx: voting power is determined by several gates, like tokens, time gate, etc.
	votingPowerType VotingPowerType
}

func NewGroupAdapter(gid groups.GroupID) *GroupAdapter {
	return &GroupAdapter{
		groupID: gid,
	}
}

func (nca *GroupAdapter) AddMember(addr string, weight int, meta string) {
	groups.AddMember(nca.groupID, addr, weight, meta)
}

func (nca *GroupAdapter) UpdateWeightCheckPoints(addr std.Address, amount int32, numTiers int32) {
	groups.UpdateWeightCheckPoints(nca.groupID, addr, amount, numTiers)
}

func (nca *GroupAdapter) TotalVotingPower() int32 {
	// here differs between different DAOs, some like multi-sig style, while some with token hold/stake style.
	return int32(groups.GetTotalWeight(nca.groupID))
}

func (nca *GroupAdapter) TotalVotingPowerAt(height int64) int32 {
	// here differs between different DAOs, some like multi-sig style, while some with token hold/stake style.
	return int32(groups.GetTotalWeightAt(nca.groupID, height))
}

func (nca *GroupAdapter) VotingPowerFromMember(addr std.Address) int32 {
	return int32(groups.GetMemberWeight(nca.groupID, addr))
}

func (nca *GroupAdapter) VotingPowerFromMemberAt(addr std.Address, block int64) int32 {
	return int32(groups.GetMemberWeightAt(nca.groupID, addr, block))
}

func (nca *GroupAdapter) RenderMembership() string {
	return groups.RenderGroup(nca.groupID)
}

var _ Membership = (*GroupAdapter)(nil)
