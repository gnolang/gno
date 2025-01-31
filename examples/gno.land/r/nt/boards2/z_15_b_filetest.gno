package main

import (
	"std"
	"strings"

	"gno.land/r/nt/boards2"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	srcBid boards2.BoardID
	dstBid boards2.BoardID
	srcTid boards2.PostID = 1024
)

func init() {
	std.TestSetOrigCaller(owner)
	srcBid = boards2.CreateBoard("src-board")
	dstBid = boards2.CreateBoard("dst-board")
}

func main() {
	// Repost should fail if source thread doesn't exist
	_ = boards2.CreateRepost(srcBid, srcTid, "foo", "bar", dstBid)
}

// Error:
// thread does not exist with ID: 1024
