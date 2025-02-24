package main

import (
	"std"
	"strings"

	boards2 "gno.land/r/nt/boards2/v1"
)

const (
	owner   = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	path    = "test-board/1/2"
	comment = "Test comment"
)

var (
	bid boards2.BoardID
	tid boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)
	bid = boards2.CreateBoard("test-board")
	tid = boards2.CreateThread(bid, "Foo", "bar")
}

func main() {
	rid := boards2.CreateReply(bid, tid, 0, comment)

	// Ensure that returned ID is right
	println(rid == 2)

	// Render content must contain the reply
	content := boards2.Render(path)
	println(strings.Contains(content, "\n> "+comment+"\n"))
}

// Output:
// true
// true
