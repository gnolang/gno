// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_is_member_00_filetest

package z_is_member_00_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner  address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	member address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	role           = boards2.RoleGuest
)

var bid boards.ID // Operate on board DAO

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross(cur), "test123", false, false)

	boards2.InviteMember(cross(cur), bid, member, role)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	println(boards2.HasMemberRole(bid, member, role))
	println(boards2.HasMemberRole(bid, member, "invalid"))
	println(boards2.IsMember(bid, member))
}

// Output:
// true
// false
// true
