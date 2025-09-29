package commondao

import (
	"gno.land/p/moul/addrset"
)

// MemberStorage defines an interface for member storages.
type MemberStorage interface {
	// Size returns the number of members in the storage.
	Size() int

	// Has checks if a member exists in the storage.
	Has(address) bool

	// Add adds a member to the storage.
	// Returns true if the member is added, or false if it already existed.
	Add(address) bool

	// Remove removes a member from the storage.
	// Returns true if member was removed, or false if it was not found.
	Remove(address) bool

	// Grouping returns member groups when supported.
	// When nil is returned it means that grouping of members is not supported.
	// Member groups can be used by implementations that require grouping users
	// by roles or by tiers for example.
	Grouping() MemberGrouping

	// IterateByOffset iterates members starting at the given offset.
	// The callback can return true to stop iteration.
	IterateByOffset(offset, count int, fn func(address) bool)
}

// NewMemberStorage creates a new member storage.
// Function returns a new member storage that doesn't support member groups.
// This type of storage is useful when there is no need to group members.
func NewMemberStorage() MemberStorage {
	return &memberStorage{}
}

// NewMemberStorageWithGrouping a new member storage with support for member groups.
// Member groups can be used by implementations that require grouping users by roles
// or by tiers for example.
func NewMemberStorageWithGrouping() MemberStorage {
	return &memberStorage{grouping: NewMemberGrouping()}
}

type memberStorage struct {
	addrset.Set

	grouping MemberGrouping
}

// Grouping returns member groups.
func (s memberStorage) Grouping() MemberGrouping {
	return s.grouping
}

// CountStorageMembers returns the total number of members in the storage.
// It counts all members in each group and the ones without group.
func CountStorageMembers(s MemberStorage) int {
	if s == nil {
		return 0
	}

	c := s.Size()
	s.Grouping().IterateByOffset(0, s.Grouping().Size(), func(g MemberGroup) bool {
		c += g.Members().Size()
		return false
	})
	return c
}
