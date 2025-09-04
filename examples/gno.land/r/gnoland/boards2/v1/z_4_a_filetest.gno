package main

import (
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	member  = address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	newRole = boards2.RoleOwner
	bid     = boards2.BoardID(0) // Operate on realm DAO instead of individual boards
)

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	boards2.InviteMember(cross, bid, member, boards2.RoleAdmin)
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.ChangeMemberRole(cross, bid, member, newRole)

	// Ensure that new role has been changed
	println(boards2.HasMemberRole(bid, member, newRole))
}

// Output:
// true
