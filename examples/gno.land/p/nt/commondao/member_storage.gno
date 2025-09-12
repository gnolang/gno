package commondao

import (
	"errors"
	"std"

	"gno.land/p/moul/addrset"
)

type (
	// MemberIterFn defines a callback to iterate DAO members.
	MemberIterFn func(std.Address) bool

	// MemberStorage defines an interface for member storages.
	MemberStorage interface {
		// Size returns the number of members in the storage.
		Size() int

		// Has checks if a member exists in the storage.
		Has(std.Address) bool

		// Add adds a member to the storage.
		// Returns true if the member is added, or false if it already existed.
		Add(std.Address) bool

		// Remove removes a member from the storage.
		// Returns true if member was removed, or false if it was not found.
		Remove(std.Address) bool

		// Grouping returns member groups when supported.
		// When nil is returned it means that grouping of members is not supported.
		// Member groups can be used by implementations that require grouping users
		// by roles or by tiers for example.
		Grouping() MemberGrouping

		// IterateByOffset iterates members starting at the given offset.
		// The callback can return true to stop iteration.
		IterateByOffset(offset, count int, fn MemberIterFn)
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

// IterateByOffset iterates members starting at the given offset.
// The callback can return true to stop iteration.
func (s memberStorage) IterateByOffset(offset, count int, fn MemberIterFn) {
	s.Set.IterateByOffset(offset, count, fn)
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

// Size returns the number of members in the storage.
func (s ReadonlyMemberStorage) Size() int {
	if s.storage == nil {
		return 0
	}
	return s.storage.Size()
}

// Has checks if a member exists in the storage.
func (s ReadonlyMemberStorage) Has(member std.Address) bool {
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
func (s ReadonlyMemberStorage) IterateByOffset(offset, count int, fn MemberIterFn) {
	if s.storage != nil {
		s.storage.IterateByOffset(offset, count, fn)
	}
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
