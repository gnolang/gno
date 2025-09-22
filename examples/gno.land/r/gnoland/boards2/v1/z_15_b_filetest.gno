package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	srcBID boards2.BoardID
	dstBID boards2.BoardID
	srcTID boards2.PostID = 1024
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	srcBID = boards2.CreateBoard(cross, "src-board", false)
	dstBID = boards2.CreateBoard(cross, "dst-board", false)
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Repost should fail if source thread doesn't exist
	boards2.CreateRepost(cross, srcBID, srcTID, "foo", "bar", dstBID)
}

// Error:
// thread does not exist with ID: 1024
