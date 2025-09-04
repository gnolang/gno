package commondao

import "chain"

// NewMemberSet creates a new readonly memberset.
func NewMemberSet(s MemberStorage) MemberSet {
	return MemberSet{s}
}

// MemberSet contains a readonly list of member addresses.
type MemberSet struct {
	storage MemberStorage
}

// Size returns the number of addresses in the set.
func (s MemberSet) Size() int {
	return s.storage.Size()
}

// Has checks if an address exists in the set.
func (s MemberSet) Has(member chain.Address) bool {
	return s.storage.Has(member)
}

// IterateByOffset walks through addresses starting at the given offset.
// The callback should return true to stop iteration.
func (s MemberSet) IterateByOffset(offset, count int, fn func(chain.Address) bool) {
	s.storage.IterateByOffset(offset, count, fn)
}
