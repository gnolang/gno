package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	member  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	newRole         = boards2.RoleAdmin
)

var bid boards.ID // Operate on board DAO

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "foo123", false)
	boards2.InviteMember(cross, bid, member, boards2.RoleGuest)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.ChangeMemberRole(cross, bid, member, newRole)

	// Ensure that new role has been changed
	println(boards2.HasMemberRole(bid, member, newRole))
}

// Output:
// true
