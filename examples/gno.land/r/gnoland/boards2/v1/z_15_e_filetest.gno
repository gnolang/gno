package main

import (
	"std"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	srcBID boards2.BoardID
	dstBID boards2.BoardID
	srcTID boards2.PostID
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))

	// Create a board with a thread
	origBID := cross(boards2.CreateBoard)("origin-board", false)
	origTID := cross(boards2.CreateThread)(origBID, "title", "text", false)

	// Create a second board and repost a thread using an empty title
	srcBID = cross(boards2.CreateBoard)("source-board", false)
	srcTID = cross(boards2.CreateRepost)(origBID, origTID, "original title", "original text", srcBID)

	// Create a third board to try reposting the repost
	dstBID = cross(boards2.CreateBoard)("destination-board", false)
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	cross(boards2.CreateRepost)(srcBID, srcTID, "repost title", "repost text", dstBID)
}

// Error:
// reposting a thread that is a repost is not allowed
