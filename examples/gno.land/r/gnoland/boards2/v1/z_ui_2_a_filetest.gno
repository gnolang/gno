// Render default board view.
package main

import (
	"testing"

	boards2 "gno.land/r/gnoland/boards2/v1"
)

const (
	owner     address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq" // @devx
	boardName         = "BoardName"
)

func init() {
	testing.SetRealm(testing.NewUserRealm(owner))

	// Create a board and then add 3 threads
	boardID := boards2.CreateBoard(cross, boardName, false)

	// Create thread "A" with a single comment
	threadID := boards2.CreateThread(cross, boardID, "A", "Body")
	boards2.CreateReply(cross, boardID, threadID, 0, "Body")

	// The other 2 threads are created without comments
	boards2.CreateThread(cross, boardID, "B", "Body")
	threadID = boards2.CreateThread(cross, boardID, "C", "Body")

	// Repost thread "C" into a different board
	dstBoardID := boards2.CreateBoard(cross, "Bar", false)
	boards2.CreateRepost(cross, boardID, threadID, dstBoardID, "Title", "Body")
}

func main() {
	println(boards2.Render(boardName))
}

// Output:
// # BoardName
// Board created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC, #1
//
// [Create Thread](/r/gnoland/boards2/v1$help&func=CreateThread&boardID=1&body=&title=) • [Request Invite](/r/gnoland/boards2/v1$help&func=RequestInvite&boardID=1) • [Manage Board](?menu=manageBoard)
//
// ---
// Sort by: [oldest first](/r/gnoland/boards2/v1:BoardName?order=desc)
//
// **≡ [A](/r/gnoland/boards2/v1:BoardName/1)**
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC
// **1 replies • 0 reposts**
//
// **≡ [B](/r/gnoland/boards2/v1:BoardName/3)**
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC
// **0 replies • 0 reposts**
//
// **≡ [C](/r/gnoland/boards2/v1:BoardName/4)**
// Created by [g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq](/u/g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq) on 2009-02-13 11:31pm UTC
// **0 replies • 1 reposts**
