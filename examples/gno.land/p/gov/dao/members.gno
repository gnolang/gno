package dao

import (
	"std"
)

// MemberStore defines the member storage abstraction
type MemberStore interface {
	// GetMembers returns all members in the store
	GetMembers() []Member

	// AddMember attempts to add a member to the store
	AddMember() error

	// RemoveMember attempts to remove a member from the store
	RemoveMember() error

	// IsMember returns a flag indicating if the given address
	// belongs to a member
	IsMember(address std.Address) bool
}

// Member holds the relevant member information
type Member struct {
	Address std.Address // bech32 gno address of the member (unique)
	Name    string      // name associated with the member
}
