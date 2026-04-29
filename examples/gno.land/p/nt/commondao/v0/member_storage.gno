package commondao

import (
	"errors"

	"gno.land/p/moul/addrset"
)

type (
	// MemberIterFn defines a callback to iterate DAO members.
	MemberIterFn func(address) bool

	// MemberStorage defines an interface for member storages.
	MemberStorage interface {
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

// NewReadonlyMemberStorage creates a new readonly member storage.
func NewReadonlyMemberStorage(s MemberStorage) (*ReadonlyMemberStorage, error) {
	if s == nil {
		return nil, errors.New("member storage is required")
	}
	return &ReadonlyMemberStorage{s}, nil
}

// MustNewReadonlyMemberStorage creates a new readonly member storage or panics on error.
func MustNewReadonlyMemberStorage(s MemberStorage) *ReadonlyMemberStorage {
	storage, err := NewReadonlyMemberStorage(s)
	if err != nil {
		panic(err)
	}
	return storage
}

// ReadonlyMemberStorage defines a readonly member storage.
type ReadonlyMemberStorage struct {
	storage MemberStorage
}

// Size returns the number of ungrouped members in the storage.
// To get the size including ungrouped members use `ReadonlyMemberStorage.TotalSize()`.
func (s ReadonlyMemberStorage) Size() int {
	if s.storage == nil {
		return 0
	}
	return s.storage.Size()
}

// TotalSize returns the number of members in the storage.
// It counts grouped and ungrouped members.
func (s ReadonlyMemberStorage) TotalSize() int {
	if s.storage == nil {
		return 0
	}
	return s.storage.TotalSize()
}

// Has checks if a member exists in the base underlying storage.
// Grouped members are not checked.
func (s ReadonlyMemberStorage) Has(member address) bool {
	if s.storage == nil {
		return false
	}
	return s.storage.Has(member)
}

// Grouping returns member groups.
func (s ReadonlyMemberStorage) Grouping() *ReadonlyMemberGrouping {
	if s.storage == nil {
		return nil
	}

	if g := s.storage.Grouping(); g != nil {
		return MustNewReadonlyMemberGrouping(g)
	}
	return nil
}

// IterateByOffset iterates members starting at the given offset.
// The callback can return true to stop iteration.
// It doesn't iterate grouped members, if you need to iterate grouped
// members use `ReadonlyMemberStorage.Grouping().IterateByOffset()`.
func (s ReadonlyMemberStorage) IterateByOffset(offset, count int, fn MemberIterFn) bool {
	if s.storage != nil {
		return s.storage.IterateByOffset(offset, count, fn)
	}
	return false
}
