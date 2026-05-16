// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_change_member_role_04_filetest

package z_change_member_role_04_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	member  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	bid             = boards.ID(0)                               // Operate on realm DAO members instead of individual boards
	newRole         = boards2.RoleOwner
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.InviteMember(cross2(cur), bid, member, boards2.RoleAdmin)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.ChangeMemberRole(cross2(cur), bid, member, newRole) // Owner can promote other members to Owner

	// Ensure that new role has been changed to owner
	println(boards2.HasMemberRole(bid, member, newRole))
}

// Output:
// true
