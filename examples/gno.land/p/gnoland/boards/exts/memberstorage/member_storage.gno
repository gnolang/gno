package memberstorage

import (
	"gno.land/p/moul/addrset"
)

type (
	// MemberIterFn defines a callback to iterate members.
	MemberIterFn func(address) bool

	// MemberStorage defines an interface for member storages.
	MemberStorage interface {
		// Size returns the number of ungrouped members in the storage.
		Size() int

		// Has checks if an ungrouped member exists in the storage.
		// Grouped members are not checked; use ExistsInMemberStorage to
		// check across grouped and ungrouped members.
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

		// IterateByOffset iterates ungrouped members starting at the given offset.
		// Grouped members are not iterated; use IterateMemberStorage to walk
		// both grouped and ungrouped members.
		// The callback can return true to stop iteration.
		IterateByOffset(offset, count int, fn MemberIterFn) (stopped bool)
	}
)

// IsCanonicalMemberStorage checks whether a member storage is the canonical
// one, created using the New or NewWithGrouping functions defined within
// this package.
//
// Use this function at any public entry point that accepts a MemberStorage
// from an external caller before invoking any of its methods.
func IsCanonicalMemberStorage(s MemberStorage) bool {
	_, ok := s.(*memberStorage)
	return ok
}

// New creates a new member storage that doesn't support member groups.
// This type of storage is useful when there is no need to group members.
func New() MemberStorage {
	return &memberStorage{}
}

// NewWithGrouping creates a new member storage with support for member groups.
// Member groups can be used by implementations that require grouping users by roles
// or by tiers for example. It also supports members without grouping.
func NewWithGrouping() MemberStorage {
	return &memberStorage{grouping: NewMemberGrouping()}
}

type memberStorage struct {
	members  addrset.Set
	grouping MemberGrouping
}

// Size returns the number of ungrouped members in the storage.
func (s memberStorage) Size() int {
	return s.members.Size()
}

// Has checks if a member exists in the base underlying storage.
// Grouped members are not checked.
func (s memberStorage) Has(member address) bool {
	return s.members.Has(member)
}

// Add adds a member to the storage.
// Returns true if the member is added, or false if it already existed.
func (s *memberStorage) Add(member address) bool {
	return s.members.Add(member)
}

// Remove removes a member from the storage.
// Returns true if member was removed, or false if it was not found.
func (s *memberStorage) Remove(member address) bool {
	return s.members.Remove(member)
}

// Grouping returns member groups.
func (s memberStorage) Grouping() MemberGrouping {
	return s.grouping
}

// IterateByOffset iterates ungrouped members starting at the given offset.
// The callback can return true to stop iteration.
// To iterate across both grouped and ungrouped members use the
// IterateMemberStorage helper.
func (s memberStorage) IterateByOffset(offset, count int, fn MemberIterFn) bool {
	if offset < 0 {
		offset = 0
	}

	var stopped bool
	s.members.IterateByOffset(offset, count, func(member address) bool {
		stopped = fn(member)
		return stopped
	})
	return stopped
}

// GetTotalMemberStorageSize returns the total number of member occurrences
// in the storage, counting grouped and ungrouped members. The same address
// appearing in multiple groups (or in both grouped and ungrouped) is counted
// once per occurrence; results are not deduplicated.
func GetTotalMemberStorageSize(s MemberStorage) int {
	// TODO: Support deduplication, for example through a UniqueMemberIterator interface

	count := s.Size()
	if s.Grouping() == nil {
		return count
	}

	s.Grouping().IterateByOffset(0, s.Grouping().Size(), func(g MemberGroup) bool {
		count += g.Members().Size()
		return false
	})
	return count
}

// ExistsInMemberStorage checks if a member exists within a storage.
// It checks grouped and ungrouped members.
func ExistsInMemberStorage(s MemberStorage, member address) bool {
	if s.Has(member) {
		// Member exists as an ungrouped member
		return true
	}

	if s.Grouping() == nil {
		return false
	}

	// Check if member exists within any of the groups
	return s.Grouping().IterateByOffset(0, s.Grouping().Size(), func(g MemberGroup) bool {
		return g.Members().Has(member)
	})
}

// IterateMemberStorage iterates members starting at the given offset across
// the entire storage. Ungrouped members are iterated first, followed by the
// members of each group in group iteration order. Members that appear in
// multiple groups (or in both grouped and ungrouped) are visited once per
// occurrence. A negative offset is treated as 0. The callback can return true
// to stop iteration.
func IterateMemberStorage(s MemberStorage, offset, count int, fn MemberIterFn) bool {
	if count <= 0 {
		return false
	}
	if offset < 0 {
		offset = 0
	}

	// Define a callback to walk storage members
	var stopped bool
	walk := func(storage MemberStorage) bool {
		size := storage.Size()
		if offset >= size {
			offset -= size
			return false
		}

		// Update current count to fit storage boundaries
		currentCount := count
		if currentCount > size-offset {
			currentCount = size - offset
		}

		if storage.IterateByOffset(offset, currentCount, fn) {
			stopped = true
			return true // Walk stopped by the callback
		}

		// Discount iterated members and reset offset for next grouping storage
		count -= currentCount
		offset = 0

		// When zero walk stops because specified amount of members were iterated
		return count == 0
	}

	// Walk ungrouped members
	if walk(s) {
		return stopped
	}

	grouping := s.Grouping()
	if grouping == nil {
		return false
	}

	// Walk grouped members
	grouping.IterateByOffset(0, grouping.Size(), func(g MemberGroup) bool {
		return walk(g.Members())
	})
	return stopped
}
