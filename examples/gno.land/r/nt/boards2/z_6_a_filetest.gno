package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const (
	owner  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	member = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	bid    = boards2.BoardID(0)                                      // Operate on realm DAO instead of individual boards
	role   = boards2.RoleGuest
)

func init() {
	std.TestSetOrigCaller(owner)
	boards2.InviteMember(bid, member, role)
}

func main() {
	println(boards2.HasMemberRole(bid, member, role))
	println(boards2.HasMemberRole(bid, member, "invalid"))
	println(boards2.IsMember(bid, member))
}

// Output:
// true
// false
// true
