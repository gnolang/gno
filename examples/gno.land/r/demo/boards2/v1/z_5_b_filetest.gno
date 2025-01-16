package main

import (
	"std"

	boards2 "gno.land/r/demo/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	bid   = boards2.BoardID(0)                                      // Operate on realm DAO members instead of individual boards
)

func init() {
	std.TestSetOrigCaller(owner)
	boards2.InviteMember(bid, admin, boards2.RoleAdmin)

	std.TestSetOrigCaller(admin)
}

func main() {
	boards2.FreezeRealm(true)
}

// Error:
// unauthorized
