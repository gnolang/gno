package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const (
	owner   = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	member  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	newRole = boards2.RoleAdmin
)

var bid boards2.BoardID // Operate on board DAO

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("foo123")
	boards2.InviteMember(bid, member, boards2.RoleGuest)
}

func main() {
	boards2.ChangeMemberRole(bid, member, newRole)

	// Ensure that new role has been changed
	println(boards2.HasMemberRole(bid, member, newRole))
}

// Output:
// true
