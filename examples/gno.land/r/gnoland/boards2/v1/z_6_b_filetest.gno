package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	bid   = boards2.BoardID(0)                                      // Operate on realm DAO instead of individual boards
	role  = boards2.RoleGuest
)

func init() {
	testing.SetOriginCaller(owner)
}

func main() {
	println(boards2.HasMemberRole(bid, user, role))
	println(boards2.IsMember(bid, user))
}

// Output:
// false
// false
