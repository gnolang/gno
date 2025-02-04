package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	srcBid boards2.BoardID
	dstBid boards2.BoardID
	srcTid boards2.PostID
)

func init() {
	std.TestSetOrigCaller(owner)
	srcBid = boards2.CreateBoard("src-board")
	dstBid = boards2.CreateBoard("dst-board")

	srcTid = boards2.CreateThread(srcBid, "Foo", "bar")
	boards2.FlagThread(srcBid, srcTid, "idk")
}

func main() {
	// Repost should fail if source thread is flagged
	_ = boards2.CreateRepost(srcBid, srcTid, "foo", "bar", dstBid)
}

// Error:
// thread has been flagged as inappropriate
