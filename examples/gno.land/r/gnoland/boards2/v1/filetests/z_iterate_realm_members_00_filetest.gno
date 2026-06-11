// PKGPATH: gno.land/r/gnoland/boards2/v1/filetests/z_iterate_realm_members_00_filetest

package z_iterate_realm_members_00_filetest

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	bid           = boards.ID(0) // Operate on realm DAO instead of individual boards
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.InviteMember(cross(cur), bid, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", boards2.RoleAdmin)
}

func main(cur realm) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))

	perms := boards2.GetRealmPermissions()
	perms.IterateUsers(0, perms.UsersCount(), func(u boards.User) bool {
		println(u.Address, string(u.Roles[0]))
		return false
	})
}

// Output:
// g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh owner
// g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj admin
