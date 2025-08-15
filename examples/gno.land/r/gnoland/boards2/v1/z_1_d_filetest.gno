package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	boards2.InviteMember(cross, 0, user, boards2.Role("foobar")) // Operate on realm DAO instead of individual boards
}

// Error:
// invalid role: foobar
