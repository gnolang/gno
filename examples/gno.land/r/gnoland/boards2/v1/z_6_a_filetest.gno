package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	member = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	role   = boards2.RoleGuest
)

var bid boards2.BoardID // Operate on board DAO

func init() {
	testing.SetOriginCaller(owner)
	bid = boards2.CreateBoard("test123")

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
