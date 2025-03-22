package main

import (
	"std"
	"strings"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	bid      boards2.BoardID
	rid, tid boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("test-board")
	tid = boards2.CreateThread(bid, "Foo", "bar")
	rid = boards2.CreateReply(bid, tid, 0, "body")
}

func main() {
	boards2.FlagReply(bid, tid, rid, "")

	// Render content must contain a message about the hidden reply
	content := boards2.Render("test-board/1/2")
	println(strings.Contains(content, "\n> _⚠ Reply is hidden as it has been flagged as inappropriate_\n"))
}

// Output:
// true
