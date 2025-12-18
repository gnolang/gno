package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner  address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	admin  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	admin2 address = "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5"
)

var bid boards.ID // Operate on board DAO

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false)

	boards2.InviteMember(cross, bid, admin, boards2.RoleAdmin)
	boards2.InviteMember(cross, bid, admin2, boards2.RoleAdmin)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(admin))

	boards2.ChangeMemberRole(cross, bid, admin2, boards2.RoleOwner)
}

// Error:
// admins are not allowed to promote members to Owner
