// Render thread repost of a flagged thread.
package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner        address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	srcBoardName         = "BoardName"
	dstBoardName         = "BoardName2"
)

var repostThreadID boards.ID

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create a board and a thread
	boardID := boards2.CreateBoard(cross, srcBoardName, false, false)
	threadID := boards2.CreateThread(cross, boardID, "Foo", "Body")

	// Repost thread into a different board
	dstBoardID := boards2.CreateBoard(cross, dstBoardName, false, false)
	repostThreadID = boards2.CreateRepost(cross, boardID, threadID, dstBoardID, "Bar", "Body2")

	// Flag original thread
	boards2.SetFlaggingThreshold(cross, boardID, 1)
	boards2.FlagThread(cross, boardID, threadID, "")
}

func main() {
	path := dstBoardName + "/" + repostThreadID.String()
	println(boards2.Render(path))
}

// Output:
// # Bar
//
// > *⚠ Source post has been flagged as inappropriate*
// >
// Body2
//
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC
// [Flag](/r/gnoland/boards2/v1$help&func=FlagThread&boardID=2&reason=&threadID=1) • [Repost](/r/gnoland/boards2/v1$help&func=CreateRepost&boardID=2&body=&destinationBoardID=&threadID=1&title=) • [Reply](/r/gnoland/boards2/v1$help&func=CreateReply&boardID=2&body=&replyID=0&threadID=1) • [Edit](/r/gnoland/boards2/v1$help&func=EditThread&boardID=2&body=Body2&threadID=1&title=Bar) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteThread&boardID=2&threadID=1)
