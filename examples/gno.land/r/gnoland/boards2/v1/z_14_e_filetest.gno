package main

import (
	"std"
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	bid      boards2.BoardID
	tid, rid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board")
	tid = cross(boards2.CreateThread)(bid, "Foo", "bar")

	rid = cross(boards2.CreateReply)(bid, tid, 0, "Parent")
	cross(boards2.CreateReply)(bid, tid, rid, "Child reply")
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	cross(boards2.DeleteReply)(bid, tid, rid)

	// Render content must contain the releted message instead of reply's body
	content := boards2.Render("test-board/1/2")
	println(strings.Contains(content, "\n> This reply has been deleted\n"))
	println(strings.Contains(content, "\n> > Child reply\n"))
}

// Output:
// true
// true
