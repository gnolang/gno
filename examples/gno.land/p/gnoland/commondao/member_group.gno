package commondao

import (
	"gno.land/p/moul/addrset"
)

func NewMemberGroup(name string, members *addrset.Set) MemberGroup {
	if members == nil {
		panic("members are required to create a new tier with members")
	}

	return MemberGroup{
		ReadonlyMemberGroup: ReadonlyMemberGroup{
			name:    name,
			members: members,
		},
	}
}

// MemberGroup stores multiple members as a group.
// Member groups can be useful in different situations,
// for example to organize users in different tiers or groups.
type MemberGroup struct {
	ReadonlyMemberGroup
}

func (g MemberGroup) Readonly() ReadonlyMemberGroup {
	return g.ReadonlyMemberGroup
}

func (g MemberGroup) Members() *addrset.Set {
	return g.members
}

func (g *MemberGroup) SetValue(name string, value any) (updated bool) {
	return g.values.Set(name, value)
}

func (g MemberGroup) DeleteValue(name string) (deleted bool) {
	_, deleted = g.values.Remove(name)
	return
}
