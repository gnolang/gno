package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	member  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	bid             = boards.ID(0)                               // Operate on realm DAO members instead of individual boards
	newRole         = boards2.RoleOwner
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.InviteMember(cross, bid, member, boards2.RoleAdmin)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.ChangeMemberRole(cross, bid, member, newRole) // Owner can promote other members to Owner

	// Ensure that new role has been changed to owner
	println(boards2.HasMemberRole(bid, member, newRole))
}

// Output:
// true
