package main

import (
	"chain"
	"chain/runtime"
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const owner = chain.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	srcBID boards2.BoardID
	dstBID boards2.BoardID
	srcTID boards2.PostID
)

func init() {
	testing.SetRealm(runtime.NewUserRealm(owner))
	srcBID = boards2.CreateBoard(cross, "src-board", false)
	dstBID = boards2.CreateBoard(cross, "dst-board", false)

	srcTID = boards2.CreateThread(cross, srcBID, "Foo", "bar")
	boards2.FlagThread(cross, srcBID, srcTID, "idk")
}

func main() {
	testing.SetRealm(runtime.NewUserRealm(owner))

	// Repost should fail if source thread is flagged
	boards2.CreateRepost(cross, srcBID, srcTID, "foo", "bar", dstBID)
}

// Error:
// thread is hidden
