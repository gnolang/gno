package main

import (
	"std"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	bid   = boards2.BoardID(0)                                      // Operate on realm DAO instead of individual boards
	role  = boards2.RoleOwner
)

func init() {
	std.TestSetOriginCaller(owner)
	boards2.InviteMember(bid, user, role)
}

func main() {
	boards2.InviteMember(bid, user, role)
}

// Error:
// user already exists
