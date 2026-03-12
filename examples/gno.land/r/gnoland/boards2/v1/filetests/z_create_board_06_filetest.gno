package main

// SEND: 1000000ugnot

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner  address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"
	member address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	name           = "test123"
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Operate on realm DAO members instead of individual boards
	boards2.InviteMember(cross, 0, member, boards2.RoleOwner)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(member))

	// Create a board as an invited realm member
	bid := boards2.CreateBoard(cross, name, false, false)
	println("ID =", bid)
}

// Output:
// ID = 1
