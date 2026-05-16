// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_change_member_role_03_filetest

package z_change_member_role_03_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner  address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	admin  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	admin2 address = "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5"
)

var bid boards.ID // Operate on board DAO

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross2(cur), "test123", false, false)

	boards2.InviteMember(cross2(cur), bid, admin, boards2.RoleAdmin)
	boards2.InviteMember(cross2(cur), bid, admin2, boards2.RoleAdmin)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(admin))

	boards2.ChangeMemberRole(cross2(cur), bid, admin2, boards2.RoleOwner)
}

// Error:
// admins are not allowed to promote members to Owner
