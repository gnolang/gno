package governor

import (
	"gno.land/r/demo/groups"
)

type Membership interface {
	AddMember(addr string, weight int, meta string)
	UpdateWeightCheckPoints(mid uint64, amount int32, numTiers int32)
	TotalVotingPower() int32                   // use int32 corresponding to maths.Rat for now
	GetVotingPowerFromMember(mid uint64) int32 // this get weight, not direct tokens held
	RenderMembership() string
}

type VotingPowerType int

//const (
//	VP_Fixed_1 VotingPowerType = iota // voting power set to 1 per member
//	VP_ByToken                        // voting power determined by tokens one hold in specific snapshot
//)

type GroupAdapter struct {
	groupID groups.GroupID // where voting power comes from
}

func NewGroupAdapter(gid groups.GroupID) *GroupAdapter {
	return &GroupAdapter{
		groupID: gid,
	}
}

func (nca *GroupAdapter) AddMember(addr string, weight int, meta string) {
	groups.AddMember(nca.groupID, addr, weight, meta)
}

func (nca *GroupAdapter) UpdateWeightCheckPoints(mid uint64, amount int32, numTiers int32) {
	groups.UpdateWeightCheckPoints(nca.groupID, mid, amount, numTiers)
}

func (nca *GroupAdapter) TotalVotingPower() int32 {
	// here differs between different DAOs, some like multi-sig style, while some with token hold/stake style.
	return int32(groups.GetTotalWeight(nca.groupID))
}

func (nca *GroupAdapter) GetVotingPowerFromMember(mid uint64) int32 {
	return int32(groups.GetMemberWeight(nca.groupID, mid))
}

func (nca *GroupAdapter) RenderMembership() string {
	return groups.RenderGroup(nca.groupID)
}

var _ Membership = (*GroupAdapter)(nil)
