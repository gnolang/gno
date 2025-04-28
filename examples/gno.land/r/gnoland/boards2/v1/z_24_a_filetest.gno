package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func init() {
	testing.SetRealm(std.NewUserRealm(owner))

	bid := cross(boards2.CreateBoard)("test123")
	cross(boards2.InviteMember)(bid, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", boards2.RoleAdmin)

	bid = cross(boards2.CreateBoard)("foo123")
	cross(boards2.InviteMember)(bid, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", boards2.RoleOwner)
	cross(boards2.InviteMember)(bid, "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5", boards2.RoleAdmin)

	testing.SetOriginCaller(owner)
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
// g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq owner
// g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj admin
// 2
// foo123
// g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq owner
// g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj owner
// g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5 admin
