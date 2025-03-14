package commondao

import (
	"std"

	"gno.land/p/moul/addrset"
)

// NewMemberSet creates a new readonly memberset.
func NewMemberSet(members addrset.Set) MemberSet {
	return MemberSet{members}
}

// MemberSet contains a readonly list of member addresses.
type MemberSet struct {
	members addrset.Set
}

// Size returns the number of addresses in the set.
func (s MemberSet) Size() int {
	return s.members.Size()
}

// Has checks if an address exists in the set.
func (s MemberSet) Has(member std.Address) bool {
	return s.members.Has(member)
}

// IterateByOffset walks through addresses starting at the given offset.
// The callback should return true to stop iteration.
func (s MemberSet) IterateByOffset(offset, count int, fn func(std.Address) bool) {
	s.members.IterateByOffset(offset, count, fn)
}
