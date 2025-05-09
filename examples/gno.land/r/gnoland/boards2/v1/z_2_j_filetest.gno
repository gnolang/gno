package main

import (
	"std"
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	path    = "test-board/1/2"
	comment = "Second comment"
)

var (
	bid      boards2.BoardID
	tid, rid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board")
	tid = cross(boards2.CreateThread)(bid, "Foo", "bar")
	rid = cross(boards2.CreateReply)(bid, tid, 0, "First comment")
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	rid2 := cross(boards2.CreateReply)(bid, tid, rid, comment)

	// Ensure that returned ID is right
	println(rid2 == 3)

	// Render content must contain the sub-reply
	content := boards2.Render(path)
	println(strings.Contains(content, "\n> > "+comment+"\n"))
}

// Output:
// true
// true
