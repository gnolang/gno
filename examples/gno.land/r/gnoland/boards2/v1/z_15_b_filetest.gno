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
	srcTID boards2.PostID = 1024
)

func init() {
	testing.SetRealm(std.NewUserRealm(owner))
	srcBID = cross(boards2.CreateBoard)("src-board")
	dstBID = cross(boards2.CreateBoard)("dst-board")
}

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	// Repost should fail if source thread doesn't exist
	cross(boards2.CreateRepost)(srcBID, srcTID, "foo", "bar", dstBID)
}

// Error:
// thread does not exist with ID: 1024
