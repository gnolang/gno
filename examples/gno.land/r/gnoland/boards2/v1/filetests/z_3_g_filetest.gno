package main

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	uinit "gno.land/r/sys/users/init"
)

const (
	owner   address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	member  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	member2 address = "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5"
	name            = "foo123"
	newName         = "barbaz123"
)

var bid boards.ID // Operate on board DAO

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, name, false, false)
	boards2.InviteMember(cross, bid, member, boards2.RoleOwner)

	// Test1 is the boards owner and its address has a user already registered
	// so a new member must register a user with the new board name.
	uinit.RegisterUser(cross, newName, member)

	// Invite a new member that doesn't own the user that matches the new board name
	boards2.InviteMember(cross, bid, member2, boards2.RoleOwner)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(member2))

	boards2.RenameBoard(cross, name, newName)
}

// Error:
// board name is a user name registered to a different user
