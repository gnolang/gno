// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_rename_board_08_filetest

package z_rename_board_08_filetest

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	uinit "gno.land/r/sys/users/init"
)

const (
	owner   address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	member  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	member2 address = "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5"
	name            = "foo123"
	newName         = "barbaz123"
)

var bid boards.ID // Operate on board DAO

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), name, false, false)
	boards2.InviteMember(cross(cur), bid, member, boards2.RoleOwner)

	// Test1 is the boards owner and its address has a user already registered
	// so a new member must register a user with the new board name.
	// uinit.RegisterUser is genesis-only since the security fix.
	testing.SetHeight(0)
	uinit.RegisterUser(cross(cur), newName, member)
	testing.SetHeight(123)

	// Invite a new member that doesn't own the user that matches the new board name
	boards2.InviteMember(cross(cur), bid, member2, boards2.RoleOwner)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(member2))

	boards2.RenameBoard(cross(cur), name, newName)
}

// Error:
// board name is a user name registered to a different user
