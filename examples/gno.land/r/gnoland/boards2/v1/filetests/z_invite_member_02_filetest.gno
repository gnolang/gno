// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_invite_member_02_filetest

package z_invite_member_02_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	admin address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	user  address = "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"
	role          = boards2.RoleAdmin
)

var bid boards.ID // Operate on board DAO

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross2(cur), "test123", false, false)

	// Add an admin member
	boards2.InviteMember(cross2(cur), bid, admin, boards2.RoleAdmin)
}

func main(cur realm) {
	// Next call will be done by the admin member
	testing.SetRealm(testing.NewUserRealm(admin))

	boards2.InviteMember(cross2(cur), bid, user, role)

	// Check that user is invited
	println(boards2.HasMemberRole(bid, user, role))
}

// Output:
// true
