package main

import (
	"strings"
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"
	body          = "Test reply"
	path          = "test-board/1/2"
)

var (
	bid      boards.ID
	tid, rid boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	bid = boards2.CreateBoard(cross, "test-board", false, false)
	tid = boards2.CreateThread(cross, bid, "Foo", "bar")

	// Create a reply and a sub reply
	parentRID := boards2.CreateReply(cross, bid, tid, 0, "Parent")
	rid = boards2.CreateReply(cross, bid, tid, parentRID, "Child")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.EditReply(cross, bid, tid, rid, body)

	// Render content must contain the modified reply
	content := boards2.Render(path)
	println(strings.Contains(content, "\n> > "+body+"\n"))
}

// Output:
// true
