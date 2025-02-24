package main

import (
	"std"
	"strings"

	boards2 "gno.land/r/nt/boards2/v1"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	body  = "Test reply"
	path  = "test-board/1/2"
)

var (
	bid      boards2.BoardID
	tid, rid boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("test-board")
	tid = boards2.CreateThread(bid, "Foo", "bar")

	// Create a reply and a sub reply
	parentRID := boards2.CreateReply(bid, tid, 0, "Parent")
	rid = boards2.CreateReply(bid, tid, parentRID, "Child")
}

func main() {
	boards2.EditReply(bid, tid, rid, body)

	// Render content must contain the modified reply
	content := boards2.Render(path)
	println(strings.Contains(content, "\n> > "+body+"\n"))
}

// Output:
// true
