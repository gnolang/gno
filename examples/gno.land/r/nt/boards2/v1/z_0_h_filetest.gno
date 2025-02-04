package main

// SEND: 200000000ugnot

import (
	"std"

	"gno.land/r/demo/users"
	"gno.land/r/nt/boards2"
)

const (
	owner  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	member = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	name   = "test123"
)

func init() {
	std.TestSetOrigCaller(owner)

	// Test1 is the boards owner and its address has a user already registered
	// so a new member must register a user with the new board name.
	boards2.InviteMember(0, member, boards2.RoleOwner) // Operate on realm DAO members instead of individual boards
	std.TestSetOrigCaller(member)
	users.Register("", name, "")
}

func main() {
	bid := boards2.CreateBoard(name)
	println("ID =", bid)
}

// Output:
// ID = 1
