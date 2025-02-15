package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var bid boards2.BoardID // Operate on board DAO

func init() {
	std.TestSetOrigCaller(owner)
	bid = boards2.CreateBoard("test123")

	boards2.InviteMember(bid, user, boards2.RoleGuest)
}

func main() {
	boards2.RemoveMember(bid, user)

	// Check that user is not a member
	println(boards2.IsMember(bid, user))
}

// Output:
// false
