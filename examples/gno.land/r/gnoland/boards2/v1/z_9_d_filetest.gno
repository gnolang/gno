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

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board", false)
	pid = cross(boards2.CreateThread)(bid, "Foo", "bar", false)
}

func main() {
	// Call using a user that has not permission to delete threads
	testing.SetRealm(std.NewUserRealm(user))

	cross(boards2.DeleteThread)(bid, pid)
}

// Error:
// unauthorized
