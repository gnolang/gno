package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	title = "Test Thread"
	body  = "Test body"
)

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board")
	pid = cross(boards2.CreateThread)(bid, title, body)

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.DeleteThread)(bid, pid)

	// Ensure thread doesn't exist
	println(cross(boards2.Render)("test-board/1"))
}

// Output:
// Thread does not exist with ID: 1
