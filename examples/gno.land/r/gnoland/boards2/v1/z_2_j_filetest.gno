package main

import (
	"std"
	"strings"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner   = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	path    = "test-board/1/2"
	comment = "Second comment"
)

var (
	bid      boards2.BoardID
	tid, rid boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("test-board")
	tid = boards2.CreateThread(bid, "Foo", "bar")
	rid = boards2.CreateReply(bid, tid, 0, "First comment")
}

func main() {
	rid2 := boards2.CreateReply(bid, tid, rid, comment)

	// Ensure that returned ID is right
	println(rid2 == 3)

	// Render content must contain the sub-reply
	content := boards2.Render(path)
	println(strings.Contains(content, "\n> > "+comment+"\n"))
}

// Output:
// true
// true
