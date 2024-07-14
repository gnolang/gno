package dao

import (
	"std"
)

// MemberStore defines the member storage abstraction
type MemberStore interface {
	// GetMembers returns all members in the store
	GetMembers() []Member

	// Size returns the current size of the store
	Size() int

	// IsMember returns a flag indicating if the given address
	// belongs to a member
	IsMember(address std.Address) bool

	// GetMember returns the requested member
	GetMember(address std.Address) (Member, error)

	// AddMember attempts to add a member to the store
	AddMember(member Member) error

	// RemoveMember attempts to remove a member from the store
	RemoveMember(address std.Address) error

	// UpdateMember attempts to update the member in the store
	UpdateMember(address std.Address, member Member) error
}

// Member holds the relevant member information
type Member struct {
	Address     std.Address // bech32 gno address of the member (unique)
	VotingPower uint64      // the voting power of the member
}
