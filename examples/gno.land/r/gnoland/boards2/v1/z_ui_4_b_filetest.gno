// Render thread sorting comments from newest to oldest.
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

var threadID boards.ID

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create a board and a thread
	boardID := boards2.CreateBoard(cross, boardName, false, false)
	threadID = boards2.CreateThread(cross, boardID, "Foo", "Body")

	// Create two comments and a reply to a comment
	boards2.CreateReply(cross, boardID, threadID, 0, "First comment")

	replyID := boards2.CreateReply(cross, boardID, threadID, 0, "Second comment")
	boards2.CreateReply(cross, boardID, threadID, replyID, "Third comment")
}

func main() {
	path := boardName + "/" + threadID.String() + "?order=desc"
	println(boards2.Render(path))
}

// Output:
// # Foo
//
// Body
//
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC
// [Flag](/r/gnoland/boards2/v1$help&func=FlagThread&boardID=1&reason=&threadID=1) • [Repost](/r/gnoland/boards2/v1$help&func=CreateRepost&boardID=1&body=&destinationBoardID=&threadID=1&title=) • [Reply](/r/gnoland/boards2/v1$help&func=CreateReply&boardID=1&body=&replyID=0&threadID=1) • [Edit](/r/gnoland/boards2/v1$help&func=EditThread&boardID=1&body=Body&threadID=1&title=Foo) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteThread&boardID=1&threadID=1)
//
// ---
// Sort by: [newest first](/r/gnoland/boards2/v1:BoardName/1?order=asc)
//
// >
// > Second comment
// >
// > Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC, [#3](/r/gnoland/boards2/v1:BoardName/1/3)
// [Flag](/r/gnoland/boards2/v1$help&func=FlagReply&boardID=1&reason=&replyID=3&threadID=1) • [Reply](/r/gnoland/boards2/v1$help&func=CreateReply&boardID=1&body=&replyID=3&threadID=1) • [Edit](/r/gnoland/boards2/v1$help&func=EditReply&boardID=1&body=Second+comment&replyID=3&threadID=1) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteReply&boardID=1&replyID=3&threadID=1)
// >
// > >
// > > Third comment
// > >
// > > Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC, [#4](/r/gnoland/boards2/v1:BoardName/1/4)
// [Flag](/r/gnoland/boards2/v1$help&func=FlagReply&boardID=1&reason=&replyID=4&threadID=1) • [Reply](/r/gnoland/boards2/v1$help&func=CreateReply&boardID=1&body=&replyID=4&threadID=1) • [Edit](/r/gnoland/boards2/v1$help&func=EditReply&boardID=1&body=Third+comment&replyID=4&threadID=1) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteReply&boardID=1&replyID=4&threadID=1)
//
// >
// > First comment
// >
// > Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC, [#2](/r/gnoland/boards2/v1:BoardName/1/2)
// [Flag](/r/gnoland/boards2/v1$help&func=FlagReply&boardID=1&reason=&replyID=2&threadID=1) • [Reply](/r/gnoland/boards2/v1$help&func=CreateReply&boardID=1&body=&replyID=2&threadID=1) • [Edit](/r/gnoland/boards2/v1$help&func=EditReply&boardID=1&body=First+comment&replyID=2&threadID=1) • [Delete](/r/gnoland/boards2/v1$help&func=DeleteReply&boardID=1&replyID=2&threadID=1)
