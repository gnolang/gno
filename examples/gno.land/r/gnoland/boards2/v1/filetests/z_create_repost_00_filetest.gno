package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"

var (
	srcBID boards.ID
	dstBID boards.ID
	srcTID boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))
	srcBID = boards2.CreateBoard(cross, "src-board", false, false)
	dstBID = boards2.CreateBoard(cross, "dst-board", false, false)

	srcTID = boards2.CreateThread(cross, srcBID, "Foo", "bar")
	boards2.FlagThread(cross, srcBID, srcTID, "idk")
}

func main() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Repost should fail if source thread is flagged
	boards2.CreateRepost(cross, srcBID, srcTID, dstBID, "foo", "bar")
}

// Error:
// thread is hidden
