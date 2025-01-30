package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const (
	owner   = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	member  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	bid     = boards2.BoardID(0)                                      // Operate on realm DAO members instead of individual boards
	newRole = boards2.RoleOwner
)

func init() {
	std.TestSetOrigCaller(owner)
	boards2.InviteMember(bid, member, boards2.RoleAdmin)
}

func main() {
	boards2.ChangeMemberRole(bid, member, newRole) // Owner can promote other members to Owner

	// Ensure that new role has been changed to owner
	println(boards2.HasMemberRole(bid, member, newRole))
}

// Output:
// true
