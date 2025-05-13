package commondao

import (
	"std"

	"gno.land/p/moul/addrset"
)

// TODO: Support member groups and metadata

// MemberStorage defines an interface for member storages.
type MemberStorage interface {
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

	// IterateByOffset iterates members starting at the given offset.
	// The callback can return true to stop iteration.
	IterateByOffset(offset, count int, fn func(std.Address) bool)
}

// NewMemberStorage creates a new member storage.
func NewMemberStorage() MemberStorage {
	return &addrset.Set{}
}
