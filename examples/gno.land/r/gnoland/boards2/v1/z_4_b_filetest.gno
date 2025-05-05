package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	member  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	newRole = boards2.RoleAdmin
)

var bid boards2.BoardID // Operate on board DAO

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("foo123")
	cross(boards2.InviteMember)(bid, member, boards2.RoleGuest)
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	cross(boards2.ChangeMemberRole)(bid, member, newRole)

	// Ensure that new role has been changed
	println(boards2.HasMemberRole(bid, member, newRole))
}

// Output:
// true
