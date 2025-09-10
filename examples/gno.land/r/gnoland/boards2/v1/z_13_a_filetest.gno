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
	bid = boards2.CreateBoard(cross, "test-board", false)
	tid = boards2.CreateThread(cross, bid, "Foo", "bar")
	rid = boards2.CreateReply(cross, bid, tid, 0, "body")
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	boards2.FlagReply(cross, bid, tid, rid, "")

	// Render content must contain a message about the hidden reply
	content := boards2.Render("test-board/1/2")
	println(strings.Contains(content, "\n> *⚠ Reply is hidden as it has been flagged as inappropriate*\n"))
}

// Output:
// true
