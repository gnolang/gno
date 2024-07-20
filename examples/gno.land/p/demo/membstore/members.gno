package membstore

import (
	"std"
)

// MemberStore defines the member storage abstraction
type MemberStore interface {
	// Members returns all members in the store
	Members(offset, count uint64) []Member

	// Size returns the current size of the store
	Size() int

	// IsMember returns a flag indicating if the given address
	// belongs to a member
	IsMember(address std.Address) bool

	// TotalPower returns the total voting power of the member store
	TotalPower() uint64

	// Member returns the requested member
	Member(address std.Address) (Member, error)

	// AddMember adds a member to the store
	AddMember(member Member) error

	// UpdateMember updates the member in the store.
	// If updating a member's voting power to 0,
	// the member will be removed
	UpdateMember(address std.Address, member Member) error
}

// Member holds the relevant member information
type Member struct {
	Address     std.Address // bech32 gno address of the member (unique)
	VotingPower uint64      // the voting power of the member
}
