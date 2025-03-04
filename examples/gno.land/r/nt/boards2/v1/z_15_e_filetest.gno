package main

import (
	"std"

	boards2 "gno.land/r/nt/boards2/v1"
)

const owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

var (
	srcBID boards2.BoardID
	dstBID boards2.BoardID
	srcTID boards2.PostID
)

func init() {
	std.TestSetOriginCaller(owner)

	// Create a board with a thread
	origBID := boards2.CreateBoard("origin-board")
	origTID := boards2.CreateThread(origBID, "title", "text")

	// Create a second board and repost a thread using an empty title
	srcBID = boards2.CreateBoard("source-board")
	srcTID = boards2.CreateRepost(origBID, origTID, "", "original text", srcBID)

	// Create a third board to try reposting the repost
	dstBID = boards2.CreateBoard("destination-board")
}

func main() {
	boards2.CreateRepost(srcBID, srcTID, "repost title", "repost text", dstBID)
}

// Error:
// reposting a thread that is a repost is not allowed
