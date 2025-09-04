package main

import (
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	role  = boards2.RoleOwner
)

var bid boards2.BoardID

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "foo123", false) // Operate on board DAO members
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.InviteMember(cross, bid, user, role)

	// Check that user is invited
	println(boards2.HasMemberRole(0, user, role)) // Operate on realm DAO
	println(boards2.HasMemberRole(bid, user, role))
}

// Output:
// false
// true
