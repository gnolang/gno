package main

import (
	"std"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	bid   = boards2.BoardID(0)                                      // Operate on realm DAO instead of individual boards
)

func init() {
	std.TestSetOriginCaller(owner)
	boards2.InviteMember(bid, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", boards2.RoleOwner)
	boards2.InviteMember(bid, "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5", boards2.RoleAdmin)
}

func main() {
	boards2.IterateRealmMembers(0, func(u boards2.User) bool {
		println(u.Address, string(u.Roles[0]))
		return false
	})
}

// Output:
// g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 owner
// g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj owner
// g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5 admin
