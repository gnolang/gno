// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_change_member_role_01_filetest

package z_change_member_role_01_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	member  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	newRole         = boards2.RoleAdmin
)

var bid boards.ID // Operate on board DAO

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross2(cur), "foo123", false, false)
	boards2.InviteMember(cross2(cur), bid, member, boards2.RoleGuest)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.ChangeMemberRole(cross2(cur), bid, member, newRole)

	// Ensure that new role has been changed
	println(boards2.HasMemberRole(bid, member, newRole))
}

// Output:
// true
