package main

import (
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	srcBID boards2.BoardID
	dstBID boards2.BoardID
	srcTID boards2.PostID
)

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	// Create a board with a thread
	origBID := boards2.CreateBoard(cross, "origin-board", false)
	origTID := boards2.CreateThread(cross, origBID, "title", "text")

	// Create a second board and repost a thread using an empty title
	srcBID = boards2.CreateBoard(cross, "source-board", false)
	srcTID = boards2.CreateRepost(cross, origBID, origTID, "original title", "original text", srcBID)

	// Create a third board to try reposting the repost
	dstBID = boards2.CreateBoard(cross, "destination-board", false)
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	boards2.CreateRepost(cross, srcBID, srcTID, "repost title", "repost text", dstBID)
}

// Error:
// reposting a thread that is a repost is not allowed
