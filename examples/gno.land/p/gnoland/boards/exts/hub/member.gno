package hub

import (
	"gno.land/p/gnoland/boards"
)

// Member defines a safe type for board members.
type Member struct {
	// address is the account address of the member.
	address address

	// roles contains the names of the roles assigned to the member.
	roles []string
}

// Address returns the account address of the member.
func (m Member) Address() address { return m.address }

// Roles returns the names of the roles assigned to the member.
func (m Member) Roles() []string { return append([]string(nil), m.roles...) }

// NewSafeMember creates a safe board member.
func NewSafeMember(ref boards.User) Member {
	roles := make([]string, len(ref.Roles))
	for i, r := range ref.Roles {
		roles[i] = string(r)
	}

	return Member{
		address: ref.Address,
		roles:   roles,
	}
}
