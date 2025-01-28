package main

// SEND: 200000000ugnot

import (
	"std"

	"gno.land/r/demo/users"
	"gno.land/r/nt/boards2"
)

const (
	owner   = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	member  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	bid     = boards2.BoardID(0)                                      // Operate on realm DAO members instead of individual boards
	name    = "foo123"
	newName = "barbaz"
)

func init() {
	std.TestSetOrigCaller(owner)

	// Test1 is the boards owner and its address has a user already registered
	// so a new member must register a user with the new board name.
	boards2.InviteMember(bid, member, boards2.RoleOwner)
	std.TestSetOrigCaller(member)
	users.Register("", newName, "")

	boards2.CreateBoard(name)
}

func main() {
	_, exists := boards2.GetBoardIDFromName(newName)
	println("Exists =", exists)

	boards2.RenameBoard(name, newName)

	bid, _ := boards2.GetBoardIDFromName(newName)
	println("ID =", bid)
}

// Output:
// Exists = false
// ID = 1
