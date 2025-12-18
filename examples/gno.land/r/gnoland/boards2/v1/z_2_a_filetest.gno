package main

import (
	"strings"
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	path            = "test-board/1/2"
	comment         = "Test comment"
)

var (
	bid boards.ID
	tid boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false)
	tid = boards2.CreateThread(cross, bid, "Foo", "bar")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	rid := boards2.CreateReply(cross, bid, tid, 0, comment)

	// Ensure that returned ID is right
	println(rid == 2)

	// Render content must contain the reply
	content := boards2.Render(path)
	println(strings.Contains(content, "\n> "+comment+"\n"))
}

// Output:
// true
// true
