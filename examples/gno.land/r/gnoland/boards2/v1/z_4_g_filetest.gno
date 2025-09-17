package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	admin = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var bid boards2.BoardID // Operate on board DAO

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false)

	boards2.InviteMember(cross, bid, admin, boards2.RoleGuest)
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	boards2.ChangeMemberRole(cross, bid, "foo", boards2.RoleModerator)
}

// Error:
// invalid member address: foo
