package main

// SEND: 1000000ugnot

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
	users "gno.land/r/gnoland/users/v1"
)

const (
	owner  address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"
	member address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	name           = "test123"
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Test1 is the boards owner and its address has a user already registered
	// so a new member must register a user with the new board name.
	boards2.InviteMember(cross, 0, member, boards2.RoleOwner) // Operate on realm DAO members instead of individual boards

	testing.SetRealm(testing.NewUserRealm(member))
	users.Register(cross, name)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(member))

	bid := boards2.CreateBoard(cross, name, false, false)
	println("ID =", bid)
}

// Output:
// ID = 1
