package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	bid           = boards.ID(0)                               // Operate on realm DAO instead of individual boards
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.InviteMember(cross, bid, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", boards2.RoleAdmin)
}

func main() {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gnoland/boards2/test"))

	// Calling as a Boards2 sub realm should succeed
	perms := boards2.GetRealmPermissions()
	perms.IterateUsers(0, perms.UsersCount(), func(u boards.User) bool {
		println(u.Address, string(u.Roles[0]))
		return false
	})
}

// Output:
// g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq owner
// g1manfred47kzduec920z88wfr64ylksmdcedlf5 owner
// g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj admin
