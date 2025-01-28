package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const (
	owner       = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	member      = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	initialRole = boards2.RoleGuest
	newRole     = boards2.RoleAdmin
	bid         = boards2.BoardID(0) // Operate on realm DAO members instead of individual boards
)

func init() {
	std.TestSetOrigCaller(owner)
	boards2.InviteMember(bid, member, boards2.RoleGuest)
}

func main() {
	if boards2.HasMemberRole(bid, member, initialRole) {
		println("ok")
	}

	boards2.ChangeMemberRole(bid, member, newRole)

	if boards2.HasMemberRole(bid, member, newRole) {
		println("ok")
	}
}

// Output:
// ok
// ok
