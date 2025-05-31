package main

// SEND: 5000000ugnot

import (
	"std"
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user    = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
	path    = "test-board/1/2"
	comment = "Test comment"
)

var (
	bid boards2.BoardID
	tid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board", false)

	// Create a public thread to allow uninvited users to add replies
	tid = cross(boards2.CreateThread)(bid, "Foo", "bar", true)
}

func main() {
	testing.SetRealm(std.NewUserRealm(user))

	rid := cross(boards2.CreateReply)(bid, tid, 0, comment)

	// Ensure that returned ID is right
	println(rid == 2)

	// Render content must contain the reply
	content := boards2.Render(path)
	println(strings.Contains(content, "\n> "+comment+"\n"))
}

// Output:
// true
// true
