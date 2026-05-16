// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_invite_member_04_filetest

package z_invite_member_04_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	role          = boards2.RoleOwner
)

var bid boards.ID

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross2(cur), "foo123", false, false) // Operate on board DAO members
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.InviteMember(cross2(cur), bid, user, role)

	// Check that user is invited
	println(boards2.HasMemberRole(0, user, role)) // Operate on realm DAO
	println(boards2.HasMemberRole(bid, user, role))
}

// Output:
// false
// true
