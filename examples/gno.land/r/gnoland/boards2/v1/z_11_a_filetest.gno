package main

import (
	"std"
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	title = "Test Thread"
	body  = "Test body"
	path  = "test-board/1"
)

var (
	bid boards2.BoardID
	pid boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	bid = cross(boards2.CreateBoard)("test-board")
	pid = cross(boards2.CreateThread)(bid, "Foo", "bar")

	testing.SetOriginCaller(owner)
}

func main() {
	cross(boards2.EditThread)(bid, pid, title, body)

	// Render content must contains thread's title and body
	content := cross(boards2.Render)(path)
	println(strings.HasPrefix(content, "## "+title))
	println(strings.Contains(content, body))
}

// Output:
// true
// true
