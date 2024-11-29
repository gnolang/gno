package main

import (
	"std"

	"gno.land/r/demo/boards2"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	user  = std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
)

func init() {
	// Add an admin user
	std.TestSetOrigCaller(owner)
	boards2.InviteMember(admin, boards2.RoleAdmin)

	// Next call will be done by the admin user
	std.TestSetOrigCaller(admin)
}

func main() {
	boards2.InviteMember(user, boards2.RoleOwner)
}

// Error:
// only owners are allowed to invite other owners
