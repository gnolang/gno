package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

func init() {
	testing.SetRealm(std.NewUserRealm(owner))

	bid := boards2.CreateBoard(cross, "test123", false)
	boards2.InviteMember(cross, bid, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", boards2.RoleAdmin)

	bid = boards2.CreateBoard(cross, "foo123", false)
	boards2.InviteMember(cross, bid, "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj", boards2.RoleOwner)
	boards2.InviteMember(cross, bid, "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5", boards2.RoleAdmin)
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	boards2.IterateBoards(0, func(b *boards2.Board) bool {
		println(b.GetID().String())
		println(b.GetName())

		b.IterateMembers(0, b.MembersCount(), func(member std.Address, r []boards2.Role) {
			println(member, string(r[0]))
		})

		return false
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
