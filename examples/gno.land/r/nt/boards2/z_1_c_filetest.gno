package main

import (
	"std"

	"gno.land/r/nt/boards2"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	user  = std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
	bid   = boards2.BoardID(0) // Operate on realm DAO instead of individual boards
	role  = boards2.RoleAdmin
)

func init() {
	// Add an admin member
	std.TestSetOrigCaller(owner)
	boards2.InviteMember(bid, admin, boards2.RoleAdmin)

	// Next call will be done by the admin member
	std.TestSetOrigCaller(admin)
}

func main() {
	boards2.InviteMember(bid, user, role)

	// Check that user is invited
	println(boards2.HasMemberRole(bid, user, role))
}

// Output:
// true
