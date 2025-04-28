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
	rid, tid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board")
	tid = cross(boards2.CreateThread)(bid, "Foo", "bar")
	rid = cross(boards2.CreateReply)(bid, tid, 0, "body")

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.FlagReply)(bid, tid, rid, "")

	// Render content must contain a message about the hidden reply
	content := cross(boards2.Render)("test-board/1/2")
	println(strings.Contains(content, "\n> _⚠ Reply is hidden as it has been flagged as inappropriate_\n"))
}

// Output:
// true
