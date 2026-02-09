package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/permissions"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var bid boards.ID // Operate on board DAO

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false, false)

	boards2.InviteMember(cross, bid, user, permissions.RoleGuest)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.RemoveMember(cross, bid, user)

	// Check that user is not a member
	println(boards2.IsMember(bid, user))
}

// Output:
// false
