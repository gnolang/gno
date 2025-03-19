package main

import (
	"std"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	srcBID boards2.BoardID
	dstBID boards2.BoardID
	srcTID boards2.PostID = 1024
)

func init() {
	std.TestSetOriginCaller(owner)
	srcBID = boards2.CreateBoard("src-board")
	dstBID = boards2.CreateBoard("dst-board")
}

func main() {
	// Repost should fail if source thread doesn't exist
	boards2.CreateRepost(srcBID, srcTID, "foo", "bar", dstBID)
}

// Error:
// thread does not exist with ID: 1024
