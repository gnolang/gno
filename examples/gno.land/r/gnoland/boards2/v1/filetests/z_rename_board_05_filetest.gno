package main

// SEND: 1000000ugnot

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   address = "g1rp7cmetn27eqlpjpc4vuusf8kaj746tysc0qgh"
	member  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	name            = "foo123"
	newName         = "barbaz123"
)

var bid boards.ID // Operate on board DAO

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))

	bid = boards2.CreateBoard(cross, name, false, false)
	boards2.InviteMember(cross, bid, member, boards2.RoleOwner)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(member))

	boards2.RenameBoard(cross, name, newName)

	// Ensure board is renamed by another board owner
	bid2, _ := boards2.GetBoardIDFromName(cross, newName)
	println("IDs match =", bid == bid2)
}

// Output:
// IDs match = true
