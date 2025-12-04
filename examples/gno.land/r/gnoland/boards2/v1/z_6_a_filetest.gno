package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner  address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	member address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	role           = boards2.RoleGuest
)

var bid boards.ID // Operate on board DAO

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false)

	boards2.InviteMember(cross, bid, member, role)
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
