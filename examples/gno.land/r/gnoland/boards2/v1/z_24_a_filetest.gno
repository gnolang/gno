package main

import (
	"std"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	std.TestSetOriginCaller(owner)

	bid := boards2.CreateBoard("test123")
	boards2.InviteMember(bid, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", boards2.RoleAdmin)

	bid = boards2.CreateBoard("foo123")
	boards2.InviteMember(bid, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", boards2.RoleOwner)
	boards2.InviteMember(bid, "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5", boards2.RoleAdmin)
}

func main() {
	boards2.IterateBoards(0, func(b boards2.ReadonlyBoard) bool {
		println(b.ID().String())
		println(b.Name())

		b.IterateMembers(0, b.MembersCount(), func(u boards2.User) bool {
			println(u.Address, string(u.Roles[0]))
			return false
		})
	})
}

// Output:
// 1
// test123
// g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 owner
// g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj admin
// 2
// foo123
// g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 owner
// g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj owner
// g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5 admin
