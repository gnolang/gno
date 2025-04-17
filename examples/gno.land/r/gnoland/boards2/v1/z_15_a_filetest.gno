package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	srcBID boards2.BoardID
	dstBID boards2.BoardID
	srcTID boards2.PostID
)

func init() {
	testing.SetOriginCaller(owner)
	srcBID = boards2.CreateBoard("src-board")
	dstBID = boards2.CreateBoard("dst-board")

	srcTID = boards2.CreateThread(srcBID, "Foo", "bar")
	boards2.FlagThread(srcBID, srcTID, "idk")
}

func main() {
	// Repost should fail if source thread is flagged
	boards2.CreateRepost(srcBID, srcTID, "foo", "bar", dstBID)
}

// Error:
// thread is hidden
