package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"
	user  address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj" // @test2
)

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.InviteMember(cross, 0, user, boards.Role("foobar")) // Operate on realm DAO instead of individual boards
}

// Error:
// invalid role: foobar
