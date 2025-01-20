package main

import (
	"std"

	"gno.land/r/demo/boards2"
)

const (
	owner       = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	member      = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	initialRole = boards2.RoleGuest
	newRole     = boards2.RoleAdmin
)

var bid boards2.BoardID

func init() {
	std.TestSetOrigCaller(owner)
	bid = boards2.CreateBoard("foo123") // Operate on board DAO members
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
