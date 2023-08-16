package governor

import (
	"gno.land/r/demo/groups"
)

type Membership interface {
	AddMember(addr string, weight int, meta string)
	UpdateMemberWeight(mid uint64, weight int)
	TotalVotingPower() int32 // use int32 corresponding to maths.Rat for now
	GetVotingPowerFromMember(mid uint64) int32
	RenderMembership() string
}

type NoCoinAdapter struct {
	groupID groups.GroupID // where voting power comes from
}

func NewNoCoinAdapter(gid groups.GroupID) *NoCoinAdapter {
	return &NoCoinAdapter{
		groupID: gid,
	}
}

func (nca *NoCoinAdapter) AddMember(addr string, weight int, meta string) {
	groups.AddMember(nca.groupID, addr, weight, meta)
}

func (nca *NoCoinAdapter) UpdateMemberWeight(mid uint64, weight int) {
	groups.UpdateMemberWeight(nca.groupID, mid, weight)
}

func (nca *NoCoinAdapter) TotalVotingPower() int32 {
	// here differs between different DAOs, some like multi-sig style, while some with token hold/stake style.
	return int32(groups.GetMemberCount(nca.groupID))
}

func (nca *NoCoinAdapter) GetVotingPowerFromMember(mid uint64) int32 {
	return int32(groups.GetMemberWeight(nca.groupID, mid))
}

func (nca *NoCoinAdapter) RenderMembership() string {
	return groups.RenderGroup(nca.groupID)
}

var _ Membership = (*NoCoinAdapter)(nil)
