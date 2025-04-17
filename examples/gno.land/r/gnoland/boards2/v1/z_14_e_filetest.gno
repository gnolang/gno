package main

import (
	"std"
	"strings"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid      boards2.BoardID
	tid, rid boards2.PostID
)

func init() {
	testing.SetOriginCaller(owner)
	bid = boards2.CreateBoard("test-board")
	tid = boards2.CreateThread(bid, "Foo", "bar")

	rid = boards2.CreateReply(bid, tid, 0, "Parent")
	boards2.CreateReply(bid, tid, rid, "Child reply")
}

func main() {
	boards2.DeleteReply(bid, tid, rid)

	// Render content must contain the releted message instead of reply's body
	content := boards2.Render("test-board/1/2")
	println(strings.Contains(content, "\n> This reply has been deleted\n"))
	println(strings.Contains(content, "\n> > Child reply\n"))
}

// Output:
// true
// true
