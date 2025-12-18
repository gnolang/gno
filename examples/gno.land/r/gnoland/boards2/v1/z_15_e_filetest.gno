package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx

var (
	srcBID boards.ID
	dstBID boards.ID
	srcTID boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create a board with a thread
	origBID := boards2.CreateBoard(cross, "origin-board", false)
	origTID := boards2.CreateThread(cross, origBID, "title", "text")

	// Create a second board and repost a thread using an empty title
	srcBID = boards2.CreateBoard(cross, "source-board", false)
	srcTID = boards2.CreateRepost(cross, origBID, origTID, srcBID, "original title", "original text")

	// Create a third board to try reposting the repost
	dstBID = boards2.CreateBoard(cross, "destination-board", false)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	boards2.CreateRepost(cross, srcBID, srcTID, dstBID, "repost title", "repost text")
}

// Error:
// reposting a thread that is a repost is not allowed
