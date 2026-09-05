package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g1sze988ga0a7sj5583cu3xt6m4vkxru4uwh6dmf"
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	bid           = boards.ID(0)                               // Operate on realm DAO instead of individual boards
	role          = boards2.RoleGuest
)

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	println(boards2.HasMemberRole(bid, user, role))
	println(boards2.IsMember(bid, user))
}

// Output:
// false
// false
