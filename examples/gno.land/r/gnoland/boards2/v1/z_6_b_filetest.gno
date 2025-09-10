package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	bid   = boards2.BoardID(0)                                      // Operate on realm DAO instead of individual boards
	role  = boards2.RoleGuest
)

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	println(boards2.HasMemberRole(bid, user, role))
	println(boards2.IsMember(bid, user))
}

// Output:
// false
// false
