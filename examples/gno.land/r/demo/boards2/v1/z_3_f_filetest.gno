package main

// SEND: 200000000ugnot

import (
	"std"

	"gno.land/r/demo/boards2"
	"gno.land/r/demo/users"
)

const (
	owner   = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	member  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	name    = "foo"
	newName = "barbaz"
)

func init() {
	std.TestSetOrigCaller(owner)

	// Test1 is the boards owner and its address has a user already registered
	// so a new member must register a user with the new board name.
	boards2.InviteMember(member, boards2.RoleOwner)
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
