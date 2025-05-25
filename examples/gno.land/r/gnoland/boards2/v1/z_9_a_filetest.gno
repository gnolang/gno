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
	bid = cross(boards2.CreateBoard)("test-board", false)
	pid = cross(boards2.CreateThread)(bid, title, body, false)
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	cross(boards2.DeleteThread)(bid, pid)

	// Ensure thread doesn't exist
	println(boards2.Render("test-board/1"))
}

// Output:
// Thread does not exist with ID: 1
