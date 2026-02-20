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
	boards2.InviteMember(cross, bid, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", boards2.RoleOwner)
	boards2.InviteMember(cross, bid, "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5", boards2.RoleAdmin)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.IterateRealmMembers(0, func(u boards.User) bool {
		println(u.Address, string(u.Roles[0]))
		return false
	})
}

// Output:
// g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq owner
// g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj owner
// g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5 admin
