package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/permissions"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	admin address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

var bid boards.ID // Operate on board DAO

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false, false)

	boards2.InviteMember(cross, bid, admin, permissions.RoleGuest)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.ChangeMemberRole(cross, bid, admin, boards.Role("foo"))
}

// Error:
// invalid role: foo
