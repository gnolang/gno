package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
	bid           = boards.ID(0)                               // Operate on realm DAO instead of individual boards
	role          = boards2.RoleOwner
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	boards2.InviteMember(cross, bid, user, role)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.InviteMember(cross, bid, user, role)
}

// Error:
// user is already a member: g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj
