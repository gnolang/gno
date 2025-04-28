package main

import (
	"std"
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	body  = "Test reply"
	path  = "test-board/1/2"
)

var (
	bid      boards2.BoardID
	tid, rid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board")
	tid = cross(boards2.CreateThread)(bid, "Foo", "bar")
	rid = cross(boards2.CreateReply)(bid, tid, 0, "body")

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.EditReply)(bid, tid, rid, body)

	// Render content must contain the modified reply
	content := cross(boards2.Render)(path)
	println(strings.Contains(content, "\n> "+body+"\n"))
}

// Output:
// true
