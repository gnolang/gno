package commondao

import (
	"gno.land/p/moul/addrset"
)

type (
	// MemberIterFn defines a callback to iterate DAO members.
	MemberIterFn func(address) bool

	// MemberStorage defines an interface for member storages.
	MemberStorage interface {
		Sealable

		// Size returns the number of ungrouped members in the storage.
		Size() int

		// TotalSize returns the number of members in the storage.
		// It counts grouped and ungrouped members.
		TotalSize() int

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
		IterateByOffset(offset, count int, fn MemberIterFn) (stopped bool)
	}
)

// NewMemberStorage creates a new member storage.
// Function returns a new member storage that doesn't support member groups.
// This type of storage is useful when there is no need to group members.
func NewMemberStorage() MemberStorage {
	return &memberStorage{}
}

// NewMemberStorageWithGrouping a new member storage with support for member groups.
// Member groups can be used by implementations that require grouping users by roles
// or by tiers for example. It also supports members without grouping.
func NewMemberStorageWithGrouping(options ...MemberGroupingOption) MemberStorage {
	return &memberStorage{grouping: NewMemberGrouping(options...)}
}

type memberStorage struct {
	members  addrset.Set
	grouping MemberGrouping
	sealed   bool
}

// Seal returns a sealed copy of the storage.
// No further changes are accepted on the returned copy once sealed.
func (s memberStorage) Seal() Sealable {
	s.sealed = true
	if s.grouping != nil {
		s.grouping = s.grouping.Seal().(MemberGrouping)
	}
	return &s
}

// IsSealed checks if storage is sealed.
func (s memberStorage) IsSealed() bool {
	return s.sealed
}

// Size returns the number of ungrouped members in the storage.
// To get the size including ungrouped members use `MemberStorage.TotalSize()`.
func (s memberStorage) Size() int {
	return s.members.Size()
}

// TotalSize returns the number of members in the storage.
// It counts grouped and ungrouped members.
func (s memberStorage) TotalSize() int {
	count := s.Size()
	if s.grouping == nil {
		return count
	}

	s.grouping.IterateByOffset(0, s.Grouping().Size(), func(g MemberGroup) bool {
		count += g.Members().Size()
		return false
	})
	return count
}

// Has checks if a member exists in the base underlying storage.
// Grouped members are not checked.
func (s memberStorage) Has(member address) bool {
	return s.members.Has(member)
}

// Add adds a member to the storage.
// Returns true if the member is added, or false if it already existed.
// It panics when storage is sealed.
func (s *memberStorage) Add(member address) bool {
	assertNotSealed(s)
	return s.members.Add(member)
}

// Remove removes a member from the storage.
// Returns true if member was removed, or false if it was not found.
// It panics when storage is sealed.
func (s *memberStorage) Remove(member address) bool {
	assertNotSealed(s)
	return s.members.Remove(member)
}

// Grouping returns member groups.
func (s memberStorage) Grouping() MemberGrouping {
	if s.sealed && s.grouping != nil {
		return s.grouping.Seal().(MemberGrouping)
	}
	return s.grouping
}

// IterateByOffset iterates members starting at the given offset.
// The callback can return true to stop iteration.
// It doesn't iterate grouped members, if you need to iterate grouped members
// use `MemberStorage.Grouping().IterateByOffset()`.
func (s memberStorage) IterateByOffset(offset, count int, fn MemberIterFn) bool {
	var stopped bool
	s.members.IterateByOffset(offset, count, func(member address) bool {
		stopped = fn(member)
		return stopped
	})
	return stopped
}
