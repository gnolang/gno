package main

import (
	"std"
	"strings"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

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
	boards2.CreateReply(bid, tid, 0, strings.Repeat("x", boards2.MaxReplyLength+1))
}

// Error:
// reply is too long, maximum allowed is 300 characters
