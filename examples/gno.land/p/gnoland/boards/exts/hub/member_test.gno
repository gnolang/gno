package hub_test

import (
	"testing"

	"gno.land/p/gnoland/boards"
	"gno.land/p/gnoland/boards/exts/hub"
	"gno.land/p/nt/urequire/v0"
)

func TestNewSafeMember(t *testing.T) {
	ref := boards.User{
		Address: alice,
		Roles:   []boards.Role{"owner", "admin"},
	}

	member := hub.NewSafeMember(ref)

	urequire.Equal(t, alice.String(), member.Address().String(), "expect address to match")
	urequire.Equal(t, 2, len(member.Roles()), "expect role count to match")
	urequire.Equal(t, "owner", member.Roles()[0], "expect first role to match")
	urequire.Equal(t, "admin", member.Roles()[1], "expect second role to match")
}

func TestNewSafeMemberWithoutRoles(t *testing.T) {
	member := hub.NewSafeMember(boards.User{Address: bob})

	urequire.Equal(t, bob.String(), member.Address().String(), "expect address to match")
	urequire.Equal(t, 0, len(member.Roles()), "expect no roles")
}

func TestSafeMemberIsASnapshot(t *testing.T) {
	roles := []boards.Role{"owner"}
	member := hub.NewSafeMember(boards.User{Address: alice, Roles: roles})

	// Mutating the source must not change what the member reports.
	roles[0] = "guest"
	urequire.Equal(t, "owner", member.Roles()[0], "expect source mutation to be ignored")

	// Mutating a returned slice must not change what the member reports.
	member.Roles()[0] = "guest"
	urequire.Equal(t, "owner", member.Roles()[0], "expect returned slice to be a copy")
}
