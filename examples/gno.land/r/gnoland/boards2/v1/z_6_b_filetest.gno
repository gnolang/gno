package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
	"gno.land/r/gnoland/boards2/v1/permissions"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	bid           = boards.ID(0)                               // Operate on realm DAO instead of individual boards
	role          = permissions.RoleGuest
)

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	println(boards2.HasMemberRole(bid, user, role))
	println(boards2.IsMember(bid, user))
}

// Output:
// false
// false
