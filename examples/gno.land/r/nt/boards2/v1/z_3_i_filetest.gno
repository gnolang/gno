package main

// SEND: 1000000ugnot

import (
	"std"

	users "gno.land/r/gnoland/users/v1"
	boards2 "gno.land/r/nt/boards2/v1"
)

const (
	owner   = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	member  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	member2 = std.Address("g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5")
	name    = "foo123"
	newName = "barbaz123"
)

var bid boards2.BoardID // Operate on board DAO

func init() {
	std.TestSetOriginCaller(owner)

	bid = boards2.CreateBoard(name)
	boards2.InviteMember(bid, member, boards2.RoleOwner)

	// Test1 is the boards owner and its address has a user already registered
	// so a new member must register a user with the new board name.
	std.TestSetOriginCaller(member)
	users.Register(newName)

	// Invite a new member that doesn't own the user that matches the new board name
	boards2.InviteMember(bid, member2, boards2.RoleOwner)
	std.TestSetOriginCaller(member2)
}

func main() {
	boards2.RenameBoard(name, newName)
}

// Error:
// board name is a user name registered to a different user
