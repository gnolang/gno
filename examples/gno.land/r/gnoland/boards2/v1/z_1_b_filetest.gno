package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	admin address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	user  address = "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"
)

var bid boards.ID // Operate on board DAO

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test123", false, false)

	// Add an admin member
	boards2.InviteMember(cross, bid, admin, boards2.RoleAdmin)
}

func main() {
	// Next call will be done by the admin member
	testing.SetRealm(testing.NewUserRealm(admin))

	boards2.InviteMember(cross, bid, user, boards2.RoleOwner)
}

// Error:
// only owners are allowed to invite other owners
