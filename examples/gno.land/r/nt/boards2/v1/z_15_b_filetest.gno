package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	srcBid boards2.BoardID
	dstBid boards2.BoardID
	srcTid boards2.PostID = 1024
)

func init() {
	std.TestSetOriginCaller(owner)
	srcBid = boards2.CreateBoard("src-board")
	dstBid = boards2.CreateBoard("dst-board")
}

func main() {
	// Repost should fail if source thread doesn't exist
	_ = boards2.CreateRepost(srcBid, srcTid, "foo", "bar", dstBid)
}

// Error:
// thread does not exist with ID: 1024
