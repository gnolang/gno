// Render comment/reply view of a flagged comment.
package main

import (
	"testing"

	"gno.land/p/gnoland/boards"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	boardName         = "BoardName"
)

var (
	threadID boards.ID
	replyID  boards.ID
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create a board and a thread
	boardID := boards2.CreateBoard(cross, boardName, false)
	threadID = boards2.CreateThread(cross, boardID, "Foo", "Body")

	// Create a comments with a reply
	replyID = boards2.CreateReply(cross, boardID, threadID, 0, "Second comment")
	boards2.CreateReply(cross, boardID, threadID, replyID, "Third comment")

	// Flag the comment
	boards2.FlagReply(cross, boardID, threadID, replyID, "")
}

func main() {
	path := boardName + "/" + threadID.String() + "/" + replyID.String()
	println(boards2.Render(path))
}

// Output:
// # Foo
//
// Body
//
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC
// [Flag](/r/gnoland/boards2/v1$help&func=FlagThread&boardID=1&reason=&threadID=1) • [Repost](/r/gnoland/boards2/v1$help&func=CreateRepost&boardID=1&body=&destinationBoardID=&threadID=1&title=) • [Reply](/r/gnoland/boards2/v1$help&func=CreateReply&boardID=1&body=&replyID=0&threadID=1) • [Edit](/r/gnoland/boards2/v1$help&func=EditThread&boardID=1&body=Body&threadID=1&title=Foo) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteThread&boardID=1&threadID=1) • [Show all Replies](/r/gnoland/boards2/v1:BoardName/1)
//
//
// >
// > *⚠ Reply is hidden as it has been flagged as inappropriate*
// >
// > >
// > > Third comment
// > >
// > > Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC, [#3](/r/gnoland/boards2/v1:BoardName/1/3)
// [Flag](/r/gnoland/boards2/v1$help&func=FlagReply&boardID=1&reason=&replyID=3&threadID=1) • [Reply](/r/gnoland/boards2/v1$help&func=CreateReply&boardID=1&body=&replyID=3&threadID=1) • [Edit](/r/gnoland/boards2/v1$help&func=EditReply&boardID=1&body=Third+comment&replyID=3&threadID=1) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteReply&boardID=1&replyID=3&threadID=1)
