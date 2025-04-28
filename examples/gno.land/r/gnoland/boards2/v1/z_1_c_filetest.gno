package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	user  = std.Address("g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn")
	role  = boards2.RoleAdmin
)

var bid boards2.BoardID // Operate on board DAO

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test123")

	// Add an admin member
	cross(boards2.InviteMember)(bid, admin, boards2.RoleAdmin)

	// Next call will be done by the admin member
	testing.SetOriginCaller(admin)
}

func main() {
	cross(boards2.InviteMember)(bid, user, role)

	// Check that user is invited
	println(boards2.HasMemberRole(bid, user, role))
}

// Output:
// true
